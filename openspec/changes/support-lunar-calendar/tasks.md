## 1. 后端接口升级（bazi_handler / model）

- [x] 1.1 在 `backend/internal/handler/bazi_handler.go` 中，为 `CalculateInput` 增加 `CalendarType` (`string`) 和 `IsLeapMonth` (`bool`) 字段的接收绑定
- [x] 1.2 在请求参数验证处，确保如果传入了 `CalendarType`，只能是 `solar` 或 `lunar`，且为空时默认取 `solar`
- [x] 1.3 顺传这 2 个新增参数进入 `service.CalculateBazi` -> `bazi.Calculate` 方法链路内

## 2. 后端核心历法预处理引擎（bazi/engine.go）

- [x] 2.1 修改 `bazi.Calculate` 的函数签名，追加 `calendarType` 与 `isLeapMonth` 两个强类型入参
- [x] 2.2 在 `bazi.Calculate` 入口第一步判断：如果输入是 `lunar`，则使用 `calendar.NewLunar` 构造农历日期（当 `isLeapMonth` 为真时，将 `month` 转负值）
- [x] 2.3 使用 `.GetSolar()` 将上述农历转化为公历，把推演所得的 `GetYear()`, `GetMonth()`, `GetDay()` 直接赋给后续真太阳时计算所需基础变量 `calcYear`, `calcMonth`, `calcDay`
- [x] 2.4 测试用例断言：通过调用带 `calendar=lunar` 的农历日子（如 2020 闰 4 月），确保得出的八字跟按这天对应公历推演的八字完全一致

## 3. 前端界面重构与交互依赖（HomePage.tsx）

- [x] 3.1 前端工程中引入 `lunar-javascript` (执行 npm install lunar-javascript并确认 tsx 中可引用)
- [x] 3.2 在 `HomePage.tsx` 的 state `form` 中，增加 `calendarType` (`solar` 或 `lunar`) 以及内部维度的 `isLeapMonth` 布尔状态
- [x] 3.3 界面增加 `radio` 或者两个 `tab` 按钮：公历、农历（默认置停在公历上）
- [x] 3.4 **基于农历日历计算动态选项**：编写推导函数，当处于农历环境且给定 `form.year` 时，渲染 1~12 月的所有选项；并穿插检测是否存在闰月（如：若有闰四月，下拉列表即展示“4月”、“闰4月”等合计13个 option 项）
- [x] 3.5 **大小月日历自适应限制**：基于 `form.year` 及 `form.month`(含闰月标记)，精确推演该农历月的天数为 29天(小建) 抑或 30天(大建)；动态调整「日」下拉菜单选项数组并处理安全回退截断逻辑
- [x] 3.6 发起计算请求 API 时，携带最新的 `calendar_type` 与 `is_leap_month` 发给后端

## 4. 联调验证

- [x] 4.1 发起首版标准公历的命盘创建查询，确保向下兼容完全通过
- [x] 4.2 发起标准农历命盘查询（平年非闰月）并与预期八字做核对
- [x] 4.3 发起带闰月的罕见农历命盘排盘查询并核对
