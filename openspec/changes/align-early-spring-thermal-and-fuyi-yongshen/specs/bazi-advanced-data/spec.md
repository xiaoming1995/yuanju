## ADDED Requirements

### Requirement: Bazi result exposes shared-priority yongshen alignment
The backend Bazi result SHALL expose structured natal shared-priority yongshen alignment data, including the aligned elements, source layers, and an explanatory detail, while preserving existing thermal and 扶抑 fields.

#### Scenario: Calculation yields shared fire priority
- **WHEN** a calculated natal assessment has `火` selected by both thermal regulation and 扶抑
- **THEN** the API response SHALL include `火` in the shared-priority alignment with both source layers

#### Scenario: Existing consumers read separate yongshen fields
- **WHEN** an API consumer reads thermal or 扶抑 yongshen data
- **THEN** those existing fields SHALL retain their names, meanings, and complete element lists
