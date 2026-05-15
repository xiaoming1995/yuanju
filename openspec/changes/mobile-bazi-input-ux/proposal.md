## Why

当前命盘输入把基础信息、专业校准项和命理术语放在同一层级展示，移动端尤其容易形成理解负担。用户的真实目标是快速、准确地完成起盘输入，因此需要重构输入体验，在保留专业校准能力的同时降低首次填写成本。

## What Changes

- 将命盘输入重构为移动优先的基础流程：性别、历法、出生日期、出生时辰优先展示。
- 增加提交前摘要，让用户确认“性别 + 历法 + 日期 + 时辰 + 校准方式”后再起盘。
- 将出生地、真太阳时、早子时等专业项收敛到“高级校准”区域，默认不干扰基础起盘。
- 优化子时细分表达，避免用单个 checkbox 承载“23:00-23:59 按前一日 / 00:00-00:59 按当日”的关键差异。
- 调整移动端布局与控件尺寸，确保手机上表单可扫读、可点击、不会出现拥挤或按钮被埋在长表单下方的问题。
- 复用并升级 `BirthProfileForm`，让首页起盘和合盘输入在基础字段体验上保持一致。

## Capabilities

### New Capabilities

- `bazi-input-ux`: 覆盖用户在前端输入八字命盘出生信息时的移动端优先交互、确认摘要和高级校准体验。

### Modified Capabilities

- `bazi-precision-engine`: 真太阳时校准能力在前端输入层的入口与文案发生变化，但后端计算契约保持不变。

## Impact

- 前端组件：`frontend/src/components/BirthProfileForm.tsx`、`frontend/src/components/BirthProfileForm.css`。
- 首页起盘：`frontend/src/pages/HomePage.tsx`、`frontend/src/pages/HomePage.css`。
- 合盘输入：`frontend/src/pages/CompatibilityPage.tsx` 可能需要接入升级后的通用出生信息输入体验。
- API：不新增后端接口，不改变 `/api/bazi/calculate` 和合盘请求字段结构。
- 依赖：不引入 UI 组件库，继续使用 React、TypeScript、CSS Variables 和现有日期计算依赖。
