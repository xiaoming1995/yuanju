# 导出图片/PDF 品牌定制化 Spec

**日期：** 2026-05-18
**作者：** Claude + 用户
**状态：** 已批准，待实施

---

## 1. 背景

当前两条导出链路均把 "缘聚命理" 品牌硬编码在产物中：

- **分享图片**（`frontend/src/components/ShareCard.tsx`）：顶部 brown gradient header 居中 "缘聚 命 理"，底部右侧 "yuanju.com"
- **PDF**（`frontend/src/components/PrintLayout.tsx`）：独立的 header / footer 同样硬编码 "缘聚" 品牌

用户希望开放品牌定制 —— 让登录用户自己上传 logo、设置标题、配置水印，导出的 PNG / PDF 反映用户自己的品牌（白标）而不是平台品牌。

## 2. 目标

- 让所有登录用户能在设置页配置：品牌 logo（图片）、品牌标题（文字）、底部品牌行（文字）、水印（无 / 底部文字 / 满页对角）。
- 用户填了字段后，导出产物上对应位置的 "缘聚" 默认值被完全替换（白标模式）。
- 同一套品牌设置同时应用于 PNG 图片导出和 PDF 导出。
- 后端新增文件上传 + 静态托管 + 速率限制基建。

## 3. 非目标

- **不动 ResultPage 网页自身渲染** —— 网页上不出现用户品牌（避免主页面 UI 被用户品牌污染），定制仅作用于导出产物。
- **不做多模板** —— 单用户单套设置（YAGNI）。后续需要时再加。
- **不做付费/角色门槛** —— MVP 对所有登录用户开放。后续要门槛时在 handler 入口加一行即可。
- **不动 AI 报告正文** —— AI 自由叙述里若提到 "缘聚" 不做替换。
- **不实现孤儿文件清扫** —— 用户注销后 logo 文件不自动删，MVP 接受，后续加 cron。

## 4. 总体架构

### 4.1 数据模型

新增单表 `user_export_brand`，每用户至多一行：

```sql
CREATE TABLE IF NOT EXISTS user_export_brand (
  user_id        VARCHAR(36)  PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
  title          VARCHAR(20)  NOT NULL DEFAULT '',
  footer_text    VARCHAR(40)  NOT NULL DEFAULT '',
  logo_path      VARCHAR(200) NOT NULL DEFAULT '',   -- 服务器相对路径 brand-logos/<uuid>.<ext>
  watermark_mode VARCHAR(16)  NOT NULL DEFAULT 'none', -- 'none' | 'bottom' | 'diagonal'
  watermark_text VARCHAR(30)  NOT NULL DEFAULT '',
  updated_at     TIMESTAMP    NOT NULL DEFAULT NOW()
);
```

用户首次访问时表里无记录 → API 返回全空对象 → 前端用 "缘聚" 默认值。

### 4.2 文件上传链路

```
[浏览器] ─multipart─▶ [POST /api/user/export-brand/logo]
                       ├─ 校验：Content-Length ≤ 2MB
                       ├─ 校验：MIME ∈ {png,jpeg,webp}
                       ├─ 校验：magic bytes 与 MIME 一致
                       ├─ 落盘：{UploadDir}/brand-logos/<uuid>.<ext>
                       ├─ UPDATE user_export_brand SET logo_path=...
                       └─ best-effort 删除旧 logo 文件
                              │
                              ▼
[浏览器拿到 logo_url] ─GET─▶ [/static/uploads/brand-logos/<uuid>.<ext>]
                              （Gin Static 静态托管，无 auth）
```

`UploadDir` 由 `configs.Config.UploadDir` 控制，默认 `./uploads`，Docker 环境可挂载 volume。

### 4.3 导出时数据流

```
ResultPage 加载（用户已登录）
   │
   ├─ GET /api/user/export-brand → brand 对象（或 null）
   │
   ▼
用户点 "保存图片" / "导出 PDF"
   │
   ▼
<ShareCard brand={brand} ... />
<PrintLayout brand={brand} ... />
   │
   ▼ 组件内部按字段优先级渲染：用户值 || 缘聚默认值
```

## 5. 改动清单

### 5.1 后端（Go）

