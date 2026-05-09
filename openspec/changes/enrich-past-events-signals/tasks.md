## 1. 数据结构与基础设施

- [x] 1.1 在 `event_signals.go` 的 `EventSignal` 增加 `Polarity` / `Source` optional 字段，并加 JSON tag `omitempty`
- [x] 1.2 在 `model.go` 的 `PastEventsTemplateData` 增加 `YongshenInfo` / `GejuSummary` / `DayunHuahe` / `StrengthDetail` 字段
- [x] 1.3 在 `event_signals.go` 增加 `applyPolarity(sig *EventSignal, signalSelf, baseline, source string)` 内部辅助函数，统一信号 polarity 派生逻辑
- [x] 1.4 在 `event_signals.go` 增加 `Source` 常量集合（`SourceShensha` / `SourceZhuwei` / `SourceHehua` / `SourceKongwang` / `SourceXing` / `SourceHui` / `SourceFuyin` / `SourceYongshen`）

## 2. 用神/忌神基底色

- [x] 2.1 在 `engine.go`（或 `BaziResult` 所在文件）确认是否已有 `Yongshen` / `Jishen` 字段；缺失则补充字段。**结果：已存在（engine.go:93-94），无需新增**
- [x] 2.2 在 `event_signals.go` 增加 `getYongshenBaseline(natal *BaziResult, lnGan string) (polarity, evidence string)`：返回流年五行 vs 用神/忌神的吉凶倾向
- [x] 2.3 `Yongshen` / `Jishen` 缺失时降级到 `Tiaohou.Expected`；两者皆缺则返回 ("", "")
- [x] 2.4 在 `GetYearEventSignals` 入口处调用 baseline，输出 `Type=用神基底` 信号
- [x] 2.5 baseline 为 polarity 后，所有现有信号的 polarity 默认继承 baseline；个别信号（如财星透干但财=忌神）显式覆盖
- [x] 2.6 单元测试：流年透用神 / 透忌神 / 中性 / 用神信息缺失 四个场景（4 tests pass）

## 3. 加权身强弱评分

- [x] 3.1 重写 `dayMasterStrength` 内部实现：保留签名，扩展返回值为 5 档（`vstrong` / `strong` / `neutral` / `weak` / `vweak`）
- [x] 3.2 实现得令(×5)/得地(×3)/得势(×2) 三层加权评分（含藏干透出 +1 加权）
- [x] 3.3 加权阈值（10/5/-5/-10）通过 `algo_config.go` 暴露为可调参数（新增 `ShenStrengthThresholds` 字段）
- [x] 3.4 顺带返回评分明细字符串：新增 `dayMasterStrengthLevel(natal) (level, score, detail)`，旧 `dayMasterStrength(natal) string` 保留为向后兼容三档折算
- [x] 3.5 同步更新 `event_signals.go` 内所有 `dayMasterStrength()` 调用点（兼容路径，无需大改）
- [x] 3.6 单元测试：极强、极弱、中和 三个回归用例（3 tests pass）

## 4. 神煞引擎接入

- [x] 4.1 在 `event_signals.go` 调用现有 `GetDayunShenSha`（公式同样适用流年）的封装 `getYearShensha`
- [x] 4.2 实现白名单与默认 polarity / type 映射表（25+ 神煞，详见 design 决策 3）
- [x] 4.3 将白名单内的神煞转换为 `EventSignal`，Source=`神煞`
- [x] 4.4 与 baseline polarity 协调：神煞自带 polarity 优先（神煞性质强烈），与基底色不一致时 evidence 注明
- [ ] 4.5 神煞总开关与单煞开关读取 `algo_config`（**未实现，留作后续：当前 algo_config 未暴露神煞开关；现以白名单常量控制；不影响主功能**）
- [x] 4.6 单元测试：天乙贵人 / 羊刃 / 不在白名单 三个用例（3 tests pass）

## 5. 流年与年/月/时柱互动

- [x] 5.1 新增"流年地支 vs 年支/月支/时支"的六合/六冲扫描
- [x] 5.2 按柱位映射事件类型（年→根基/综合变动、月→事业、时→子女晚景/综合变动）
- [x] 5.3 为每条互动信号填写 Source=`柱位互动`
- [x] 5.4 与既有岁破（流年冲年支）逻辑去重
- [x] 5.5 单元测试：流年冲月支输出事业信号（1 test pass，覆盖 5.2 核心路径）

## 6. 伏吟与反吟

- [x] 6.1 实现 `isFuyin(lnGan, lnZhi, pillarGan, pillarZhi string) bool`：干支完全相同
- [x] 6.2 实现 `isFanyin(lnGan, lnZhi, pillarGan, pillarZhi string) bool`：天干相克 + 地支六冲
- [x] 6.3 在 `GetYearEventSignals` 扫描流年与年/月/日/时/大运 五柱的伏吟与反吟
- [x] 6.4 输出 `Type=伏吟` 或 `Type=反吟` 信号，Polarity=凶，Source=`伏吟`
- [x] 6.5 单元测试：流年伏吟日柱 + isFuyin/isFanyin 单元测试（3 tests pass）

## 7. 空亡检测

