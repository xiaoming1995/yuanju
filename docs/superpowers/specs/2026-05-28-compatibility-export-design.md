# 合盘导出（图片 + PDF）设计

**Date:** 2026-05-28
**Status:** Draft
**Author:** 刘明的MacBook (with Claude Opus 4.7)

## 1. 背景与目标

合盘功能已实现完整的命理合参流程：评分引擎、决策仪表盘、证据列表、阶段风险、AI 六章节报告。但 `CompatibilityResultPage.tsx`（1089 行）目前**没有任何导出能力**——用户看完只能截屏，无法分享或保存。

单盘 `ResultPage.tsx` 早已有完整的双导出栈：
- **图片** = `ShareCard` 组件 + `html-to-image`（iOS Web Share / Android 下载 / 桌面下载）
- **PDF** = `PrintLayout` 组件 + 桌面 `window.print()` / 移动 `html2canvas + jsPDF` 多页拼接

依赖（`html-to-image`、`html2canvas`、`jspdf`）已在 `frontend/package.json`，品牌水印基建（`lib/brandText.ts` 的 `resolveFooter` / `showDiagonalWatermark`）也已抽离。

**目标**：在 `CompatibilityResultPage` 上加上与单盘对等的两按钮导出能力，独立新建组件，不动单盘代码。

## 2. 范围

### In scope

- 合盘结果详情页（`/compatibility/result/:id`）顶部新增「分享图片」「导出 PDF」两个按钮
- 新建 `CompatibilityShareCard.tsx` 分享卡组件（400px 宽国风长图，与单盘视觉对齐）
- 新建 `CompatibilityPrintLayout.tsx` A4 打印版式组件（§1–§7 全部章节）
- 分享卡通过 modal 预览（点击「分享图片」打开，modal 内有保存/分享按钮）
- PDF 复用单盘双轨：桌面 `window.print()` + 移动 `html2canvas + jsPDF`
- 两按钮在 AI polished 报告生成后才启用

### Out of scope

- 合盘档案列表页（`/compatibility/history`）的导出入口（用户明确选只在结果详情页）
- 后端 PDF 生成（不引入 wkhtmltopdf/Chrome headless）
- 单盘 `ShareCard` / `PrintLayout` 的改造（保留独立，避免回归）
- 分享卡尺寸/版面切换（固定 400px 宽 / C 详版）
- 流年报告、其他报告类型的导出
- 抽离公共子组件到 `components/print/`（按 CLAUDE.md §5.1，第 2 次出现保持原样）

## 3. 设计

### 3.1 总体架构

```
CompatibilityResultPage
  │
  ├─ 顶部按钮组
  │   ├─ [分享图片] (disabled 直到 structuredReport != null)
  │   │     → setShareModalOpen(true)
  │   └─ [导出 PDF]  (同上 disabled 条件)
  │         → handleExportPDF()
  │             ├─ 桌面: window.print()
  │             └─ 移动: html2canvas → jsPDF 多页拼接
  │
  ├─ {shareModalOpen && <Modal>
  │     <CompatibilityShareCard ref={shareCardRef} ... />
  │     <button onClick={handleSaveImage}>保存/分享</button>
  │   </Modal>}
  │
  └─ <CompatibilityPrintLayout className="print-only" ... />
        常驻挂载，display:none，仅 @media print 显示
```

### 3.2 复用既有基建

| 用途 | 复用 |
|---|---|
| 品牌 title / footer / 水印开关 | `lib/brandText.ts` (`resolveFooter`, `showDiagonalWatermark`) |
| 品牌 API | `brandAPI.get()` |
| 报告文本清洗 / 分段 | `lib/reportText.ts` (`cleanReportText`, `splitParagraphs`) |
| 决策面板数据派生 | `lib/compatibilityDecision.ts` (`buildDecisionDashboardData`, `buildDecisionStageRisks`) |
| 三端判定 | `/iPhone|iPad|iPod|Android/i.test(navigator.userAgent)` |
| 评分类型判定 | `lib/api.ts` `isV3DimensionScores` |
| 库 | `html-to-image` / `html2canvas` / `jspdf` 已装 |