| 文件 | 操作 | 备注 |
|------|------|------|
| `backend/pkg/database/database.go` | + DDL `user_export_brand` | 加在用户表后面，索引仅 PK |
| `backend/internal/model/user_brand.go` | **新建** | `ExportBrand` struct |
| `backend/internal/repository/user_brand_repository.go` | **新建** | `Get(userID)` / `Upsert(userID, ExportBrand)` / `UpdateLogoPath(userID, path)` / `ClearLogoPath(userID)` |
| `backend/internal/handler/user_brand_handler.go` | **新建** | 4 个 handler + 1 个 rate-limit middleware 应用 |
| `backend/pkg/ratelimit/inmem.go` | **新建** | 进程内 token-bucket rate limiter |
| `backend/cmd/api/main.go` | 修改 | 注册 4 路由 + `r.Static("/static/uploads", cfg.UploadDir)` + `os.MkdirAll(uploadDir+"/brand-logos", 0755)` |
| `backend/configs/config.go` | 修改 | 加 `UploadDir string` 字段，env `UPLOAD_DIR` 默认 `./uploads` |
| `backend/internal/handler/user_brand_handler_test.go` | **新建** | 11 个 handler 测试（详见 §10） |

### 5.2 前端（React + TS）

| 文件 | 操作 | 备注 |
|------|------|------|
| `frontend/src/lib/api.ts` | 修改 | 加 `ExportBrand` 类型 + `brandAPI` 模块 |
| `frontend/src/pages/BrandSettingsPage.tsx` | **新建** | 设置页主体 |
| `frontend/src/pages/BrandSettingsPage.css` | **新建** | 样式 |
| `frontend/src/components/BrandPreviewCard.tsx` | **新建** | 设置页内嵌的轻量预览组件（约 80 行） |
| `frontend/src/components/ShareCard.tsx` | 修改 | 加 `brand?` prop，按字段覆盖 |
| `frontend/src/components/PrintLayout.tsx` | 修改 | 加 `brand?` prop，按字段覆盖 |
| `frontend/src/pages/ResultPage.tsx` | 修改 | 加载 brand → 传给两个 export 组件 |
| `frontend/src/pages/ProfilePage.tsx` | 修改 | 加 "导出品牌设置" 入口卡片 |
| `frontend/src/App.tsx` | 修改 | 加 `/settings/brand` 路由 |
| `frontend/vite.config.ts` | 修改 | dev server proxy 加 `/static` → `localhost:9002` |
| `frontend/tests/brand-settings.test.mjs` | **新建** | 4 个静态正则测试 |

合计约 16 个文件改动（含新建）。

## 6. API 契约

所有 endpoint 在 `JWTAuth()` 中间件之后。返回结构：成功 `{"data": ...}`，错误 `{"error": "..."}`。

### 6.1 `GET /api/user/export-brand`

读取当前用户的品牌设置。无记录返回空对象，**不**报错。

**Response 200**
```json
{
  "data": {
    "title": "清雨堂",
    "footer_text": "清雨堂 · 命理咨询",
    "logo_url": "/static/uploads/brand-logos/a3f4b2c1-...png",
    "watermark_mode": "diagonal",
    "watermark_text": "清雨堂 · 仅供咨询参考"
  }
}
```

`logo_url` 由后端拼接：`logo_path` 非空 → `/static/uploads/<logo_path>`，空 → `""`。前端拿到直接用。

### 6.2 `PUT /api/user/export-brand`

更新文字类设置（不含 logo）。upsert 语义。

**Request body**
```json
{
  "title": "清雨堂",
  "footer_text": "清雨堂 · 命理咨询",
  "watermark_mode": "diagonal",
  "watermark_text": "清雨堂 · 仅供咨询参考"
}
```

**字段约束**

| 字段 | 类型 | 长度（rune） | 校验 |
|---|---|---|---|
| `title` | string | ≤ 20 | 去首尾空白；禁止 `<>"'&` 字符 |
| `footer_text` | string | ≤ 40 | 同上 |
| `watermark_mode` | string | — | 枚举：`none` / `bottom` / `diagonal` |
| `watermark_text` | string | ≤ 30 | 同 title 字符校验；`mode != 'none'` 时建议非空但不强制 |

校验用 `utf8.RuneCountInString` 而非 `len()`。空字符串 = 清除该项（回退默认值）。

**Response 200**：返回与 GET 相同的完整对象。
**Response 400**：字段超长 / mode 非法 / 含非法字符。