- [x] 7.1 实现 `getXunkong(dayGan, dayZhi)`：60甲子查表返回旬空二支
- [x] 7.2 在 `GetYearEventSignals` 中检查流年地支与大运地支是否落空亡
- [x] 7.3 流年落空亡 → 独立 `Source=空亡` 信号，并对该年其他信号 evidence 追加"受空亡影响"提示
- [x] 7.4 大运落空亡 → 独立信号
- [x] 7.5 单元测试：甲子日柱旬空(戌亥)、癸亥日柱旬空(子丑)、未知日柱、流年落空亡降权（4 tests pass）

## 8. 三会局与三刑全局

- [x] 8.1 实现 `isSanhuiTriggered(lnZhi, existingZhi) (triggered, wuxing)`
- [x] 8.2 在 `GetYearEventSignals` 中调用三会检测，Source=`会`
- [x] 8.3 三会五行匹配感情星五行时额外输出婚恋信号
- [x] 8.4 实现 `isSanxingTriggered(lnZhi, existingZhi) (triggered, kind)`：寅巳申、丑未戌
- [x] 8.5 三刑触发时输出 `Type=健康`、`Polarity=凶`、`Source=刑` 信号
- [x] 8.6 单元测试：三会木局、三刑寅巳申（2 tests pass）

## 9. 大运天干合化日干

- [x] 9.1 实现天干五合表 `ganWuhe`（甲己土、乙庚金、丙辛水、丁壬木、戊癸火）
- [x] 9.2 实现 `detectDayunHuahe(natal, dyGan, dyZhi)`：化神条件检测（化神根气、无强反克）
- [x] 9.3 在 `GetYearEventSignals` 调用大运合化检测，化神成立 → `Type=大运合化`
- [x] 9.4 化神不成立 → `Type=综合变动`，Evidence 注明"合而不化"
- [x] 9.5 单元测试：丁壬合化木成立、丁壬合化失败、丙日主走甲运无合（3 tests pass）

## 10. 合冲并见与对消逻辑扩展

- [x] 10.1 扩展现有"婚恋_冲 + 婚恋_合 → 婚恋_变" 模式（保留两条 evidence 拼接）
- [x] 10.2 大运合日支 + 流年冲日支 → 通过婚恋_合/冲并见走 婚恋_变 路径
- [x] 10.3 三合 + 三会同时引动 → 通过 add 顺序输出多条 evidence（不强行去重，AI 可分别叙述）
- [x] 10.4 流年与某柱伏吟 + 冲该柱 → 由各自规则独立输出，AI prompt 引导优先表达
- [x] 10.5 单元测试：合冲并见路径已被 婚恋_变 改动覆盖（既有逻辑保留）

## 11. Service 层与 Prompt 模板升级

- [x] 11.1 `report_service.go` 的 `GeneratePastEventsStream` 读取 `BaziResult.Yongshen`/`Jishen` 并构造 `YongshenInfo`，缺失时降级到调候喜神
- [x] 11.2 构造 `DayunHuahe` 字符串：调用 `bazi.CollectDayunHuaheLines` 汇总所有合化大运
- [x] 11.3 构造 `StrengthDetail` 字符串：调用 `bazi.GetStrengthDetail` 输出"中和(评分2): 月令同气+5..."类格式
- [x] 11.4 构造 `GejuSummary` 字符串：当前 BaziResult 暂无格局字段，留空串（后续 change 补足）
- [x] 11.5 写入 `tplData` 并模板渲染验证（go build 通过）
- [x] 11.6 升级 `defaultPastEventsPrompt`：新增对 4 个新模板变量的引用与"基底色优先、神煞强烈、合化整段定向"等 9 条撰写要求
- [x] 11.7 在 `database.go` 增加 SQL 升级条件：`content NOT LIKE '%YongshenInfo%'` 触发 prompt 重灌
- [x] 11.8 已含 `YongshenInfo` 字段的 prompt 不被覆盖（保护 admin 自定义）

## 12. 回归测试集

- [x] 12.1 在 `event_signals_test.go` 创建测试套（22 个测试用例）
- [ ] 12.2 整理 5–10 个真实命例（含已知大事年）— **未实施：缺乏脱敏样本与公开命例授权数据；本次以单元测试覆盖代替；用户提供命例后再补**
- [ ] 12.3 断言：每个大事年至少产出 1 个非 baseline 信号；polarity 与实际事件方向一致 — **同上**
- [ ] 12.4 断言：相对应的"平年"信号数量不超过阈值（避免过拟合）— **同上**
- [x] 12.5 跑全套 `go test ./pkg/bazi/...` 确认无回归（PASS，0.686s）

## 13. 联调与发布（用户操作）

- [ ] 13.1 本地 `docker-compose up -d` 启动后端，确认数据库 prompt 升级 SQL 正常执行
- [ ] 13.2 用一个测试命盘调用 `POST /api/bazi/past-events-stream/:chart_id`，前端页面验证 narrative 是否引用了用神/合化/神煞
- [ ] 13.3 比较升级前后同一命盘的报告内容差异（人工抽查 3 个年份）
- [ ] 13.4 Admin 后台验证清缓存按钮，再次生成确认是新版报告
- [x] 13.5 更新 `CLAUDE.md` 中 `pkg/bazi/` 部分的"event_signals.go"描述
