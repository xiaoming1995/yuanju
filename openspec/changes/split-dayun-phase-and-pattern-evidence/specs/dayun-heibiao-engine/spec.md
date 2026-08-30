## ADDED Requirements

### Requirement: Dayun evidence reuses the established front/back boundary
The system SHALL use the existing 金不换 timing boundary when assigning Dayun road evidence to phases.

#### Scenario: Road evidence is grouped with 金不换 phases
- **WHEN** a Dayun has front-five and back-five 金不换 results and phase-aware road evidence
- **THEN** stem-led road evidence SHALL be grouped with the same front-five phase as `qian_road`
- **AND** branch-led road evidence SHALL be grouped with the same back-five phase as `hou_road`
- **AND** the new grouping SHALL NOT alter the existing 金不换 rating or description fields.
