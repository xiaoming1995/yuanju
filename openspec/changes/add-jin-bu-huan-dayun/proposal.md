## Why

当前系统通过五行比例和《穷通宝鉴》提供了原局基础素质的判定，但在大运流年生效周期的具体吉凶解读上缺乏算法级的理论背书，完全依赖底层大模型自由解读，极易产生“幻觉”或不严谨的断语。
引入经典命理文献《金不换大运歌》能构建极高精度查表系统，作为“确凿”标签注入大运阶段（通过东南西北行运定吉凶），极大提高大运命理分析的准确性和前台可视化程度。

## What Changes

- 在后端的引擎计算大运 (`DayunItem`) 时，解析对应的地支三会方向（如寅卯辰东方木，巳午未南方火）。
- 建立以“日主+月令”为 Key 的《金不换》核心判定字典文件（`jin_bu_huan_dict.go`）。
- 挂载金不换的解析结果，并在前端的时间轴上针对每一柱大运渲染具体的“金不换断语”（吉凶定性与词条）。

## Capabilities

### New Capabilities
- `dayun-jin-bu-huan`: 金不换大运测算引擎

### Modified Capabilities
- (None)

## Impact

- **Backend**: 修改 `pkg/bazi` 包下的计算引擎 `engine.go` 以挂载运算；新增 `jin_bu_huan_dict.go` 和计算逻辑；调整 `BaziResult` 和 `DayunItem` 回包 JSON 结构。
- **Frontend**: 更新 `ResultPage.tsx` 和 `DayunTimeline.tsx` 组件以接收并渲染新的评分与原诗文标签。