### 6.3 `POST /api/user/export-brand/logo` (multipart)

上传 logo。`form-data` 字段名 `file`。

**三层校验**
1. `c.Request.ContentLength > 2*1024*1024` → 413（不消耗 IO 早拦截）
2. multipart part 的 `Content-Type` ∈ `{image/png, image/jpeg, image/webp}` → 否则 400
3. 文件前 12 字节 magic bytes 匹配声明的 MIME → 否则 400

**Magic bytes**
- PNG: `89 50 4E 47 0D 0A 1A 0A`
- JPEG: `FF D8 FF`
- WebP: `52 49 46 46 ?? ?? ?? ?? 57 45 42 50`（字节 0-3 + 8-11）

**落盘**
- 路径：`{UploadDir}/brand-logos/<uuid-v4>.<ext>`
- `<ext>` 从 magic bytes 推导（**不**信任客户端文件名）：`.png` / `.jpg` / `.webp`

**事务序**
1. 写入新文件到磁盘
2. UPDATE DB `logo_path`
3. 删除旧文件（best-effort，失败仅 log warn）

若步骤 2 失败：删掉刚写的新文件，返回 500（不留孤儿）。

**Response 200**
```json
{ "data": { "logo_url": "/static/uploads/brand-logos/a3f4b2c1-...png" } }
```

**速率限制**
- 进程内 token bucket：每 `user_id` 10 次/分钟
- 超限 → 429 `{"error": "上传过于频繁，请稍后重试"}`
- 实现：`pkg/ratelimit/inmem.go` 的 `Limiter.Allow(userID)`
- 仅作用于此 endpoint，其它三个不限

### 6.4 `DELETE /api/user/export-brand/logo`

清除 logo。删磁盘文件（best-effort）+ DB `logo_path` 置空。Response 204。

### 6.5 静态托管

`cmd/api/main.go` 启动时：

```go
if err := os.MkdirAll(filepath.Join(cfg.UploadDir, "brand-logos"), 0755); err != nil {
    log.Fatalf("create upload dir: %v", err)
}
r.Static("/static/uploads", cfg.UploadDir)
```

- 该路由 **不** 走 `JWTAuth()` —— 浏览器导出时 `<img>` 标签直接拉，必须公开
- Gin `Static` 内部用 `http.FileServer`，自动拒绝 `..` 路径穿越
- 路径含 uuid，足够防穷举

## 7. 前端设置页 UX

### 7.1 路由 + 入口

- 路由 `/settings/brand`（在 `App.tsx`），未登录 redirect `/login`
- 入口：`ProfilePage` 新增一个 "导出品牌设置" 卡片，点入

### 7.2 页面布局

```
┌──────────────────────────────────────────────────────────┐
│  ← 返回                       导出品牌设置                │
├──────────────────────────────────────────────────────────┤
│                                                          │
│  〔 顶部品牌 〕                                          │
│  ┌──────────────────────────────────────────────────┐   │
│  │  品牌标题  [ 清雨堂              ] N/20         │   │
│  │  覆盖默认 "缘聚 命 理"，留空则保留默认             │   │
│  │                                                  │   │
│  │  品牌 Logo                                       │   │
│  │  ┌────────┐                                      │   │
│  │  │ [当前  │  [更换]  [删除]                      │   │
│  │  │  logo] │                                      │   │
│  │  └────────┘   PNG / JPG / WebP，≤ 2MB           │   │
│  │  （未上传时显示带 + 的虚线拖拽框）                │   │
│  └──────────────────────────────────────────────────┘   │
│                                                          │
│  〔 底部品牌 〕                                          │
│  ┌──────────────────────────────────────────────────┐   │
│  │  底部文字  [ 清雨堂 · 命理咨询      ] N/40       │   │
│  │  覆盖默认 "yuanju.com"，留空保留默认              │   │
│  └──────────────────────────────────────────────────┘   │
│                                                          │
│  〔 水印 〕                                              │
│  ┌──────────────────────────────────────────────────┐   │
│  │  ○ 无水印                                        │   │
│  │  ● 底部文字水印（小字一行）                       │   │
│  │  ○ 满页对角水印（防盗图）                        │   │
│  │                                                  │   │
│  │  水印文字 [ 清雨堂 · 仅供咨询参考  ] N/30        │   │
│  │  （选 "无水印" 时此输入框禁用）                   │   │
│  └──────────────────────────────────────────────────┘   │
│                                                          │
│  〔 预览 〕                                              │
│  ┌──────────────────────────────────────────────────┐   │
│  │ [BrandPreviewCard：缩略版，含 logo + title + 占  ]│   │
│  │ [位灰条 + footer + 必要时叠 diagonal 水印]       │   │
│  └──────────────────────────────────────────────────┘   │
│                                                          │
│  [ 重置默认 ]                          [ 保存 ]          │
└──────────────────────────────────────────────────────────┘
```