### 3.3 `CompatibilityShareCard` 组件

**文件**：`frontend/src/components/CompatibilityShareCard.tsx` + `.css`

**Props**：
```ts
interface CompatibilityShareCardProps {
  reading: CompatibilityReading
  participants: CompatibilityParticipant[]
  evidences: CompatibilityEvidence[]
  structured: CompatibilityStructuredReport | null
  brand?: ExportBrand | null
}
```

**身份识别**：`participants` 数组无位置保证，必须用 `participants.find(p => p.role === 'self'/'partner')` 派生。所有"我/伴侣"的渲染都基于此。

**结构（自上而下）**：
1. Header：brand title（默认 "缘 聚 合 盘"）+ "YUANJU · 命理合参" 副标
2. 双盘四柱：`<ChartColumn participant={selfP} label="我" />` 和 `<ChartColumn participant={partnerP} label="伴侣" />`，中间竖渐变分隔线，每列含性别 label + 4 个 ganzhi 单元
3. 综合分大数字：`reading.overall_score` 36px 古铜金
4. 四维评分：根据 `isV3DimensionScores` 分支渲染（v2 = 吸引/稳定/沟通/现实，v3 = 八字/纳音/日柱/属相），每项一行条形图
5. 核心证据：按 `weight` 降序取前 3 条，左侧极性色条 + 来源 + 摘要
6. Verdict：`structured?.verdict` 或兜底文案 `综合契合度 ${score} 分，详见完整报告`
7. Footer：`resolveFooter(brand, 'yuanju.com')`
8. 可选斜线水印：`showDiagonalWatermark(brand)` 为真时叠加

**视觉规范**：
- 宽度 400px 固定，高度自适应
- 背景 `#fdf9f2`，分隔/边框 `#d4b896`，主色 `#c9a96e`，正文 `#4a3728`
- 字体：`Noto Serif SC` (天干地支) + `Noto Sans SC` (正文)
- 字体子集 URL 复用单盘的 22 字符 Google Fonts 子集

**内部小组件**（文件内部用，不导出）：
- `ChartColumn` — 单人列
- `Divider` — 横/竖渐变线
- `EvidenceItem` — 单条证据卡（极性色条）
- `DiagonalWatermark` — 斜线水印

**边界**：
- `participants.length !== 2` → 渲染兜底文字 "双盘数据缺失"
- `evidences.length < 3` → 按实际数量渲染
- `structured == null` → verdict 用兜底文案
- `dimension_scores` 缺失 → 跳过四维 section

### 3.4 `CompatibilityPrintLayout` 组件

**文件**：`frontend/src/components/CompatibilityPrintLayout.tsx` + `.css`

**Props**：
```ts
interface CompatibilityPrintLayoutProps {
  reading: CompatibilityReading
  participants: CompatibilityParticipant[]
  evidences: CompatibilityEvidence[]
  decision: DecisionDashboardData            // 由 buildDecisionDashboardData(...) 派生
  stageRisks: CompatibilityStageRisk[]       // 由 buildDecisionStageRisks(...) 派生
  structured: CompatibilityStructuredReport | null
  brand?: ExportBrand | null
}
```

> 这两个派生数据在 `CompatibilityResultPage.tsx:935-949` 已经存在（`decisionDashboard` / `decisionStageRisks`），由现有的 `consulting = normalizeConsultingAssessment(detail)` 和 `durationAssessment = normalizeDurationAssessment(...)` 产出。集成阶段**复用**这些既有变量，不重复计算。

**结构（A4 portrait，章节强制分页）**：
- §1 合参概要：双方姓名/性别/出生 + 综合分（`ParticipantsHero`）
- §2 决策仪表盘：核心结论 / 优势 / 风险（`DecisionBlock`）
- §3 评分明细：v2 四维或 v3 八字纳音日柱（`ScoreLegacyPrint` / `ScoreV3Print`）
- §4 命理证据：每条来源/极性/十神依据（`EvidenceTable`）
- §5 阶段风险与验证计划：30 天可验证项（`StageRisksBlock`）
- §6 命理解读：AI 报告六章节（`structured.chapters`，每章 `cleanReportText` + `splitParagraphs`）
- §7 双盘原图：A4 一页两列并排（`<ChartFull participant={selfP} label="我" />` 和 `<ChartFull participant={partnerP} label="伴侣" />`，按 role 区分而非数组位置）

