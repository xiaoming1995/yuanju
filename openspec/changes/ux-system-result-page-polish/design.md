## Context

缘聚前端已经形成 MVP 级完整功能，但用户端和后台在多轮增量后出现样式和布局分散。全局 `index.css` 已有基础变量，但页面内仍存在大量 inline style、硬编码颜色和局部 CSS 体系。结果页尤其突出：`ResultPage.tsx` 与 `ResultPage.css` 行数较大，首屏同时承载命盘、专业细节和行动入口，用户需要滚动寻找重点。

本变更只覆盖已确认审计蓝图中的前两批：前端 UI foundation 与结果页结论优先结构。合盘、后台和 Auth/Profile 后续另起 change，避免一次性改动过大。

约束：

- 前端使用 React 18 + Vite + TypeScript。
- 样式继续使用 CSS Variables 和普通 CSS，不引入 Ant Design、MUI、TailwindCSS 等 UI 框架。
- 不改后端 API、八字算法、AI prompt、鉴权、数据库结构。
- 已确认的视觉方向见 `.superpowers/mockups/yuanju-high-fidelity-ux-mockup.html`。

## Goals / Non-Goals

**Goals:**

- 建立最小可用的共享 UI primitives，让后续页面优化有统一落点。
- 扩展全局 CSS token，收敛颜色、字号、间距、圆角和状态语义。
- 将结果页首屏改成「命盘摘要 + 核心结论 + 主行动 + 次行动」。
- 通过分段导航承载专业细节，减少首屏信息压力。
- 移动端结果页主 CTA 始终可达，避免用户滚到底寻找按钮。
- 历史和过往事件入口的文案、跳转和空状态与结果页行动保持一致。

**Non-Goals:**

- 不重做合盘入口、合盘结果、后台页面；这些属于后续批次。
- 不把结果页一次性拆成完全不同的信息架构或横滑章卷模式。
- 不修改分享卡和 PDF 导出逻辑，除非入口文案需要同步。
- 不追求清零所有 inline style；本变更只为后续清理建立基础件。
- 不引入新的品牌主色、字体系统或复杂动效。

## Decisions

### 1. 先建 UI primitives，再改结果页

先新增 `frontend/src/components/ui/` 下的基础件，再让结果页消费它们。这样结果页优化不会继续生成新的页面级局部组件，后续合盘和后台也能复用。

替代方案是直接在 `ResultPage.css` 内重写首屏样式。该方案短期更快，但会加重当前结果页局部设计系统的问题，因此不采用。

### 2. CSS token 扩展放在 `index.css`

全局 token 继续放在 `frontend/src/index.css`，原因是项目当前没有 CSS-in-JS 或主题 provider。新增 token 只补语义缺口，不迁移已有变量命名。

替代方案是新建 `theme.css` 并在入口引入。该方案更整洁，但会增加迁移面；当前更适合增量扩展。

### 3. UI primitives 保持轻量和无业务依赖

`PageShell`、`SectionPanel`、`Button`、`SegmentedTabs`、`StatusBadge`、`EmptyState`、`ConfirmDialog`、`Toast`、`FormField` 不读取业务上下文、不调用 API、不绑定命理数据结构。业务页面通过 props 组合。

这样做的代价是每个页面仍需写少量业务装配代码，但能防止组件过早变成庞大的领域组件。

### 4. 结果页采用线性分段导航，不引入横滑章节容器

结果页本次采用分段导航和锚点式信息结构：总览、命盘、用神、大运、AI 解读。移动端可横向滚动 tab，但内容仍是普通线性页面。

原因是本次目标是首屏减负和行动清晰，不是重新发明结果页交互。横滑章卷类结构会带来滚动、hash、打印、回退和手势复杂度，适合单独 change。

### 5. 移动端主 CTA 用底部固定条，但避开既有 BottomNav

结果页移动端新增主 CTA 固定区域时，必须考虑 `BottomNav`、safe area 和现有页面底部 padding。主 CTA 不得遮挡内容，也不得与底部导航抢同一层级。

桌面端不使用固定 CTA，主行动放在首屏和页面顶部内容区。

### 6. 测试以源码约束 + 构建验证为主

项目当前前端测试体系偏源码级检查和构建校验。本变更应新增轻量测试，验证：

- UI primitives 文件存在并导出预期组件。
- 关键组件不使用 inline style 作为主样式。
- 结果页包含首屏摘要、行动条和分段导航。
- 移动端 CTA 和分段导航有稳定 class hook。

最终仍需 `npm run lint` 和 `npm run build` 验证。

## Risks / Trade-offs

- [Risk] 新增 UI primitives 后短期组件数量增加。→ Mitigation: 只建最小闭环组件，不一次性迁移所有页面。
- [Risk] 结果页已有大量局部 CSS，改首屏可能引发样式冲突。→ Mitigation: 新组件使用清晰 class 前缀，保留旧结构，先新增再替换首屏。
- [Risk] 移动端固定 CTA 可能遮挡底部导航或内容。→ Mitigation: 明确 safe-area padding 和 `BottomNav` 偏移，并做 390px 人工验收。
- [Risk] 结论文案可能依赖 AI 报告或已有命理推断，数据缺失时首屏空洞。→ Mitigation: 首屏核心结论允许 fallback 到 deterministic summary，不阻塞页面渲染。
- [Risk] 共享组件抽象过早。→ Mitigation: primitives 只封装样式和结构，不封装业务规则。

## Migration Plan

1. 扩展全局 token，并新增 primitives，不改业务页面。
2. 为 primitives 增加源码级测试，保证导出和 class hook 稳定。
3. 在结果页新增首屏摘要与行动条组件，保留旧详情内容。
4. 添加结果页分段导航，将旧内容分组到稳定锚点。
5. 调整历史和过往事件入口文案与空状态。
6. 运行 lint/build，并做移动端 390px、桌面 1440px 人工验收。

Rollback 策略：由于不改后端和数据结构，若结果页改造有问题，可以回退结果页组件消费 primitives 的部分，保留基础 UI primitives 不影响现有页面。

## Open Questions

- 首屏「核心结论」是否优先使用现有 AI 报告摘要，还是只用 deterministic 命盘摘要作为第一版 fallback？建议实现时优先使用现有可得数据，不新增接口。
- 移动端底部 CTA 与当前 `BottomNav` 的具体高度是否在所有机型一致？实现时需要实际截图确认。
