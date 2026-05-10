## ADDED Requirements

### Requirement: Layer 0 凶信号压制 Layer 4 吉信号
当同一流年存在 Layer 0（应期位置信号）极性为凶的信号时，系统 SHALL 移除该年 Layer 4（十神宏观，Source=柱位互动）中极性为吉的信号，避免矛盾输出。

#### Scenario: Layer 0 凶压制 Layer 4 吉
- **WHEN** 流年 collectYingqiSignals 返回至少一条凶信号，且 Layer 4 生成了吉信号（Source=柱位互动，Polarity=吉）
- **THEN** 该 Layer 4 吉信号从最终 signals 列表中移除，不输出给调用方

### Requirement: Layer 0 吉信号不压制 Layer 4 凶信号
当 Layer 0 存在吉信号时，系统 SHALL 保留 Layer 4 的凶信号，凶的提醒不因吉信号存在而消除。

#### Scenario: Layer 0 吉不压制 Layer 4 凶
- **WHEN** 流年 collectYingqiSignals 返回吉信号，且 Layer 4 生成了凶信号
- **THEN** Layer 4 凶信号正常保留并输出

### Requirement: 同向信号叠加输出
当 Layer 0 与 Layer 4 极性相同时，系统 SHALL 两条信号都输出，不去重，表示该年该领域力度加强。

#### Scenario: 同向叠加
- **WHEN** Layer 0 和 Layer 4 均产生凶信号（或均产生吉信号）
- **THEN** 两条信号均出现在最终输出中

### Requirement: Layer 0 无信号时 Layer 4 正常输出
当该流年 collectYingqiSignals 返回空列表时，系统 SHALL 将 Layer 4 信号作为主要参考正常输出。

#### Scenario: Layer 0 无信号
- **WHEN** collectYingqiSignals 对该流年返回空列表
- **THEN** Layer 4 所有信号不受压制，正常输出
