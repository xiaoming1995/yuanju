## ADDED Requirements

### Requirement: Structured ten-god relation data
The backend bazi result SHALL provide structured ten-god relation data derived from the day master while preserving existing raw ten-god fields.

#### Scenario: New chart calculation returns advanced bazi data
- **WHEN** the calculation engine returns `BaziResult`
- **THEN** the result includes structured day-master, heavenly-stem relation, and hidden-stem relation data in addition to existing raw fields such as `*_gan_shishen`, `*_hide_gan`, and `*_zhi_shishen`

#### Scenario: Existing raw ten-god fields remain compatible
- **WHEN** API consumers read existing ten-god fields
- **THEN** those fields keep their current names and meanings and are not removed or renamed

#### Scenario: Old saved chart is loaded
- **WHEN** a saved chart snapshot does not contain the structured relation matrix
- **THEN** the service can derive equivalent relation data from existing raw chart fields or recalculate the chart snapshot using existing lazy-load behavior
