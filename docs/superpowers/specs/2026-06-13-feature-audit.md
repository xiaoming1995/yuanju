# 全面功能审计报告（功能/体验 + 性能 + 代码质量）

> Date: 2026-06-13
> 方法：三路并行只读扫描（设计见 `2026-06-13-feature-audit-design.md`），关键证据已由控制者逐条复核
> 用法：勾选想做的条目，按 P0 → P1 → P2 顺序执行；每条带证据、建议做法、预估工作量（S <半天 / M 半天–2天 / L >2天）

## 构建产物现状（npm run build 实测）

```
dist/assets/index-C4mmSmVN.css       136.02 kB │ gzip:  23.10 kB
dist/assets/purify.es-BEO7sGA_.js     22.44 kB │ gzip:   8.81 kB
dist/assets/index.es-zx29uYdq.js     151.41 kB │ gzip:  48.88 kB
dist/assets/index-3E2Lludt.js      1,629.73 kB │ gzip: 496.71 kB   ← 单一主 chunk
```

> 2026-06-13 P1-1/P1-2 修复后（902ddfe）：主 chunk 576.89 kB（gzip 187.89 kB），23 个页面各自成 chunk，html2canvas / jspdf 按需加载。

---

## P0：影响用户正常使用

- [x] **P0-1 历史页加载失败被静默吞掉，伪装成"空历史"**（S）✅ 已修（2f95ab7）
  接口失败时错误被丢弃，`loading` 正常结束、列表为空，用户会以为自己的命盘记录全部丢失。
  证据：`frontend/src/pages/HistoryPage.tsx:41` `.catch(() => {})`。
  建议：记录 error 状态，展示"加载失败 + 重试"面板；顺带排查 `CompatibilityHistoryPage`、`ProfilePage` 是否有同款静默 catch。

---

## P1：明显改善体验或省维护成本

### 性能

- [x] **P1-1 路由零代码分割，23 个页面打进单个 1.63MB chunk（gzip 497KB）**（M）✅ 已修（902ddfe）
  首页用户也要下载 admin 后台、合婚、PDF 导出的全部代码。
  证据：`frontend/src/App.tsx:6-31` 全部静态 import，`grep lazy|import(` 计 0；构建输出见上表。
  做法：除首页外全部页面 `React.lazy`，Suspense 占位复用 `.skeleton`。主 chunk 降至 576.89 kB（gzip 187.89 kB）。

- [x] **P1-2 html2canvas / html-to-image 打进首屏 bundle**（S，建议与 P1-1 一起做）✅ 已修（902ddfe）
  导出库只在点击"导出"时才需要。
  证据：`frontend/src/pages/ResultPage.tsx:20-22`、`CompatibilityResultPage.tsx:22-24` 顶层 import。
  做法：改为导出操作内动态 `import()`，html2canvas（199.56 kB）/ jspdf（399.62 kB）成独立按需 chunk。

- [x] **P1-3 HistoryPage 分页缺失**（M）✅ 已修（6c5fd62）
  ~~`getHistory()` 无分页参数，列表一次渲染全部，命盘多了会卡顿~~
  **审计修正**：实际问题相反——后端 GetHistory 本就 limit 20 分页（`bazi_handler.go:273-275`），前端 API 也支持 page 参数，但 HistoryPage 只取第 1 页且响应无 total，**超过 20 条的旧命盘被静默截断不可见**。
  做法：repository 新增一条聚合 SQL 返回 total/男女命分布，响应补 3 个字段；前端"加载更多"追加分页，统计卡改用服务端总数。

### 架构硬规

- [x] **P1-4 admin_handler.go 在 handler 层直接写 SQL（违反 AGENTS.md:140 硬规）**（M）✅ 已修（27597c3）
  「所有数据库操作集中在 `internal/repository/` 层，严禁在 handler 中直接写 SQL」。
  证据：`backend/internal/handler/admin_handler.go:241-250`（AdminGetStats 6 次独立 QueryRow）、`:304-345`（AdminGetUsers 查询 + count）。
  做法：AdminGetStats / AdminGetAIStats（同款违规，一并迁移）/ AdminGetUsers 迁入 `admin_repository.go`；总览统计 6 次往返合并为 1 条多子查询 SQL，并补上被忽略的错误处理。handler 层 `database.DB` 引用清零。