**CSS 要点**（`@media print`）：
- `body > *:not(.print-only) { display: none !important }` 隐藏正常详情页
- `.compat-print-layout { display: block }` 显式开启
- `.print-page-table thead { display: table-header-group }` 每页重复 header
- `.compat-print-page-break { page-break-after: always }` 章节强制分页
- `@page { size: A4 portrait; margin: 18mm 16mm }`
- 双盘并列 `grid-template-columns: 1fr 1fr; gap: 12mm`
- 正文 11pt / 章节标题 16pt

### 3.5 页面集成（`CompatibilityResultPage.tsx`）

**新增 state**：
```ts
const [brand, setBrand] = useState<ExportBrand | null>(null)
const [shareModalOpen, setShareModalOpen] = useState(false)
const [savingImage, setSavingImage] = useState(false)
const [exportingPDF, setExportingPDF] = useState(false)
const shareCardRef = useRef<HTMLDivElement>(null)
const isMobileDevice = /iPhone|iPad|iPod|Android/i.test(navigator.userAgent)
const isIOS = /iPhone|iPad|iPod/i.test(navigator.userAgent)
```

**复用现有派生**（已存在于 `CompatibilityResultPage.tsx:927-949`）：
```ts
// 现有，无需新增——只是在导出 handler 和 JSX 里引用
const selfP = detail.participants.find(p => p.role === 'self')
const partnerP = detail.participants.find(p => p.role === 'partner')
const structuredReport = detail.latest_report?.content_structured
const durationAssessment = normalizeDurationAssessment(...)
const consulting = normalizeConsultingAssessment(detail)
const decisionStageRisks = buildDecisionStageRisks(consulting.stage_risks, durationAssessment)
const decisionDashboard = buildDecisionDashboardData({
  diagnosis: consulting.relationship_diagnosis,
  advice: consulting.decision_advice,
  stageRisks: consulting.stage_risks,
  duration: durationAssessment,
  evidences: detail.evidences,
  overallLevel: reading.overall_level,
})
```

**新增派生**（仅文件命名用）：
```ts
const selfName = selfP?.display_name || '我'
const partnerName = partnerP?.display_name || '伴侣'
```

**新增 useEffect**（brand 拉取）：
```ts
useEffect(() => {
  if (!user) return
  brandAPI.get().then(r => setBrand(r.data.data)).catch(() => setBrand(null))
}, [user])
```

**新增 handler**：
- `handleSaveImage`：照搬 `ResultPage.tsx:396–464` 的三端分流（iOS Web Share / Android 下载 / 桌面 toPng + a[download]）
- `handleExportPDF`：照搬 `ResultPage.tsx:469–516` 的桌面 `window.print()` + 移动 `html2canvas + jsPDF`（仅把选择器 `.print-only` 改为 `.compat-print-layout`）

**JSX 改动**：
1. 顶部新增按钮组 `.compat-export-actions`（紧贴 hero 区域右侧）
2. 文件末尾新增 modal 条件渲染
3. 文件末尾新增常驻 `<CompatibilityPrintLayout ... />`（组件内部 root 自带 `className="print-only compat-print-layout"`，外层无需传 className）

**文件命名规范**：
- 图片：`缘聚合盘-{selfName}-{partnerName}.png`
- PDF：`缘聚合盘-{selfName}-{partnerName}.pdf`

### 3.6 Modal 行为

- 点击「分享图片」→ `setShareModalOpen(true)` → modal 挂载 → ShareCard 渲染到 DOM
- 关闭路径：点遮罩 / ESC / × 按钮 / 保存成功后**保持打开**（让用户再次保存或截屏）
- Modal 打开时焦点落到关闭按钮；关闭后焦点回到「分享图片」按钮
- `max-height: 90vh; overflow-y: auto;` 应对分享卡高于视口
- `@media print { .compat-share-modal { display: none } }` 防止打印被 modal 污染

