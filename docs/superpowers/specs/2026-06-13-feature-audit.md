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

---

## P0：影响用户正常使用

- [ ] **P0-1 历史页加载失败被静默吞掉，伪装成"空历史"**（S）
  接口失败时错误被丢弃，`loading` 正常结束、列表为空，用户会以为自己的命盘记录全部丢失。
  证据：`frontend/src/pages/HistoryPage.tsx:41` `.catch(() => {})`。
  建议：记录 error 状态，展示"加载失败 + 重试"面板；顺带排查 `CompatibilityHistoryPage`、`ProfilePage` 是否有同款静默 catch。

---

## P1：明显改善体验或省维护成本

### 性能

- [ ] **P1-1 路由零代码分割，23 个页面打进单个 1.63MB chunk（gzip 497KB）**（M）
  首页用户也要下载 admin 后台、合婚、PDF 导出的全部代码。
  证据：`frontend/src/App.tsx:6-31` 全部静态 import，`grep lazy|import(` 计 0；构建输出见上表。
  建议：`React.lazy` + `Suspense` 按路由分包，至少切出 admin、合婚、结果页三块。

- [ ] **P1-2 html2canvas / html-to-image 打进首屏 bundle**（S，建议与 P1-1 一起做）
  导出库只在点击"导出"时才需要。
  证据：`frontend/src/pages/ResultPage.tsx:20-22`、`CompatibilityResultPage.tsx:22-24` 顶层 import。
  建议：改为 `await import('html-to-image')` 等动态导入。

- [ ] **P1-3 HistoryPage 全量渲染无分页**（M）
  `getHistory()` 无分页参数，列表 `.map()` 一次渲染全部；命盘数到几百条时会卡顿。
  证据：`frontend/src/pages/HistoryPage.tsx:39`（无分页参数）、`:174-291`（全量 map）。admin 用户列表有分页，C 端历史页没有。
  建议：后端 getHistory 加 limit/offset，前端"加载更多"或分页。

### 架构硬规

- [ ] **P1-4 admin_handler.go 在 handler 层直接写 SQL（违反 AGENTS.md:140 硬规）**（M）
  「所有数据库操作集中在 `internal/repository/` 层，严禁在 handler 中直接写 SQL」。
  证据：`backend/internal/handler/admin_handler.go:241-250`（AdminGetStats 6 次独立 QueryRow）、`:304-345`（AdminGetUsers 查询 + count）。
  建议：迁移到 `internal/repository/admin_repository.go`；顺带把 AdminGetStats 的 6 次数据库往返合并为 1 条多子查询 SQL。

- [ ] **P1-5 超长文件拆分：14 个文件超 500 行上限（ENGINEERING.md:26 硬规）**（L，建议分批、改哪个拆哪个）
  实测行数：`report_service.go` 1505、`ResultPage.tsx` 1457、`PastEventsPage.tsx` 910、`lib/api.ts` 882、`PrintLayout.tsx` 735、`compatibility_service.go` 706、`TokenUsagePage.tsx` 702、`admin_handler.go` 599、`bazi_handler.go` 588 等。
  建议拆法（择要）：ResultPage 抽 useResultData / useReportState / 导出逻辑三个 hook + 子组件；api.ts 按域拆 baziApi / compatibilityApi / brandApi / adminApi；report_service.go 按大运/流年/流月拆服务 + PromptBuilder。
  注：不建议为拆而拆专项重写，优先在"下次要改这个文件时"先拆再改（CLAUDE.md §5.3 流程）。

### 测试

- [ ] **P1-6 前端零测试**（M 起步）
  `find frontend/src -name '*.test.*'` 计 0。最划算的起步：给 `lib/` 纯函数补 vitest 单测——`chartLabel.ts`、`reportText.ts`、`brandText.ts`、`compatibilityPersonality.ts`、`components/pillarsInput.ts`（parseEightChars）。

- [ ] **P1-7 pkg/crypto 加密模块零测试**（S）
  AES 加密是安全关键路径，`backend/pkg/crypto/` 仅 crypto.go、无 `_test.go`。
  建议：加解密往返 + 密文篡改 + 错误密钥三个用例。

### 体验

