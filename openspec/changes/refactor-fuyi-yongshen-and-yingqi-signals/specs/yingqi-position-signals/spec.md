## ADDED Requirements

### Requirement: 用神/忌神覆盖位置收集
系统 SHALL 收集原局所有"含用神/忌神五行"的干支位置，包括：天干五行属用神/忌神、地支本身五行属用神/忌神、地支藏干（主气+中气）含用神/忌神五行的地支。

#### Scenario: 天干直接属用神五行
- **WHEN** 用神为木，年干为甲
- **THEN** 年干甲纳入用神位集合

#### Scenario: 地支本身五行属用神
- **WHEN** 用神为水，日支为子
- **THEN** 日支子纳入用神位集合

#### Scenario: 地支藏干含用神五行
- **WHEN** 用神为木，年支藏干中气含甲（如亥支藏甲）
- **THEN** 年支亥纳入用神位集合

#### Scenario: 余气不纳入覆盖位置
- **WHEN** 用神为火，某支藏干余气为丁，但主气中气均非火
- **THEN** 该地支不纳入用神位集合

### Requirement: 刑冲克穿对用神/忌神位产生独立信号
系统 SHALL 检测流年/大运干支对原局用神/忌神位的刑（六刑）、冲（六冲）、克（天干五行相克）、穿（六害）交互，每条交互独立输出 EventSignal，不影响其他信号极性。

#### Scenario: 流年地支冲用神地支位 → 凶信号
- **WHEN** 流年地支与原局某用神位地支构成六冲
- **THEN** 输出 EventSignal{Type:"应期_用神", Polarity:"凶", Evidence:"流年XX冲原局用神位YY"}

#### Scenario: 流年天干克用神天干位 → 凶信号
- **WHEN** 流年天干五行克原局某用神位天干五行
- **THEN** 输出 EventSignal{Type:"应期_用神", Polarity:"凶", Evidence:"流年XX克原局用神位YY"}

#### Scenario: 流年地支刑忌神地支位 → 吉信号
- **WHEN** 流年地支与原局某忌神位地支构成六刑
- **THEN** 输出 EventSignal{Type:"应期_忌神", Polarity:"吉", Evidence:"流年XX刑原局忌神位YY"}

#### Scenario: 流年地支穿（六害）忌神地支位 → 吉信号
- **WHEN** 流年地支与原局某忌神位地支构成六害
- **THEN** 输出 EventSignal{Type:"应期_忌神", Polarity:"吉", Evidence:"流年XX穿原局忌神位YY"}

### Requirement: 合对用神/忌神位按化出五行定吉凶
系统 SHALL 对天干五合和地支六合分别检测，按化出五行（或合入方五行）属用神/忌神定信号极性。合而不化时按"锁定"处理（用神被锁=凶，忌神被锁=吉），evidence 注明"合而不化"。

#### Scenario: 流年天干合用神天干，化出五行属忌神 → 凶
- **WHEN** 流年天干与原局用神位天干构成天干五合，化出五行属忌神
- **THEN** 输出 EventSignal{Type:"应期_用神", Polarity:"凶"}

#### Scenario: 流年天干合用神天干，化出五行属用神 → 吉
- **WHEN** 流年天干与原局用神位天干构成天干五合，化出五行属用神
- **THEN** 输出 EventSignal{Type:"应期_用神", Polarity:"吉"}

#### Scenario: 合而不化锁定用神 → 凶
- **WHEN** 流年地支与原局用神位地支六合，但无根气不化
- **THEN** 输出 EventSignal{Type:"应期_用神", Polarity:"凶", Evidence:"...合而不化，用神被锁"}

### Requirement: 大运与原局交互使用同一套规则
系统 SHALL 对大运干支使用与流年相同的刑冲克合穿规则检测原局用神/忌神位，输出独立信号。

#### Scenario: 大运天干克原局忌神天干 → 吉信号
- **WHEN** 大运天干五行克原局某忌神位天干五行
- **THEN** 输出 EventSignal{Type:"应期_忌神", Polarity:"吉", Source:"大运应期"}

### Requirement: 应期信号独立输出，不影响其他信号极性
系统 SHALL 不再输出全年底色（baseline），应期信号的 Polarity 独立设定，不作为 applyPolarity 的 baseline 参数传入其他信号。

#### Scenario: 应期凶信号不染黑其他吉信号
- **WHEN** 流年刑用神位产生凶信号，同年印星透干产生吉信号
- **THEN** 两条信号各自保持独立极性，印星透干仍输出吉
