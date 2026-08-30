## ADDED Requirements

### Requirement: Dayun road evidence preserves opposing phase contributions
The system SHALL expose phase-specific evidence for each newly calculated Dayun road without replacing its existing aggregate road score or aggregate evidence.

#### Scenario: Helpful stem and adverse branch offset across one Dayun
- **WHEN** a Dayun stem contributes a positive five-element or Ten God delta and its branch contributes a negative delta for the same signal
- **THEN** the result SHALL expose the positive contribution in a `front` phase evidence group and the negative contribution in a `back` phase evidence group
- **AND** the aggregate evidence SHALL retain their summed delta, including `0` when they offset
- **AND** the road score SHALL continue to use that aggregate delta.

#### Scenario: Phase evidence is additive for existing consumers
- **WHEN** a Dayun road is returned by the calculate API
- **THEN** it SHALL retain `score`, `road_type`, `qian_road`, `hou_road`, and `evidences`
- **AND** it SHALL add structured `phase_evidences` for phase-aware consumers
- **AND** a consumer without `phase_evidences` support SHALL still be able to render the aggregate evidence.

### Requirement: Phase attribution follows Dayun timing semantics
The system SHALL attribute stem-led and branch-led signals to the same front-five and back-five timing model used by the existing Dayun interpretation.

#### Scenario: Stem-led signal is assigned to the front phase
- **WHEN** Dayun five-element, Ten God, or verified stem-led pattern evidence is computed from the Dayun stem
- **THEN** that evidence SHALL be included in the `front` phase group
- **AND** the group SHALL identify it as the front-five, stem-led phase.

#### Scenario: Branch-led signal is assigned to the back phase
- **WHEN** Dayun five-element, Ten God, 十二长生, or branch-derived modifier evidence is computed from the Dayun branch
- **THEN** that evidence SHALL be included in the `back` phase group
- **AND** the group SHALL identify it as the back-five, branch-led phase.

### Requirement: Ten God confidence applies to each phase contribution
The system SHALL apply the established Ten God confidence policy to each phase contribution before calculating the aggregate Ten God delta.

#### Scenario: Medium or hard confidence retains opposing contributions
- **WHEN** the Dayun stem and branch have opposing favorable and adverse Ten God effects under medium or hard confidence
- **THEN** the system SHALL retain both weighted phase deltas
- **AND** it SHALL calculate the aggregate from the weighted phase deltas rather than cancelling the raw effects first.

#### Scenario: Soft confidence remains non-scoring
- **WHEN** the natal Ten God confidence is soft
- **THEN** no phase Ten God contribution SHALL change the Dayun road score
- **AND** the evidence SHALL state that Ten God confidence is too soft for a strong correction.

### Requirement: Verified 偏印格杀印相生 affects the front phase only
The system SHALL score a Dayun 七杀 stem as a 偏印格的杀印相生 contribution only when the natal chain is explicitly verifiable.

#### Scenario: Visible rooted 偏印 receives a Dayun 七杀 stem
- **WHEN** the natal primary pattern is 偏印格
- **AND** natal 偏印 is visible and rooted
- **AND** the Dayun stem is 七杀 and its element generates the natal 偏印 element
- **AND** the interaction has not already been counted by another formation rule
- **THEN** the system SHALL add a positive `杀印相生` pattern contribution to the `front` phase evidence
- **AND** the aggregate pattern evidence SHALL include that same contribution exactly once.

#### Scenario: Incomplete 偏印 chain stays neutral
- **WHEN** the natal 偏印 is hidden-only, the Dayun 七杀 appears only in the branch, the elemental relay does not hold, or the primary pattern is not 偏印格
- **THEN** the system SHALL NOT add a 杀印相生 score
- **AND** the pattern evidence SHALL explain the non-match without mislabeling it as a formed interaction.

### Requirement: Professional road evidence exposes phase and aggregate views
The result page SHALL render phase-aware Dayun evidence in professional mode while preserving a fallback for earlier result data.

#### Scenario: Professional user opens current-road evidence
- **WHEN** a current Dayun road includes `phase_evidences` and the user opens its evidence modal in professional mode
- **THEN** the modal SHALL show an aggregate summary and separate front-five and back-five evidence groups
- **AND** each group SHALL render the source, label, signed delta, and deterministic explanation.

#### Scenario: Earlier result has aggregate evidence only
- **WHEN** a current Dayun road lacks `phase_evidences`
- **THEN** the professional modal SHALL render the existing aggregate evidence list without error.
