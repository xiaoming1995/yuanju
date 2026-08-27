## ADDED Requirements

### Requirement: AI report receives deterministic vehicle-road context
The backend SHALL provide deterministic vehicle profile and Dayun road context to AI report generation when those fields are available.

#### Scenario: Report generation has vehicle-road data
- **WHEN** the backend builds an AI report prompt for a Bazi result containing `vehicle_profile` and `dayun_roadmap`
- **THEN** the prompt includes the vehicle grade, vehicle type, vehicle summary, current Dayun road type, road phase labels, and major evidence
- **AND** the prompt identifies the context as algorithm-calculated data

#### Scenario: Report generation lacks vehicle-road data
- **WHEN** the backend builds an AI report prompt for a chart that does not contain `vehicle_profile` or `dayun_roadmap`
- **THEN** report generation continues without vehicle-road context
- **AND** the prompt does not ask the AI to invent missing vehicle grades or road labels

### Requirement: AI report explains without overriding
AI-generated Bazi reports SHALL explain deterministic vehicle-road labels without recalculating or overriding them.

#### Scenario: AI responds to vehicle-road context
- **WHEN** the AI report receives vehicle-road context
- **THEN** the report may translate the labels into ordinary-user language
- **AND** the report does not output a conflicting vehicle grade, road type, or score
- **AND** the report avoids deterministic outcome claims