- [ ] **P1-5 超长文件拆分：14 个文件超 500 行上限（ENGINEERING.md:26 硬规）**（L，建议分批、改哪个拆哪个）
  实测行数：`report_service.go` 1505、`ResultPage.tsx` 1457、`PastEventsPage.tsx` 910、`lib/api.ts` 882、`PrintLayout.tsx` 735、`compatibility_service.go` 706、`TokenUsagePage.tsx` 702、`admin_handler.go` 599、`bazi_handler.go` 588 等。
  建议拆法（择要）：ResultPage 抽 useResultData / useReportState / 导出逻辑三个 hook + 子组件；api.ts 按域拆 baziApi / compatibilityApi / brandApi / adminApi；report_service.go 按大运/流年/流月拆服务 + PromptBuilder。
  注：不建议为拆而拆专项重写，优先在"下次要改这个文件时"先拆再改（CLAUDE.md §5.3 流程）。

### 测试

- [x] **P1-6 前端 src 零单测**（M 起步）✅ 已修（8a3961f）
  `find frontend/src -name '*.test.*'` 计 0。最划算的起步：给 `lib/` 纯函数补 vitest 单测——`chartLabel.ts`、`reportText.ts`、`brandText.ts`、`compatibilityPersonality.ts`、`components/pillarsInput.ts`（parseEightChars）。
  做法：引入 vitest，上述 5 个模块共 40 个用例，`npm test` 运行。
  **审计修正**：原表述"前端零测试"不准确——`frontend/tests/` 下实有 51 个 node:test 静态断言文件（断言源码内容，非逻辑单测），此前无 npm 脚本入口、无人执行，已有 3 个断言被既往重构破坏而无人发现（52dd969 与 902ddfe），本次一并修复并补 `npm run test:static` 入口。

- [x] **P1-7 pkg/crypto 加密模块零测试**（S）✅ 已修（8a3961f）
  AES 加密是安全关键路径，`backend/pkg/crypto/` 仅 crypto.go、无 `_test.go`。
  做法：往返加解密（空串/中文/长密钥截断）+ 错误密钥 + 密文篡改 + 非法输入 + Key 脱敏共 6 组用例。

### 体验

- [x] **P1-8 PDF 导出错误提示与操作脱节**（S）✅ 已修（1298b93）
  `handleExportPDF` 把错误写进 `reportError`，展示在 AI 解读区的报错面板里，面板按钮是"重新生成报告"而非重试导出；用户在导出按钮附近看不到任何反馈。
  证据：`frontend/src/pages/ResultPage.tsx:479-486`（setReportError）、`:1219-1224`（展示位置）。
  做法：导出前置提示改 toast 就近反馈，与报告生成错误分离。

---

## P2：锦上添花

- [x] **P2-1 401 硬跳转整页 reload**（S）✅ 已修（d829663）
  `window.location.href = '/login'` 丢失 SPA 状态与在途操作。
  做法：401 时派发 `yj:unauthorized` 事件，AuthContext 清状态并走路由跳转。admin 经核实有独立 axios 实例和 401 处理（`adminApi.ts`），互不影响，待确认项销掉。

- [x] **P2-2 合婚结果页生成无进度提示**（S）✅ 已修（57076c5）
  单人结果页有五步 LOADING_STEPS 进度文案，合婚报告生成只有静态 loading。
  做法：DeepReportNarrative 内置 5 步 × 4s 步进文案，节奏对齐单人结果页。

- [x] **P2-3 LiuYueDrawer 快速切换年份可能串台**（S）✅ 已修（d829663）
  切换时未中止上一个请求，响应可能后到覆盖。
  做法：请求序号守卫丢弃过期响应（axios abort 错误会被拦截器改写文案，序号守卫等效且更简单）。

- [x] **P2-4 流式中断错误文案不区分原因**（S）✅ 已修（1298b93）
  超时/网络/服务端错误统一显示"生成中断，点击重试"。
  证据：`frontend/src/pages/PastEventsPage.tsx:73-75`。
  做法：在 `api.ts` 流式 catch 处按超时/主动中止/网络失败/服务端 HTTP 错误区分中文文案，页面层透传。

- [x] **P2-5 重复代码三组**（合计 S–M）✅ 部分完成（83d48e6）
  ① ✅ `formatDate` 移入 `lib/chartLabel.ts`，HistoryPage/ProfilePage 复用；ProfilePage 的重复 `genderText` 一并收敛。
  ② ❌ 不抽 `usePageData`：5 个候选页面的鉴权守卫/错误展示/分页形状各异，强行抽象不满足单一职责标准（CLAUDE.md §5.1）。顺带修了 CompatibilityHistoryPage 加载缺 `.catch` 的静默吞错（P0-1 同款，当时排查遗漏）。
  ③ ⏳ ShareCard↔CompatibilityShareCard、PrintLayout↔CompatibilityPrintLayout 同构组件对 → 留给 P1-5 拆分时一并处理。

