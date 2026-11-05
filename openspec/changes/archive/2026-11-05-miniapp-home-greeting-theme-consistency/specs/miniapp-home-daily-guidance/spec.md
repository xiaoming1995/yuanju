## MODIFIED Requirements

### Requirement: 首页提供有依据的个人每日鼓励语
小程序首页 SHALL 在日期三柱之后、通用宜忌之前，为当前账号已绑定且资料有效的本人命盘展示一条由同日个人关系状态及可见四项本人每日主题共同生成的简短鼓励语及「查看依据」入口。文案 SHALL 保留关系复杂或需留意时的谨慎语气；四项主题全部为「平」时 MUST NOT 输出积极文案，任一主题为「慎」时 MUST NOT 输出积极文案。资料不完整时 MUST NOT 输出总括短句。文案 MUST NOT 把通用黄历宜忌转换为个人好运结论，也 MUST NOT 保证具体事件或鼓励高风险决策。展示 SHALL 不改变原有模块顺序与页面标题状态。

#### Scenario: 个人关系偏合且有明确主题
- **WHEN** 当前已绑定命盘的同日个人结果为 `supportive`，四项可见主题完整，至少一项为 `candidate` 且没有 `review`
- **THEN** 首页在三柱下方展示积极但不保证结果的短句，点击「查看依据」进入同日个人详情

#### Scenario: 四项主题全部为平
- **WHEN** 四项可见主题均为 `neutral`，即使当日关系为 `supportive` 或「日干偏合」
- **THEN** 首页只显示平稳短句，不声称今天顺风或好运

#### Scenario: 任一主题需留意
- **WHEN** 四项可见主题中有 `review`
- **THEN** 首页不展示积极文案

#### Scenario: 关系复杂或需留意
- **WHEN** 同日个人结果为 `mixed`、`caution` 或单边忌神的 `undetermined`
- **THEN** 首页展示与状态相符的平稳或谨慎短句，不输出「直接冲」或确定的好坏运断言

#### Scenario: 未登录、未绑定或依据不足
- **WHEN** 没有当前账号绑定的有效同日个人结果，四项主题不完整，或结果为 `insufficient`／未知状态
- **THEN** 首页不展示个人鼓励语，原通用日期与宜忌仍可独立显示

#### Scenario: 日期或账号变化
- **WHEN** 北京时间跨日、账号切换或旧请求晚于新请求返回
- **THEN** 旧日期或旧账号的鼓励语不在首页显示

#### Scenario: 窄屏或较长称呼
- **WHEN** 页面在窄屏显示个人鼓励语
- **THEN** 文案及依据入口保持可读、可点，不横向溢出

### Requirement: 单边喜用使用轻正向文案
首页个人鼓励语 SHALL 区分双喜用、单边喜用且另一边中性、完全中性及含忌神的日期。单边喜用且另一边中性、四项可见主题中至少一项为 `candidate` 且没有 `review` 时 SHALL 显示轻正向语气；四项主题全部为 `neutral` 时 SHALL 显示平稳语气。文案 MUST NOT 声称确定好运或保证具体结果。

#### Scenario: 日干喜用、日支主气中性且有明确主题
- **WHEN** 有效个人结果为 `undetermined`，日干方向为 `favorable`、日支主气为 `neutral`，至少一项主题为 `candidate` 且无 `review`
- **THEN** 首页显示轻正向鼓励语，语气弱于双喜用结果

#### Scenario: 日支主气喜用、日干中性且有明确主题
- **WHEN** 有效个人结果为 `undetermined`，日支主气方向为 `favorable`、日干为 `neutral`，至少一项主题为 `candidate` 且无 `review`
- **THEN** 首页显示同等级的轻正向鼓励语

#### Scenario: 含忌神或完全中性
- **WHEN** 有效个人结果含 `adverse` 方向，或日干与日支主气均为 `neutral`
- **THEN** 首页不因单边喜用规则显示积极语句
