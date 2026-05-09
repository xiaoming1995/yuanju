## Why

经专业命理师审核，现有算法有两处核心缺陷：扶抑判断用五行比例（>40%阈值）粗判强弱，精度不足，与传统命理"根→比劫→印"优先级体系不符；过往事件推算的底色逻辑只看流年天干五行类别是否属用神，未检测原局位置，漏掉了"刑冲克合穿破用神/忌神位"这一传统命理应期判断的核心方法。

## What Changes

- **修改** `calcWeightedYongshen()`：以基于位置的规则判断取代五行比例>40%阈值法，实现有根/无根→比劫→印的优先级扶抑体系
- **新增** 藏干（主气+中气）查询能力：有根/比劫/印的判断均需读取地支中气，当前代码只存主气
- **修改** `getYongshenBaseline()`：不再输出"全年底色"，改为检测流年/大运干支对原局用神/忌神所在位置的刑冲克合穿破，每条交互独立输出信号
- **新增** 六害（穿破）关系表：现有代码缺少六害对照表，应期判断需要
- **移除** `getYongshenBaseline()` 给其他信号染底色的机制：各信号极性独立，不再受全年基底色影响

## Capabilities

### New Capabilities

- `fuyi-strength-engine`：基于有根/无根→比劫多寡→印旺三段优先级判断日主强弱，取代五行比例阈值法
- `yingqi-position-signals`：检测流年/大运干支刑冲克合穿破原局用神/忌神所在干支位置，独立输出应期信号

### Modified Capabilities

- `bazi-precision-engine`：扶抑结果（身强/弱）的判断逻辑变更，yongshen/jishen 派生方式同步调整

## Impact

- `backend/pkg/bazi/yongshen.go`：替换 `calcWeightedYongshen()`
- `backend/pkg/bazi/event_signals.go`：替换 `getYongshenBaseline()`，新增六害表，新增位置信号检测函数
- `backend/pkg/bazi/engine.go`：补充藏干中气数据结构（`ZhiZhongQi`），供扶抑和应期判断使用
- 下游不受影响：`getYearEventSignals()` 的其他信号层（十神透干、地支关系、神煞、伏吟反吟、空亡）逻辑不变
