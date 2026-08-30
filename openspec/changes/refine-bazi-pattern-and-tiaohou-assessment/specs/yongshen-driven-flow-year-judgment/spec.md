## ADDED Requirements

### Requirement: Flow and AI contexts preserve Tiaohou dimension semantics
The system SHALL provide day-stem Tiaohou and thermal Tiaohou context separately to flow-year and AI consumers. It MUST NOT represent day-stem Tiaohou availability as a thermal urgency result, or replace global Fuyi favorable elements with either Tiaohou result.

#### Scenario: Day-stem Tiaohou is available while thermal urgency is absent
- **WHEN** a chart has an available day-stem Tiaohou stem and no unresolved thermal urgency
- **THEN** the generated context SHALL identify the day-stem support and the non-urgent thermal result separately
- **AND THEN** it SHALL retain the global Fuyi favorable elements as the Fuyi conclusion
