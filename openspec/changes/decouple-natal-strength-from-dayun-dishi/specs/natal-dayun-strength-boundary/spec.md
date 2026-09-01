## ADDED Requirements

### Requirement: Natal Fuyi strength is immutable across Dayun interpretation
The system SHALL calculate Day Master strength exclusively from the complete natal chart and SHALL use the same natal Fuyi conclusion for every Dayun of that chart. A Dayun branch, its 十二长生 stage, or another Dayun signal MUST NOT recalculate or relabel the natal Day Master as strong or weak.

#### Scenario: Natal-strong chart enters a Dayun with 长生 branch
- **WHEN** a natal chart is assessed as strong or very strong and a selected Dayun branch is at 长生
- **THEN** the Dayun summary retains the natal-strong conclusion and MUST NOT state that the chart is weak because of that Dayun.

#### Scenario: Frontend receives no natal assessment
- **WHEN** an older result payload lacks `natal_assessment.fuyi.day_master_strength`
- **THEN** the Dayun summary SHALL use neutral wording and MUST NOT derive a strong/weak conclusion from flat five-element percentages.

### Requirement: Dayun branch stage modifies the branch's Fuyi force
The system SHALL determine the Dayun branch's natal Fuyi polarity before applying 十二长生. The stage SHALL only amplify, moderate, or suppress the branch's own favorable/adverse contribution and SHALL not provide a standalone directional verdict.

#### Scenario: Favorable branch at a vigorous stage
- **WHEN** a Dayun branch is a natal Fuyi favorable element and its 十二长生 is 帝旺、临官、长生, or 冠带
- **THEN** the back-five evidence SHALL record an additional positive branch-strength contribution and identify the favorable branch and stage.

#### Scenario: Adverse branch at a vigorous stage
- **WHEN** a Dayun branch is a natal Fuyi adverse element and its 十二长生 is 帝旺、临官、长生, or 冠带
- **THEN** the back-five evidence SHALL record an additional negative branch-strength contribution and identify the adverse branch and stage.

#### Scenario: Adverse branch at a weak stage
- **WHEN** a Dayun branch is a natal Fuyi adverse element and its 十二长生 is 衰、病、死, or 绝
- **THEN** the stage SHALL reduce or neutralize that branch's adverse intensity and MUST NOT reverse it into a favorable contribution.

#### Scenario: Neutral branch has no directional stage score
- **WHEN** a Dayun branch is neither a natal Fuyi favorable nor adverse element
- **THEN** its 十二长生 evidence SHALL be explanatory with a zero directional contribution.

### Requirement: Dayun summary distinguishes stem and branch roles
The system SHALL render the selected Dayun stem and branch as independent natal-Fuyi effects. The summary MUST describe a favorable/adverse branch-stage interaction without using the branch stage to change the natal strength label.

#### Scenario: 戊申 for a natal-strong 壬日主
- **WHEN** the natal Fuyi assessment is strong, 戊土 is favorable, 申金 is adverse, and 戊申 is selected
- **THEN** the Dayun summary SHALL identify the favorable 戊土 and adverse 申金 roles and MUST NOT contain “身弱遭杀克身”.