### 3.7 PDF 双轨实现

> 组件根选择器是 `.compat-print-layout`（同时也是 `.print-only` 的实例之一），handler 用此 selector 定位移动端要截图的 DOM 节点。

**桌面端**：
1. 点击按钮 → `window.print()`
2. CSS `@media print` 隐藏正常详情页，仅显示 `.print-only.compat-print-layout`
3. 浏览器原生打印对话框，用户选"另存为 PDF"

**移动端**：
1. 点击按钮 → `setExportingPDF(true)`，loading 状态
2. 临时把 `.compat-print-layout` 的 `style.display = 'block'`
3. `await document.fonts.ready`
4. `html2canvas(el, { scale: 2, useCORS: true, logging: false })`
5. 还原 `display`
6. canvas → jpeg dataUrl → `jsPDF` A4 portrait
7. 按 A4 高度切片 `pdf.addPage()` 多页拼接
8. `pdf.save(fileName)`
9. `finally` 块 `setExportingPDF(false)` + 还原 display

## 4. 数据流

```
detail (CompatibilityDetail)
  │
  ├─ reading           ─┬─→ ShareCard       (overall_score / dimension_scores)
  │                     └─→ PrintLayout     (decision/stageRisks 派生)
  │
  ├─ participants[]    ─┬─→ ShareCard       (两人四柱 + 性别 + display_name)
  │                     └─→ PrintLayout     (§1 hero / §7 双盘原图)
  │
  ├─ evidences[]       ─┬─→ ShareCard       (top 3 by weight)
  │                     └─→ PrintLayout     (§4 完整表格)
  │
  └─ latest_report?.content_structured
                       ─┬─→ ShareCard       (verdict 一句话)
                        └─→ PrintLayout     (§6 chapters 六章节)

brand (ExportBrand | null)
  ├─→ ShareCard.brand
  ├─→ PrintLayout.brand
  └─ 决定 title / footer / 是否水印
```

## 5. 错误处理

| 错误源 | 检测 | UI 反馈 |
|---|---|---|
| 未生成 AI 报告 | `structuredReport == null` | 按钮 disabled + title 提示 |
| `participants.length !== 2` | 组件 props 校验 | 渲染空状态卡 |
| 字体未加载完 | `await document.fonts.ready` | 阻塞至 fonts ready |
| `html-to-image` 失败 | try/catch | `alert('生成图片失败，请稍后重试')` |
| iOS 用户取消分享 | `err.message.includes('AbortError'/'cancel')` | 静默（不报错） |
| iOS Web Share 不可用 | `navigator.canShare` 检测 | 退化到 Blob URL 下载 |
| `html2canvas` 失败 | try/catch | `alert('生成 PDF 失败，请稍后重试')` |
| 桌面打印取消 | 浏览器原生 | 无反馈 |
| `brandAPI.get()` 失败 | `.catch(() => setBrand(null))` | 用默认 brand |
| 重复点击 | 按钮 `disabled={savingImage || exportingPDF}` | 防抖 |

**Finally 必须执行**：移动 PDF 流程的 `el.style.display = prevDisplay` 和 `setExportingPDF(false)` 必须在 finally 块里，避免失败后状态卡死。

## 6. 验证标准（实施完成后逐条验）

### 功能（10 条）

1. 进入合盘结果详情页，能看到「分享图片」「导出 PDF」两个按钮
2. 未生成 AI 报告时，两按钮 disabled + hover 提示「请先生成命理解读」
3. 生成 polished 报告后两按钮立刻可点（无需刷新）
4. 点「分享图片」打开 modal；点遮罩 / ESC / × 都能关闭；关闭后焦点回到原按钮
5. Modal 内 ShareCard 显示双盘四柱、综合分、四维评分条、3 条核心证据、verdict、brand title/footer
6. 桌面点 modal 内「保存到本地」→ 下载 `缘聚合盘-{self}-{partner}.png`，字体不发虚
7. iOS Safari 点「保存 / 分享」→ 系统分享面板出现，可选"存储图像"
8. Android 点「保存」→ 浏览器下载 .png
9. 桌面点「导出 PDF」→ 打印对话框打开，预览中 §1–§7 全部章节，强制分页，"另存为 PDF" 得到 5–8 页文件
10. 移动端点「导出 PDF」→ loading 状态显示「生成中…」→ 几秒后下载 `缘聚合盘-{self}-{partner}.pdf`

