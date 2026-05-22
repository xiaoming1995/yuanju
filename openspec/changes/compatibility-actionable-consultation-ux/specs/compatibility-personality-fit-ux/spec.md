## ADDED Requirements

### Requirement: Compatibility result leads with personality fit
The compatibility result page SHALL present a personality-fit summary before generic scores, timing windows, and professional evidence.

#### Scenario: User opens a compatibility result
- **WHEN** a user opens a compatibility result detail
- **THEN** the first result path includes an overall personality-fit judgment
- **AND** it shows whether the relationship is naturally smooth, mutually attractive but high-friction, slow-burn, repeatedly pulling, or reality-pressured

#### Scenario: AI report is absent
- **WHEN** the result has no AI compatibility report
- **THEN** the personality-fit section still renders from deterministic scores, evidence, and consulting assessment data
- **AND** it labels generated AI report as a deeper explanation rather than a prerequisite for personality-fit reading

#### Scenario: AI report exists
- **WHEN** the result has structured AI report data
- **THEN** the personality-fit section can enrich the summary with report diagnosis and advice while preserving deterministic fallbacks

### Requirement: Both relationship personality patterns are shown
The compatibility result page SHALL explain each participant's relationship personality pattern in practical interaction language.

#### Scenario: Participant summaries render
- **WHEN** participant chart snapshots or relationship evidence are available
- **THEN** the page shows a concise self pattern and partner pattern
- **AND** the copy focuses on needs, pressure response, communication rhythm, and relationship behavior rather than clinical personality labels

#### Scenario: Participant data is incomplete
- **WHEN** one participant lacks enough chart or evidence data
- **THEN** the page uses neutral fallback copy and does not hide the whole personality-fit section

### Requirement: Fit and clash points are explicit
The compatibility result page SHALL distinguish natural fit points from repeated conflict points.

#### Scenario: Fit points render
- **WHEN** positive or supportive evidence exists
- **THEN** the page shows what feels easy, attractive, complementary, or emotionally reinforcing between the two people

#### Scenario: Clash points render
- **WHEN** negative, mixed, or pressure evidence exists
- **THEN** the page shows where communication, pacing, emotional needs, realism, or conflict repair may repeatedly create friction

### Requirement: Communication guidance follows personality fit
The compatibility result page SHALL provide communication guidance that follows from the personality-fit judgment.

#### Scenario: Guidance is rendered
- **WHEN** the personality-fit section is shown
- **THEN** it includes practical guidance for how to speak, slow down, set boundaries, or advance the relationship
- **AND** the guidance is conditional and avoids deterministic fate claims

### Requirement: Personality fit remains evidence-linked
The personality-fit section SHALL provide a route to supporting scores, stage risks, or professional evidence when available.

#### Scenario: User inspects a personality claim
- **WHEN** a personality-fit claim is based on compatibility evidence or score dimensions
- **THEN** the page exposes an affordance to view the relevant score, stage-risk, or evidence section