- [x] **P2-6 神煞注解加载失败完全无声**（S）✅ 已修（1298b93）
  可选数据失败可以容忍，但连 console.warn 都没有，线上排查困难。
  证据：`frontend/src/pages/ResultPage.tsx:381-389`。

- [x] **P2-7 报告 tab 切换不滚动**（S）✅ 已修（57076c5）
  原版/润色版切换后视口停在原处。做法：切换后 scrollIntoView 回报告区顶部，scroll-margin-top 避开 fixed 导航栏。

- [x] **P2-8 BottomNav 高亮判断粗糙**（S）✅ 已修（57076c5）
  做法：抽成 `activeTab(pathname)` 显式映射，补上 `/result`→测算、`/settings`→我的 的归属。

- [x] **P2-9 lint 残留 1 个 warning**（S，5 分钟）✅ 已修（2f95ab7）
  `frontend/src/pages/admin/AdminChartsPage.tsx:90` useEffect 缺 `fetchCharts` 依赖。按文件内既有写法显式 disable（fetchCharts 引用搜索条件，加入依赖会改变请求时机）。

- [x] **P2-10 构建缺体积监控**（S）✅ 已修（7fa82c5）
  做法：`ANALYZE=1 npm run build` 生成 `dist/stats.html`（rollup-plugin-visualizer，含 gzip），常规构建不受影响。

- [x] **P2-11 静态数据无缓存策略**（M）✅ 部分完成（d829663）
  做法：新增 `lib/staticCache.ts`（localStorage+TTL），神煞注解缓存 1 小时，两处调用受益。品牌配置不缓存——按用户隔离 + 后台可改需失效处理，复杂度大于收益。

### 待确认（已于 2026-06-13 代码级核实）

- [x] **P2-12 命盘详情网格小屏可能横向溢出** ❎ 不成立，销掉
  代码已有三重防线：① `.bazi-data-grid` 容器 `overflow-x: auto` 兜底滑动（`ResultPage.css:413`，注释自述"核心修复"）；② ≤640px 断点专门收窄（`:1571` 起，36px 标签列 + `minmax(0,1fr)` 允许列收缩到内容宽度以下）；③ 神煞标签 `flex-wrap: wrap`。最坏情况是极窄屏出现可滑动，不会破版遮挡。无需改码。
- [ ] **P2-13 品牌 logo 上传失败需重选文件**（S）✔ 确认成立，待修
  复现路径（代码级）：选文件 → 裁剪弹窗 → `handleCropConfirm` 开头即 `setCropSourceUrl(null)` 关弹窗丢弃裁剪结果（`BrandSettingsPage.tsx:136`）→ 上传失败仅 `setError`（`:146`），裁剪好的 file 已丢，用户须重选文件重新裁剪。
  修复方向：失败时把裁剪后的 file 留在 state，错误提示旁给"重试上传"按钮。
  注：input 的 `e.target.value = ''`（`:115`）没问题，同文件可重选。

### 已有专项文档的遗留

- CSS 归一化候选（近似金色合并、13px 字号档、999px 圆角、过渡时长、748 处内联样式等）见 `2026-06-12-css-style-suggestions.md`，全部未勾选。

---

## 附录：审计中排除的误报

复核时推翻的子代理发现，记录在此避免重复上当：

1. ~~合婚表缺 user_id 索引~~：`idx_compatibility_readings_user_id` 已存在（`00001_baseline.sql:778`），相关 3 条性能发现作废。
2. ~~AdminGetUsers 查询失败缺 return~~：实有 `return`（`admin_handler.go:317-320`）。
3. ~~八字反查显示候选时不清除旧错误~~：`handlePillarsSubmit` 开头即 `setError('')`（`HomePage.tsx:71`）。
4. ~~起盘提交无重复点击保护~~：提交按钮已有 `disabled={loading}`。
5. 后端 TODO/FIXME/HACK：全库 grep 为 0，无技术债注释积压。

## 整体观察

- 最弱的两环：**错误反馈链路**（静默 catch、错误显示错位、文案不区分原因——P0-1/P1-8/P2-4/P2-6 同根）和**首屏体积**（P1-1/P1-2/P2-10 同根），各自适合打包成一个专项做。
- 后端质量整体好于前端：算法层测试扎实（49 个单测），但 handler 层有分层违规、repository 层 17 个文件仅 2 个有测试。
- 推荐起手顺序：P0-1 + P2-9（半天内）→ P1-1+P1-2（一次分包专项）→ P1-4（分层修复）→ P1-6/P1-7（测试起步）。
