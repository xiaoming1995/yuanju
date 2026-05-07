## ADDED Requirements

### Requirement: 流年事件信号检测
系统 SHALL 提供 `GetYearEventSignals(natal BaziResult, lnYear int, lnGanZhi string, dayunGanZhi string, gender string) []EventSignal` 函数，对指定流年计算激活的事件信号列表。

每个 `EventSignal` 包含：
- `Type`：事件类型（婚恋/事业/财运_得/财运_损/健康/迁变）
- `Evidence`：触发本信号的命理证据描述（如"偏财星庚透干"、"流年子与日支午相冲"）

#### Scenario: 男命财星透干产生婚恋信号
- **WHEN** 男命流年天干为日干之偏财或正财
- **THEN** 返回信号列表中包含 Type=婚恋、Evidence 说明该财星透干及其含义

#### Scenario: 女命官星透干产生婚恋信号
- **WHEN** 女命流年天干为日干之正官或七杀
- **THEN** 返回信号列表中包含 Type=婚恋、Evidence 说明该官星透干及其含义

#### Scenario: 流年地支与日支六合引动夫妻宫
- **WHEN** 流年地支与日柱地支构成六合关系
- **THEN** Evidence 中注明"夫妻宫（日支）合住，感情宫位被激活"

#### Scenario: 流年地支与日支相冲引动夫妻宫
- **WHEN** 流年地支与日柱地支构成六冲关系
- **THEN** Evidence 中注明"夫妻宫（日支）受冲，感情宫位震动"

#### Scenario: 日主受克产生健康信号
- **WHEN** 流年天干五行克制日柱天干五行
- **THEN** 返回信号列表中包含 Type=健康、Evidence 说明克制关系

#### Scenario: 无信号激活年份
- **WHEN** 该流年未触发任何规则
- **THEN** 返回空信号列表（[]EventSignal{}），不返回 error

---

### Requirement: 批量扫描所有过往流年
系统 SHALL 提供 `GetPastYearSignals(result *BaziResult, gender string, currentYear int, minAge int) []YearSignals` 函数，对从 minAge 岁起至 currentYear（含）的所有流年批量调用信号检测，返回有序列表。

每个 `YearSignals` 包含：
- `Year`：公历年份
- `Age`：虚岁
- `GanZhi`：流年干支
- `DayunGanZhi`：所在大运干支
- `Signals`：该年 EventSignal 列表

#### Scenario: 过滤最小年龄
- **WHEN** 命主某些流年对应虚岁小于 minAge
- **THEN** 这些年份不出现在返回列表中

#### Scenario: 不包含未来年份
- **WHEN** 流年年份大于 currentYear
- **THEN** 该年份不出现在返回列表中

#### Scenario: 返回顺序
- **WHEN** 调用 GetPastYearSignals
- **THEN** 返回列表按年份升序排列
