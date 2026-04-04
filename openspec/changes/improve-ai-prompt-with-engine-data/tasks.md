## 1. 重构 buildBaziPrompt 函数

- [x] 1.1 在 Prompt 中新增「十神关系」区块：注入四柱天干十神（`YearGanShiShen` / `MonthGanShiShen` / `DayGanShiShen` / `HourGanShiShen`）
- [x] 1.2 在 Prompt 中新增地支主气十神：注入 `YearZhiShiShen[0]` / `MonthZhiShiShen[0]` / `DayZhiShiShen[0]` / `HourZhiShiShen[0]`（取主气，即第一个元素）
- [x] 1.3 在 Prompt 中新增「十二长生（日主星运）」区块：注入 `YearDiShi` / `MonthDiShi` / `DayDiShi` / `HourDiShi`
- [x] 1.4 在 Prompt 中新增「神煞」区块：注入四柱 `ShenSha`，空数组显示「无」
- [x] 1.5 在 Prompt 中新增「大运序列」区块：注入完整10步大运（干、支、`GanShiShen`、`ZhiShiShen`、`StartAge`、`StartYear`）
- [x] 1.6 在 Prompt 中新增「引擎初步推算」区块：注入 `result.Yongshen` / `result.Jishen`（若非空），标注为参考

## 2. 重构 CoT 推理步骤

- [x] 2.1 将原「第一步：在心中完成专业分析（判断月令/得令/身强弱/用神）」改为「综合精算数据整合判断」措辞
- [x] 2.2 新的第一步明确三项整合任务：① 结合月令十神和日主星运评估得令情况；② 参考引擎初步用神确认或微调；③ 归纳相关神煞特质
- [x] 2.3 保留「在心中完成，不在报告中输出」的指令

## 3. 报告章节扩展

- [x] 3.1 在第二步报告格式中新增第五章「【大运走势】」
- [x] 3.2 为【大运走势】章节提供写作指引：结合起运年龄和各大运十神，解读人生各阶段，重点分析当前及近1~2步大运对事业/感情的影响

## 4. 参数调整

- [x] 4.1 将 `callOpenAICompatible` 中的 `MaxTokens` 从 `2000` 改为 `3500`

## 5. 验证

- [x] 5.1 `go build ./cmd/api/` 编译通过，无错误无警告
- [x] 5.2 验证脚本确认引擎精算数据完整：十神/十二长生/神煞/旬空/大运/初步用神均有值
- [x] 5.3 `yongshen` / `jishen` 字段由引擎推算并注入 Prompt，AI 返回后解析链路不变
- [x] 5.4 神煞验证通过（天乙贵人、将星、太极贵人、劫煞均正确计算）
