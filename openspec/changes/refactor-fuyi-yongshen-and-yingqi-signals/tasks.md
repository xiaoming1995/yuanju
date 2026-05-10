## 1. 基础数据层：六害表 + 藏干工具函数

- [x] 1.1 在 `event_signals.go` 中新增 `sixHai`（六害/穿破）关系表：子未、丑午、寅巳、卯辰、申亥、酉戌
- [x] 1.2 新增 `hideGanMainZhong(hideGan []string) []string` 工具函数，取藏干数组前两项（主气+中气），余气不计

## 2. 扶抑算法重构（yongshen.go）

- [x] 2.1 新增 `calcFuyiStrength(natal *BaziResult) (isStrong bool, reason string)` 实现有根/无根判断逻辑
- [x] 2.2 在 `calcFuyiStrength` 中实现有根判断：遍历四柱地支藏干（主气+中气），任一含日主比肩则有根
- [x] 2.3 在 `calcFuyiStrength` 中实现有根分支：统计官杀+食伤（天干+四柱主气/中气）总数，<3→身强，≥3→身弱
- [x] 2.4 在 `calcFuyiStrength` 中实现印旺判断：地支主气/中气含印星≥2，或天干1个+地支1个
- [x] 2.5 在 `calcFuyiStrength` 中实现无根分支：比劫（天干+主气+中气）≥3 且 印旺→身强，否则身弱
- [x] 2.6 修改 engine.go fallback 分支：在调候未命中后以 `calcFuyiYongshen` 取代 `calcWeightedYongshen`
- [x] 2.7 编写 `calcFuyiStrength` 单元测试，覆盖：有根身强、有根身弱、无根身强（比劫多+印旺）、无根身弱四个场景

## 3. 应期位置信号（event_signals.go）

- [x] 3.1 新增 `collectYongshenPositions(natal *BaziResult) (yongPos, jiPos []string)` 收集原局用神/忌神覆盖的天干和地支列表（含藏干中气匹配的地支）
- [x] 3.2 新增 `collectYingqiSignals(natal *BaziResult, lnGan, lnZhi, dyGan, dyZhi string) []EventSignal` 实现刑冲克穿合五种交互检测
- [x] 3.3 在 `collectYingqiSignals` 中实现天干克检测（流年/大运天干克用神/忌神天干位）
- [x] 3.4 在 `collectYingqiSignals` 中实现地支冲检测（六冲对用神/忌神地支位）
- [x] 3.5 在 `collectYingqiSignals` 中实现地支刑检测（六刑、三刑、自刑对用神/忌神地支位）
- [x] 3.6 在 `collectYingqiSignals` 中实现地支穿检测（六害对用神/忌神地支位）
- [x] 3.7 在 `collectYingqiSignals` 中实现天干五合/地支六合检测，按化出五行属用神/忌神定极性，合而不化按锁定处理
- [x] 3.8 编写 `collectYingqiSignals` 单元测试，覆盖：克用神凶、冲忌神吉、穿用神凶、合用神（化忌）凶、合而不化凶

## 4. 集成：替换 GetYearEventSignals 中的底色逻辑

- [x] 4.1 在 `GetYearEventSignals` 开头调用 `collectYingqiSignals`，将返回的信号追加到 signals 中，替换原 `getYongshenBaseline()` 调用
- [x] 4.2 移除 `baseline` 变量的传递：将所有 `addP(..., baseline, ...)` 改为 `addP(..., "", ...)`，让各信号极性独立由 signalSelf 决定
- [x] 4.3 删除 `getYongshenBaseline` 的调用点（保留函数定义暂不删除，方便回滚）
- [x] 4.4 运行现有测试套件：`go test ./pkg/bazi/...`，修复因 baseline 移除导致的断言变化

## 5. 验证

- [x] 5.1 用1995年10月12日午时（乙亥 丙戌 丙子 甲午）验证2024年甲辰、2025年乙巳的信号输出，确认应期信号合理
- [x] 5.2 检查2025年「健康」信号不再出现 polarity=吉（岁破应为凶）的 bug — 已修复，巳冲亥（用神位）→凶
- [x] 5.3 人工复核扶抑结果：丙火日主生于戌月（乙亥丙戌丙子甲午），有根（时支午藏丁引），克泄4≥3，身弱，calcFuyiStrength 输出与命理解读公式一致
