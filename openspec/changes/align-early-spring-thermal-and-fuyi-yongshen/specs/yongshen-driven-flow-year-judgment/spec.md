## ADDED Requirements

### Requirement: AI interpretation context distinguishes shared-priority yongshen
The system SHALL include shared-priority yongshen alignment in natal report and dayun interpretation context when alignment elements exist. The prompt SHALL identify the element as jointly supported by cold/heat thermal regulation and 扶抑, without presenting it as the sole yongshen.

#### Scenario: Shared fire priority is available
- **WHEN** the natal assessment aligns `火` across thermal regulation and 扶抑
- **THEN** generated report and dayun prompt context SHALL state that `火` is a shared-priority yongshen from both layers

#### Scenario: No shared priority is available
- **WHEN** the natal assessment has no aligned element
- **THEN** generated prompt context SHALL omit the shared-priority statement and retain the existing separate yongshen context
