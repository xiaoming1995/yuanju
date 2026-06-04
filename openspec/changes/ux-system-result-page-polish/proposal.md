## Why

当前前端已经完成 MVP 功能，但关键页面的体验和样式开始分散：结果页内容过长且重点不够靠前，页面间按钮、状态、反馈和布局规则不一致，后续继续叠功能会进一步放大维护成本。

本变更先处理已确认审计蓝图中的 Batch 1 + Batch 2：建立最小设计系统闭环，并把结果页调整为「先结论、再依据、最后行动」的阅读结构。

## What Changes

- 新增前端基础 UI primitives：页面容器、区块容器、按钮、分段导航、状态标签、空状态、确认弹窗、Toast、表单字段。
- 扩展全局 CSS variables，使颜色、字号、间距、圆角和状态色有统一来源。
- 将结果页首屏改为结论优先：展示命盘身份摘要、核心结论、主行动和次行动。
- 为结果页建立分段导航：总览、命盘、用神、大运、AI 解读。
- 移动端结果页提供始终可达的主行动入口，避免用户滚到底寻找按钮。
- 收敛结果页与历史 / 过往事件之间的行动文案和跳转关系。
- 不引入 UI 框架，不改变八字算法、后端 API、AI prompt、鉴权或数据模型。

## Capabilities

### New Capabilities

- `frontend-ui-foundation`: 统一前端基础视觉 token、通用 UI primitives、反馈组件和表单状态规则。
- `result-page-decision-first`: 结果页采用结论优先的信息架构，首屏展示核心结论和主行动，并通过分段导航承载专业细节。

### Modified Capabilities

- 无。本变更新增前端 UX 能力，不改变既有后端业务能力规格。

## Impact

- 主要影响 `frontend/src/index.css`、`frontend/src/components/ui/*`、`frontend/src/pages/ResultPage.tsx`、`frontend/src/pages/ResultPage.css`。
- 次要影响历史、过往事件入口相关页面：`frontend/src/pages/HistoryPage.tsx`、`frontend/src/pages/PastEventsPage.tsx`。
- 需要新增或调整前端测试，至少覆盖 UI primitives 的源码约束、结果页首屏结构、移动端 CTA 和分段导航。
- 验证命令为 `cd frontend && npm run lint` 与 `cd frontend && npm run build`。
