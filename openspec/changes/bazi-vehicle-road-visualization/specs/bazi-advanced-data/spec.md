## ADDED Requirements

### Requirement: Bazi result exposes vehicle and road structures
The backend Bazi result SHALL expose structured vehicle and road data while preserving existing advanced Bazi fields.

#### Scenario: New chart calculation returns advanced profile data
- **WHEN** the calculation engine returns a `BaziResult`
- **THEN** the result includes the existing advanced fields such as Ming Ge, Tiaohou, Ten God relation, Shen Sha, Dayun, and favorable/adverse Ten Gods
- **AND** the result includes `vehicle_profile` and `dayun_roadmap` when enough deterministic inputs are available

#### Scenario: Existing fields remain compatible
- **WHEN** API consumers read existing Bazi fields
- **THEN** existing field names, meanings, and JSON shapes are not removed or renamed by this change

#### Scenario: Old saved chart lacks profile fields
- **WHEN** a saved chart snapshot does not contain `vehicle_profile` or `dayun_roadmap`
- **THEN** services and frontend consumers continue to render the saved chart without error
- **AND** any fallback profile display is explicitly limited to data that can be derived from the saved snapshot
