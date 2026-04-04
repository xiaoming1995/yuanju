## ADDED Requirements

### Requirement: Calculate Day Master's Life Stages (Xingyun)
The backend must compute the 12 Life Stages of the Day Master against the Earthly Branch of each pillar (Year, Month, Day, Hour).

#### Scenario: User requests Bazi calculation
- **WHEN** the `Calculate` API is invoked
- **THEN** it extracts the numeric value or string name of the life stage (Xingyun) using the Day Master and each branch, ensuring `xing_yun` is set correctly for each pillar.

### Requirement: Calculate Xun Kong (Void Branches)
The backend must evaluate the Xun Kong based on the calculated Bazi context and assign the proper void branch(es) to each pillar.

#### Scenario: Chart generation implies voidness
- **WHEN** the `bazi.GetXunKong()` methods return valid void branches
- **THEN** it maps them into string variables `year_xun_kong`, `month_xun_kong`, `day_xun_kong`, and `hour_xun_kong`.

### Requirement: Extract Shensha (Gods and Demons)
The backend must gather all divine stars (Shensha) attached to each of the four branches utilizing lunar-go logic.

#### Scenario: Bazi properties are serialized
- **WHEN** lunar-go returns Shensha arrays/maps
- **THEN** the system flattens them into string arrays: `year_shen_sha`, `month_shen_sha`, `day_shen_sha`, and `hour_shen_sha`.