### 视觉/版面（5 条）

11. 分享卡宽度精确 400px，国风纸色 `#fdf9f2`，分隔线 `#d4b896`，与 mockup C 详版一致
12. PDF 章节强制分页（§1 到 §7 每节首页都从新页起）
13. PDF §7 双盘原图在 A4 单页内左右分列，宽度均等
14. PDF 每页都有 brand title 页眉 + footer
15. `showDiagonalWatermark(brand)` 为真时，分享卡和 PDF 都有对角水印

### 错误 / 边界（5 条）

16. 未生成报告时 disabled 按钮真的不响应点击，console 无报错
17. iOS 用户取消系统分享面板不弹 alert，UI 静默回到 modal
18. 人为破坏 `.compat-print-layout` 后点 PDF，看到 `alert('生成 PDF 失败，请稍后重试')` + loading 复位
19. 失败 alert 后按钮恢复可点，重试能正常导出
20. Modal 打开时按 Ctrl+P，打印预览里没有 modal 遮罩污染

### 性能（3 条）

21. 详情页 FCP / TTI 与旧详情页一致（PrintLayout 常驻但 display:none 不参与布局）
22. Modal 首次打开 < 500ms（ShareCard 挂载 + 字体加载 + 渲染）
23. 移动 PDF 一份完整报告 < 8s 完成

### 回归（2 条）

24. 单盘 `ResultPage` 的图片/PDF 按钮、ShareCard、PrintLayout 行为完全不变
25. `npm run lint` 和 `npm run build` 全绿

## 7. 关键决策摘要

| 决策 | 选项 | 取值 | 理由 |
|---|---|---|---|
| 导出入口 | 详情页 / 档案列表 / 两处 | **仅详情页** | 最小改动，唯一可信数据来源 |
| 图片 vs PDF 分工 | 分享卡/完整报告 / 同内容不同格式 / 仅 AI 报告 | **图片=分享卡 / PDF=完整报告** | 与单盘对齐，场景区分清晰 |
| PDF 章节范围 | 全部 1–7 / 核心 1–3 + AI / 仅 AI | **全部 1–7** | 完整报告体验 |
| PDF 技术路线 | 双轨复用 / 统一 jsPDF / 后端生成 | **双轨复用** | 桌面矢量质量优，与单盘一致 |
| 分享卡版面 | A 极简 / B 中等 / C 详版 | **C 详版** | 含 3 条核心证据，更像报告封面 |
| 可用时机 | 都等 AI / 图片随时 / 都随时 | **都等 AI** | 与单盘体验一致 |
| 组件抽离粒度 | A 独立新建 / B 改造单盘 / C 抽公共 | **A 独立新建** | 符合 CLAUDE.md §5.1 第 2 次保持原样 |
| 分享卡呈现方式 | 屏外渲染 / 常驻预览 / 点预览展开 | **modal 弹框** | 用户能预览真实样子，截图可靠 |

## 8. 文件清单

```
新增（4 个）:
  frontend/src/components/CompatibilityShareCard.tsx     (~200 行)
  frontend/src/components/CompatibilityShareCard.css     (~150 行)
  frontend/src/components/CompatibilityPrintLayout.tsx   (~500 行)
  frontend/src/components/CompatibilityPrintLayout.css   (~250 行)

修改（2 个）:
  frontend/src/pages/CompatibilityResultPage.tsx         (+~100 行)
  frontend/src/pages/CompatibilityResultPage.css         (+~80 行)
```

## 9. 不引入的依赖

- 后端生成（wkhtmltopdf / Chrome headless）
- 新的 PDF 库（如 react-pdf / pdfkit）
- 公共抽象组件（Pillars / PrintHeader / DiagonalWatermark 下沉到 `components/print/`）
