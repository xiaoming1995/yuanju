## MODIFIED Requirements

### Requirement: 年份信号扫描范围
系统 SHALL 扫描命主所有流年（起运年至命盘终点）的命理信号，不再限于过往年份。

#### Scenario: 未来年份信号扫描
- **WHEN** `GetAllYearSignals` 被调用
- **THEN** 返回的 `YearSignals` 列表包含过往和未来所有流年，按年份升序排列

#### Scenario: 最小年龄过滤仍生效
- **WHEN** `GetAllYearSignals` 被调用
- **THEN** 起运前（age < minAge）的年份仍被过滤，不出现在结果中

### Requirement: DayunList 传入范围
系统 SHALL 将全部大运（包含未来大运）的信息传入 Prompt 模板变量 `{{.DayunList}}`，不再以当前年份截断。

#### Scenario: DayunList 包含未来大运
- **WHEN** `GeneratePastEventsStream` 构建 Prompt 数据
- **THEN** `DayunList` 包含所有大运条目，包括 StartYear > currentYear 的未来大运