### 7.3 交互细节

| 行为 | 触发 | API |
|---|---|---|
| 进页面 | mount | `GET /api/user/export-brand` |
| 选 logo 文件 | `<input type=file>` change | 立即 `POST /api/user/export-brand/logo` |
| 删 logo | "删除" 按钮 + 二次确认 | `DELETE /api/user/export-brand/logo` |
| 改文字 / 单选 | 输入 | 仅更新本地 state |
| 点 "保存" | 显式点击 | `PUT /api/user/export-brand` |
| 点 "重置默认" | 显式点击 + 二次确认 | `PUT` 全空对象 + `DELETE` logo |

**未保存提示**：本地 state 与服务器最近一次值不同 → 顶部出黄条 "有未保存的修改"，"保存" 按钮高亮。

**logo 即时上传 vs 文字延后保存** 的分裂式交互（已确认）：
- 上传是文件单次 IO，立即看效果体验自然
- 文字字段多，攒一次 PUT 减少抖动

### 7.4 BrandPreviewCard

新建 ~80 行的轻量预览组件，**不**复用 `ShareCard`（后者依赖 `structured` 报告数据太重）。
- 用占位灰条模拟四柱区
- 唯一职责：实时把当前 brand state 渲染出来
- 接收 props：`{ title, footer_text, logo_url, watermark_mode, watermark_text }`

## 8. 导出时渲染（ShareCard + PrintLayout 改造）

### 8.1 brand prop 类型

```ts
// frontend/src/lib/api.ts
export interface ExportBrand {
  title: string
  footer_text: string
  logo_url: string
  watermark_mode: 'none' | 'bottom' | 'diagonal'
  watermark_text: string
}
```

两个组件接受 `brand?: ExportBrand | null`。`null`/`undefined` 走默认。

### 8.2 ShareCard 改造点（约 40 行）

**顶部品牌栏**
- logo 渲染：`brand?.logo_url` 非空时，绝对定位左侧 24px、宽高 40×40、`object-fit: contain`、`crossOrigin="anonymous"`
- 标题文字：`brand?.title || '缘聚 命 理'`，**位置始终居中**，不为 logo 让位
- 副标题（生辰行）不变

**底部品牌行**
- 左侧 "仅供参考，不作决策依据" **始终保留**（合规免责）
- 右侧文字按以下优先级：

```ts
function footerRightText(brand: ExportBrand | null): string {
  if (!brand) return 'yuanju.com'
  if (brand.watermark_mode === 'bottom' && brand.watermark_text && brand.footer_text) {
    return `${brand.footer_text} · ${brand.watermark_text}`   // 撞位 → 拼接
  }
  if (brand.watermark_mode === 'bottom' && brand.watermark_text) {
    return brand.watermark_text
  }
  return brand.footer_text || 'yuanju.com'
}
```

**满页斜文字水印**（仅当 `brand?.watermark_mode === 'diagonal' && brand.watermark_text`）

根 `<div>` 加 `position: 'relative'`，新增覆盖层：

```tsx
{showDiagonal && (
  <div style={{
    position: 'absolute', inset: 0, pointerEvents: 'none',
    overflow: 'hidden', zIndex: 1,
  }}>
    <div style={{
      position: 'absolute',
      top: '-30%', left: '-30%', right: '-30%', bottom: '-30%',
      transform: 'rotate(-30deg)',
      display: 'grid',
      gridTemplateColumns: 'repeat(auto-fill, 180px)',
      gap: '60px 40px',
      opacity: 0.06,
      color: '#000',
      fontSize: 14,
      fontFamily: '"Noto Sans SC", sans-serif',
      whiteSpace: 'nowrap',
    }}>
      {Array.from({ length: 60 }).map((_, i) => (
        <span key={i}>{brand.watermark_text}</span>
      ))}
    </div>
  </div>
)}
```

