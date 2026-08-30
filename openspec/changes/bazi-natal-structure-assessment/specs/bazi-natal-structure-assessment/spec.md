## ADDED Requirements

### Requirement: Versioned natal structure assessment
The backend SHALL attach a versioned natal structure assessment to every newly calculated Bazi result. The assessment MUST contain climate urgency, Fuyi use-god support, pattern quality, formation or break evidence, natal flow or relation evidence, carrying state, score, grade, and human-readable evidence records.

#### Scenario: A new chart is calculated
- **WHEN** the Bazi calculation engine completes a chart
- **THEN** the result contains a current-version `natal_assessment` with evidence for every non-neutral grade adjustment.

#### Scenario: Assessment has insufficient source data
- **WHEN** a saved or constructed chart lacks inputs needed to confirm a pattern condition
- **THEN** the assessment marks that condition as partial or undecided and MUST NOT award it as a formed pattern.

### Requirement: Climate urgency and Fuyi priority
The assessment SHALL treat a confirmed extreme cold or extreme heat condition as a climate urgency gate. When no climate urgency exists, Fuyi support SHALL be the base determination for natal carrying; a general Tiaohou dictionary match MUST NOT replace the Fuyi conclusion for grade calculation.

#### Scenario: Urgent climate condition is unresolved
- **WHEN** a chart has a confirmed climate urgency and its required element is absent from both visible and hidden natal stems
- **THEN** the assessment records an unresolved urgency and caps the natal grade at the configured conservative ceiling.

#### Scenario: Climate is not urgent
- **WHEN** a chart does not meet the confirmed extreme cold or hot condition
- **THEN** the assessment records climate as non-urgent and determines its base carrying from Fuyi use-god support.

### Requirement: Pattern formation, break, and flow affect natal grade
The assessment SHALL not award a fixed score solely because a pattern name exists. It MUST evaluate supported pattern stars, applicable formation or transformation combinations, applicable break conditions, and natal five-element flow before producing the grade.

#### Scenario: Pattern has an effective transformation
- **WHEN** the detected pattern has a configured formation or transformation combination supported by natal ten-god evidence
- **THEN** the assessment records the named combination as positive pattern evidence and includes it in the grade calculation.

#### Scenario: Pattern has an unremedied break condition
- **WHEN** the detected pattern has a configured break condition without its configured remedy
- **THEN** the assessment records the break evidence, reduces the structure score, and applies the configured grade ceiling when the break is critical.

#### Scenario: Natal flow is incomplete
- **WHEN** the natal five elements do not form a configured continuous productive path
- **THEN** the assessment records incomplete or blocked flow as a conservative relation adjustment rather than treating every branch combination as favorable.

### Requirement: Assessment evidence is consumable by presentation layers
The assessment SHALL expose stable source, label, impact, score delta, and detail fields so the result page and AI prompt can explain why a grade was assigned without reconstructing rules on the client.

#### Scenario: Presentation reads an assessment
- **WHEN** the result page or report service consumes a current assessment
- **THEN** it can render the grade and its supporting or limiting evidence directly from the response data.
