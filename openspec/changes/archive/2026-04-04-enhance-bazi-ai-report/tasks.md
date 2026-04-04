## 1. 调候引擎：新建查表模块

- [x] 1.1 新建 `backend/pkg/bazi/tiaohou.go`，实现 `tiaohouTable`（`map[string]map[string]string`，覆盖 10天干 × 12月支 = 120条规则，数据来源：《穷通宝鉴》）
- [x] 1.2 实现 `LookupTiaohou(dayGan, monthZhi string) string` 函数，未命中时返回空字符串
- [x] 1.3 在 `BaziResult` 结构体（`engine.go`）中新增 `Tiaohou string` 字段（json tag: `tiaohou`）
- [x] 1.4 在 `Calculate()` 函数末尾调用 `LookupTiaohou(dayGan, monthZhi)` 并赋值至 `BaziResult.Tiaohou`
- [x] 1.5 在 `engine_test.go` 中验证至少 5 个典型命例的调候结果（如甲木子月、壬水午月等）

## 2. Prompt 全面重写

- [x] 2.1 重写 `callOpenAICompatible()` 中的 System message：定义「现代解读风格」（通俗直接、结论先行、术语作精准点缀），移除旧版「避免晦涩术语」等模糊描述，添加「必须输出合法 JSON」格式要求
- [x] 2.2 重写 `buildBaziPrompt()` 数据区，在现有注入字段后新增「===调候用神（穷通宝鉴）===」区块：仅当 `r.Tiaohou != ""` 时注入，格式为「日主[X]生于[月]，调候用神：[内容]」
- [x] 2.3 重写 CoT 第一步「综合精算数据整合判断」：新增子任务 a（月令格局推断，含格局名称输出逻辑）和子任务 b（调候用神与格局用神整合，确认最终喜用神）
- [x] 2.4 重写第二步「命局分析总览（analysis.logic）」指令：移除「先写专业术语版推导」要求，改为「现代叙事风格」；要求总览中包含格局定性的结论句（如「整体而言，此命为XXX格，格局XXX」）
- [x] 2.5 重写第三步六章节指令，为每章添加命盘数据锚点：
  - 【性格特质】：指定参考日主五行特质、日支十二长生、比劫/印绶十神力量
  - 【感情运势】：指定参考男命财星/女命官杀的透干位置，日支星运，桃花/红鸾/天喜神煞
  - 【事业财运】：指定参考官杀/食伤的天干透出情况，天乙贵人/文昌神煞，财星与食伤的配合
  - 【健康提示】：指定参考五行过旺/过衰对应脏腑（木肝、火心、土脾、金肺、水肾），旬空地支
  - 【大运走势】：指定注入当前年份（`time.Now().Year()`），要求 AI 推算当前大运，重点解读当前步及下一步
  - 【命理分身】：指定优先匹配日主+格局名称+神煞特质相似的名人
- [x] 2.6 将 `MaxTokens` 从 4500 调整为 6000，`Temperature` 从 1.0 调整为 0.75（`callOpenAICompatible()` 函数）

## 3. 验证

- [x] 3.1 本地启动后端，调用 `POST /api/bazi/calculate` 验证响应中包含 `result.tiaohou` 字段且非空
- [x] 3.2 调用 `POST /api/bazi/report` 生成新报告，检查 Prompt 日志（或在测试中打印 Prompt），确认包含「调候用神」区块和当前年份注入
- [x] 3.3 检查返回报告的 `analysis.logic` 字段，确认包含格局名称结论句
- [x] 3.4 检查「大运走势」章节，确认明确指出当前大运步次
- [x] 3.5 检查各章节详细分析，确认有具体命盘数据支撑（非泛化通用内容）