- `top/left/right/bottom: -30%` 让旋转后四角不留白
- `opacity 0.06` 是经验值
- 内容默认 z-index 0，水印 1 → 盖在内容上

### 8.3 PrintLayout 改造点

完全对称：
1. Header 加 logo + title 覆盖逻辑
2. Footer 按 8.2 同样的优先级算 footer 文字
3. 根 div 加 `position: relative` + 满页斜水印层（PDF 多页时 grid auto-fill 自动覆盖整长度）
4. 加 `brand?` prop

### 8.4 优先级表

| 字段 | 用户填了 | 用户没填 |
|---|---|---|
| `title` | 顶部用户文字 | "缘聚 命 理" |
| `logo_url` | 顶部左侧 40×40 logo | 不渲染 logo |
| 底部右侧文字 | 见 8.2 优先级函数 | "yuanju.com" |
| 满页斜水印 | `mode === 'diagonal'` 时叠 0.06 alpha 文字层 | 不叠 |

### 8.5 ResultPage 集成

```ts
const [brand, setBrand] = useState<ExportBrand | null>(null)
useEffect(() => {
  if (!user) return
  brandAPI.get().then(r => setBrand(r.data.data)).catch(() => setBrand(null))
}, [user])

<ShareCard ... brand={brand} />
<PrintLayout ... brand={brand} />
```

未登录用户 `brand` 始终 `null` → 默认 "缘聚"。

### 8.6 Cross-origin 注意

`logo_url` = `/static/uploads/brand-logos/<uuid>.<ext>`：

- **生产**：前后端同域，html2canvas 直接渲染没问题
- **开发**：`vite.config.ts` 需要在 server.proxy 加 `'/static'` → `localhost:9002`，否则 dev 模式 html2canvas 拉不到 logo

```ts
// vite.config.ts
server: {
  proxy: {
    '/api':    { target: 'http://localhost:9002', changeOrigin: true },
    '/static': { target: 'http://localhost:9002', changeOrigin: true },  // 新增
  }
}
```

logo `<img>` 加 `crossOrigin="anonymous"` 保险。

## 9. 安全 / 校验细节

### 9.1 文件上传三层防御

| 层 | 检查 | 失败响应 |
|---|---|---|
| 1 | `c.Request.ContentLength > 2MB` | 413 |
| 2 | multipart part `Content-Type` ∈ 白名单 | 400 |
| 3 | 文件前 12 字节 magic bytes 与 MIME 一致 | 400 |

第 3 层不可省 —— 仅靠 Content-Type 可被伪造，配合公开静态托管就成了任意文件分发器。

### 9.2 文件名

`<uuid-v4>.<ext>`，`<ext>` 从 magic bytes 推导。彻底切断路径穿越 / 扩展名混淆。

### 9.3 失败路径

| 场景 | 行为 |
|---|---|
| 落盘 OK，DB UPDATE 失败 | 删刚写的新文件，500 |
| DB UPDATE OK，删旧文件失败 | log warn，仍然 200 |
| 用户注销 | DDL `ON DELETE CASCADE` 自动删 DB 行；磁盘文件保留（接受，后续 cron 清扫） |
| 上传中断网 | 400，文件未写完不落盘 |

### 9.4 文字字段服务端校验

```go
func validateBrand(req BrandUpdateReq) error {
    if utf8.RuneCountInString(req.Title) > 20 { return errTitleTooLong }
    if utf8.RuneCountInString(req.FooterText) > 40 { return errFooterTooLong }
    if utf8.RuneCountInString(req.WatermarkText) > 30 { return errWatermarkTooLong }
    if req.WatermarkMode != "none" && req.WatermarkMode != "bottom" && req.WatermarkMode != "diagonal" {
        return errInvalidMode
    }
    if hasUnsafeChars(req.Title) || hasUnsafeChars(req.FooterText) || hasUnsafeChars(req.WatermarkText) {
        return errUnsafeChars
    }
    return nil
}
```

`hasUnsafeChars` 拒绝 `<` `>` `"` `'` `&`。React 渲染会转义，但留一道。

### 9.5 速率限制实现

