## Why

大运时间轴卡片当前只把天干按五行着色，并只显示天干十神；地支既未按自身五行区分，也缺少对应十神。用户无法在总览中完整辨认一柱大运的干支信息，需要进入下方内容才能补全判断。

## What Changes

- 为每张大运卡片分别识别天干与地支的五行，并以各自的五行颜色展示两个字。
- 在大运卡片中同时展示天干十神与地支十神，明确标识为“干”和“支”。
- 调整紧凑卡片的排版与响应式规则，保证十步大运总览在桌面和移动端清晰、不溢出。

## Capabilities

### New Capabilities

- `dayun-timeline-labels`: 在大运时间轴总览中完整、紧凑地展示干支五行和双十神标签。

### Modified Capabilities

- None.

## Impact

- 前端：`frontend/src/components/DayunTimeline.tsx` 的大运卡片，以及 `frontend/src/pages/ResultPage.css` 的卡片文本样式。
- 测试：补充时间轴展示的静态或组件级覆盖，验证地支五行与 `zhi_shishen` 均被渲染。
- 不涉及后端接口、计算算法或既有大运数据结构变更。
