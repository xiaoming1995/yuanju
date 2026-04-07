## 1. 数据结构扩展

- [x] 1.1 在 `jin_bu_huan_dict.go` 的 `JinBuHuanResult` 结构体中新增 `ZhiLevel`、`GanModifier`、`GanDesc` 字段
- [x] 1.2 在 `jin_bu_huan_dict.go` 的 `JinBuHuanRule` 结构体中新增 `EarthGanEval *JBHEval` 字段

## 2. 核心算法实现

- [x] 2.1 在 `jin_bu_huan_dict.go` 中实现 `dirToWuxing` 方向→五行映射辅助函数
- [x] 2.2 在 `jin_bu_huan_dict.go` 中实现 `ganToWuxing` 天干→五行映射辅助函数（可复用 shishen.go 中已有逻辑）
- [x] 2.3 在 `jin_bu_huan_dict.go` 中实现 Level 升降矩阵函数 `adjustLevel(zhiLevel, modifier string) string`
- [x] 2.4 修改 `CalcJinBuHuanDayun` 函数签名，增加 `dayunGan string` 入参
- [x] 2.5 实现天干五行与喜忌方向匹配逻辑（加成/减损/中性路径）
- [x] 2.6 实现土干（戊/己）专属处理：优先读 `EarthGanEval`，nil 则执行互克通关检测

## 3. 调用处更新

- [x] 3.1 更新 `engine.go` 中 `CalcJinBuHuanDayun` 的调用，传入 `gan` 参数

## 4. 前端展示

- [x] 4.1 在 `DayunTimeline.tsx` 中读取 `jin_bu_huan.gan_modifier`，在大运卡片右上角增加小徽章（加成/减损/通关/中性），颜色与 Level 联动

## 5. 验证

- [x] 5.1 在 `jin_bu_huan_dict.go` 相关测试或手动调用中验证：取一个典型命盘（如甲日子月），走南方火大运的天干=丙（喜）vs. 天干=壬（忌），确认 Level 和 GanModifier 符合预期
- [x] 5.2 验证土干（戊/己）的通关逻辑：取 GoodDirections 含南北互克的规则，走戊土大运，确认 GanModifier = "通关"