```go
// pkg/ratelimit/inmem.go
type Limiter struct {
    mu      sync.Mutex
    buckets map[string]*bucket
    rate    int           // 10
    window  time.Duration // 1 min
}

type bucket struct {
    tokens int
    last   time.Time
}

func New(rate int, window time.Duration) *Limiter { ... }
func (l *Limiter) Allow(key string) bool { ... }  // 滑动窗口 + token refill
```

仅作用于 `POST /api/user/export-brand/logo`。

## 10. 测试

### 10.1 后端 `backend/internal/handler/user_brand_handler_test.go`

| 测试 | 内容 |
|---|---|
| `TestBrandGet_Empty` | 新用户 GET 返回空字符串/默认值，不报错 |
| `TestBrandPut_Valid` | PUT 后再 GET 数据一致 |
| `TestBrandPut_TitleTooLong` | 21 rune 标题 → 400 |
| `TestBrandPut_InvalidMode` | mode='evil' → 400 |
| `TestBrandPut_UnsafeChars` | title 含 `<` → 400 |
| `TestBrandLogo_Upload_PNG` | 上传合法 PNG → 200 + 文件存在 + logo_path 写入 |
| `TestBrandLogo_Upload_TooLarge` | 3MB 文件 → 413，文件不落盘 |
| `TestBrandLogo_Upload_FakeMime` | Content-Type=image/png 但内容是文本 → 400 |
| `TestBrandLogo_RateLimit` | 11 次连续上传 → 第 11 次 429 |
| `TestBrandLogo_OverwriteDeletesOld` | 第二次上传后旧文件不存在 |
| `TestBrandLogo_Delete` | DELETE 后文件不存在 + DB logo_path 空 |

### 10.2 前端 `frontend/tests/brand-settings.test.mjs`

静态正则风格（与项目现有 `node --test` 约定一致）：

| 测试 | 断言 |
|---|---|
| `ShareCard accepts brand prop` | `ShareCard.tsx` 源码含 `brand?: ExportBrand` 与 `brand?.title \|\|` |
| `ShareCard preserves default fallback` | `ShareCard.tsx` 同时含 `'缘聚 命 理'` 字面量与 `brand?.title \|\|` |
| `PrintLayout accepts brand prop` | 同上对 `PrintLayout.tsx` 检查 |
| `Vite dev proxy includes /static` | `vite.config.ts` 含 `'/static'` 条目 |

不做浏览器端到端测试（与项目现状一致）。

## 11. 验收

1. **设置页生效**：登录用户访问 `/settings/brand`，填 title=`"清雨堂"`、上传 logo、选 `diagonal` 水印 + 文字 → 保存 → 回到 ResultPage 点 "保存图片"：
   - PNG 顶部左侧出现 logo，标题居中显示 `清雨堂`
   - PNG 中部叠半透明斜文字 `清雨堂 ...`
   - PNG 底部右侧不再有 `yuanju.com`
2. **PDF 同上**：导出 PDF 后所有页面均有满页水印，header/footer 用用户品牌。
3. **白标完整性**：用户填满后导出产物中**搜不到** "缘聚" 三字。
4. **未登录 / 未配置兜底**：导出产物完全是当前 "缘聚" 默认样式。
5. **撞位规则**：mode='bottom' + footer_text + watermark_text 都填 → 底部显示 `footer_text · watermark_text`。
6. **速率限制**：脚本连续打 11 次 `POST /logo` → 第 11 次 429。
7. **后端 11 个 test 全绿**；**前端 4 个静态 test 全绿**；`npm run build` + `npm run lint` 通过。

## 12. 风险与回退

- **风险**：新增静态托管 + 文件上传基建是新增攻击面，三层防御 + uuid 文件名是关键。
- **风险**：进程内 rate limit 在多实例部署下失效（每实例独立计数）。MVP 单实例可接受。
- **回退**：单 PR revert；DDL `DROP TABLE user_export_brand`；磁盘上 `uploads/` 目录可保留或删除。
- **可能漏点**：满页 diagonal 水印在不同浏览器 / html2canvas 版本下可能渲染差异，需手测 Chrome + Safari。

## 13. 实施分支

从 main 起 `feat/export-brand-customization`，单 PR 完成全部 16 个文件改动（含 DDL + 15 个新测试，11 后端 + 4 前端）。
