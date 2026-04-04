## 1. 样式表与动画配置

- [x] 1.1 在 `ResultPage.css` 中移除原本属于 `report-loading-area` 下 `.skeleton` 相关的旧冗余代码结构（如果不再需要）。
- [x] 1.2 在 `ResultPage.css` 中新增 `ai-loading-container`、`ai-loading-step` 等容器类。
- [x] 1.3 在 `ResultPage.css` 中新增动态呼吸灯和淡入淡出动效（例如 `@keyframes stepFadeIn` 和 `pulse-glow` 调整版）。

## 2. 前端组件实现

- [x] 2.1 在 `frontend/src/pages/ResultPage.tsx` 中定义一个进度文案数组 `LOADING_STEPS`，内容依次为：「☯️ 飞盘排签，提取四柱大运神煞...」、「🔍 校准星运，结合真太阳时精算...」、「📚 翻阅《子平真诠》推断月令格局...」、「🌙 对照《穷通宝鉴》抓取调候用神...」、「✒️ 宗师沉思，正在精排你专属的命局详析...」。
- [x] 2.2 在 `ResultPage` 内新增 `loadingStepIndex` 的 React state，初始值为 0。
- [x] 2.3 在 `ResultPage` 内利用 `useEffect` 在 `reportLoading === true` 时启动一个定时器（例如 `setInterval` 4秒递增步数），最多递增到数组最后一个元素的索引位置，在停止/完成 loading 时清除定时器。
- [x] 2.4 在 `ResultPage.tsx` 的 `{reportLoading && (...)}` 渲染块中，替换掉原本的五条灰显 `skeleton`，改为渲染动态文案组件：带有高亮的主副标题以及随着 `loadingStepIndex` 切换而淡入显示的当前进度文案。

## 3. 功能验证

- [ ] 3.1 前端起服验证：以普通用户或未登录态重试生成命理报告。
- [ ] 3.2 观察在点击按钮之后的数秒内，进度播报是否按照每几秒一条平滑推进，且抵达最后一条后不报错，最终报告成功加载并替换掉 loading 屏。
