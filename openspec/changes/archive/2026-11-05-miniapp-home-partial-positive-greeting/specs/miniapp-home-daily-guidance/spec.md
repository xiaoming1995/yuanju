## ADDED Requirements

### Requirement: 单边喜用使用轻正向文案
首页个人鼓励语 SHALL 区分双喜用、单边喜用且另一边中性、完全中性及含忌神的日期。单边喜用且另一边中性时 SHALL 显示轻正向语气，但 MUST NOT 声称确定好运或保证具体结果。

#### Scenario: 日干喜用、日支主气中性
- **WHEN** 有效个人结果为 `undetermined`，日干方向为 `favorable` 且日支主气为 `neutral`
- **THEN** 首页显示轻正向鼓励语，语气弱于双喜用结果

#### Scenario: 日支主气喜用、日干中性
- **WHEN** 有效个人结果为 `undetermined`，日支主气方向为 `favorable` 且日干为 `neutral`
- **THEN** 首页显示同等级的轻正向鼓励语

#### Scenario: 含忌神或完全中性
- **WHEN** 有效个人结果含 `adverse` 方向，或日干与日支主气均为 `neutral`
- **THEN** 首页不因单边喜用规则显示积极语句