- [ ] **P1-8 PDF 导出错误提示与操作脱节**（S）
  `handleExportPDF` 把错误写进 `reportError`，展示在 AI 解读区的报错面板里，面板按钮是"重新生成报告"而非重试导出；用户在导出按钮附近看不到任何反馈。
  证据：`frontend/src/pages/ResultPage.tsx:479-486`（setReportError）、`:1219-1224`（展示位置）。
  建议：导出错误改 toast 或显示在导出按钮旁，与报告生成错误分离。

---

## P2：锦上添花

- [ ] **P2-1 401 硬跳转整页 reload**（S）
  `window.location.href = '/login'` 丢失 SPA 状态与在途操作；若 admin 接口共用此实例，admin 过期也会被甩到 C 端登录页（这点待确认）。
  证据：`frontend/src/lib/api.ts:33-36`。
  建议：派发自定义事件由 AuthContext 处理导航；区分 admin 路径。

- [ ] **P2-2 合婚结果页生成无进度提示**（S）
  单人结果页有五步 LOADING_STEPS 进度文案，合婚报告生成只有静态 loading。
  证据：`frontend/src/pages/CompatibilityResultPage.tsx:150` 起。
  建议：复用 LOADING_STEPS 模式。

- [ ] **P2-3 LiuYueDrawer 快速切换年份可能流式串台**（S）
  切换时未中止上一个流式请求，响应可能交错写入。
  证据：`frontend/src/components/LiuYueDrawer.tsx:80-84`；全文件 grep 无 AbortController。
  建议：每次请求挂 AbortController，切换时 abort。

- [ ] **P2-4 流式中断错误文案不区分原因**（S）
  超时/网络/服务端错误统一显示"生成中断，点击重试"。
  证据：`frontend/src/pages/PastEventsPage.tsx:73-75`。

- [ ] **P2-5 重复代码三组**（合计 S–M）
  ① `formatDate` 在 `HistoryPage.tsx:13-16` 与 `ProfilePage.tsx:9-12` 重复 → 移入 `lib/chartLabel.ts`；
  ② fetch+loading+error 手写模式 ≥7 处（HistoryPage/ProfilePage/BrandSettingsPage/CompatibilityResultPage/CompatibilityHistoryPage 等）→ 抽 `usePageData` hook（做 P0-1 时顺手）；
  ③ ShareCard↔CompatibilityShareCard、PrintLayout↔CompatibilityPrintLayout 两对同构组件（重复的字体/色值常量与 section 辅助函数）→ 抽公共基础组件（可与 P1-5 拆分合并做）。

- [ ] **P2-6 神煞注解加载失败完全无声**（S）
  可选数据失败可以容忍，但连 console.warn 都没有，线上排查困难。
  证据：`frontend/src/pages/ResultPage.tsx:381-389`。

- [ ] **P2-7 报告 tab 切换不滚动**（S）
  原版/润色版切换后视口停在原处。证据：`frontend/src/pages/ResultPage.tsx:1061-1074`。

- [ ] **P2-8 BottomNav 高亮判断粗糙**（S）
  `startsWith('/history')` 类判断在 `/bazi/:id/past-events` 等路径下高亮归属不直观。证据：`frontend/src/components/BottomNav.tsx:31`。

- [ ] **P2-9 lint 残留 1 个 warning**（S，5 分钟）
  `frontend/src/pages/admin/AdminChartsPage.tsx:90` useEffect 缺 `fetchCharts` 依赖。

- [ ] **P2-10 构建缺体积监控**（S）
  vite 仅有内置 500KB 警告。建议加 `rollup-plugin-visualizer`，防止 P1-1 治理后体积回潮。

- [ ] **P2-11 静态数据无缓存策略**（M）
  神煞注解、品牌配置等准静态数据每次挂载都重新请求。
  证据：`frontend/src/pages/ResultPage.tsx:381-397`。
  建议：localStorage + TTL 的小缓存工具，优先级不高。

### 待确认（有线索、未实证，做之前先复现）

- [ ] **P2-12 命盘详情网格小屏可能横向溢出**（M）：`ResultPage.tsx:804-895` 多列网格，需真机确认。
- [ ] **P2-13 品牌 logo 上传失败需重选文件**（S）：`BrandSettingsPage.tsx:245,254`，需复现确认。

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
