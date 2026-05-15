## Purpose

Define advanced bazi calculation fields exposed by the backend beyond the core four pillars, including life stages, void branches, shensha, and structured ten-god relation data.
## Requirements
### Requirement: Calculate Day Master's Life Stages (Xingyun)
The backend MUST compute the 12 Life Stages of the Day Master against the Earthly Branch of each pillar (Year, Month, Day, Hour).

#### Scenario: User requests Bazi calculation
- **WHEN** the `Calculate` API is invoked
- **THEN** it extracts the numeric value or string name of the life stage (Xingyun) using the Day Master and each branch, ensuring `xing_yun` is set correctly for each pillar.

### Requirement: Calculate Xun Kong (Void Branches)
The backend MUST evaluate the Xun Kong based on the calculated Bazi context and assign the proper void branch(es) to each pillar.

#### Scenario: Chart generation implies voidness
- **WHEN** the `bazi.GetXunKong()` methods return valid void branches
- **THEN** it maps them into string variables `year_xun_kong`, `month_xun_kong`, `day_xun_kong`, and `hour_xun_kong`.

### Requirement: Extract Shensha (Gods and Demons)
The backend MUST gather all divine stars (Shensha) attached to each of the four branches utilizing lunar-go logic.

#### Scenario: Bazi properties are serialized
- **WHEN** lunar-go returns Shensha arrays/maps
- **THEN** the system flattens them into string arrays: `year_shen_sha`, `month_shen_sha`, `day_shen_sha`, and `hour_shen_sha`.

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
