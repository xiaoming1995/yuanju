## ADDED Requirements

### Requirement: Dayun Phase Classification

The past-events yearly inference SHALL classify each liunian by its position inside the containing 10-year dayun.

#### Scenario: First half of dayun uses gan phase

- **GIVEN** a dayun with 10 liunian items
- **WHEN** the system evaluates liunian positions 1 through 5
- **THEN** each evaluated year is classified as dayun phase `gan`

#### Scenario: Second half of dayun uses zhi phase

- **GIVEN** a dayun with 10 liunian items
- **WHEN** the system evaluates liunian positions 6 through 10
- **THEN** each evaluated year is classified as dayun phase `zhi`

### Requirement: Phase-Aware Yearly Signals

The past-events yearly signal engine SHALL use dayun phase as weighting context while preserving direct liunian triggers.

#### Scenario: Gan phase emphasizes dayun heavenly stem background

- **GIVEN** a year in positions 1 through 5 of a dayun
- **WHEN** yearly signals are generated
- **THEN** dayun heavenly-stem context is eligible for stronger background emphasis
- **AND** dayun earthly-branch context remains available but is not treated as the phase-leading background

#### Scenario: Zhi phase emphasizes dayun earthly branch background

- **GIVEN** a year in positions 6 through 10 of a dayun
- **WHEN** yearly signals are generated
- **THEN** dayun earthly-branch context is eligible for stronger background emphasis
- **AND** dayun heavenly-stem context remains available but is not treated as the phase-leading background

#### Scenario: Direct yearly events are preserved

- **GIVEN** a liunian with a direct health, school, relationship, movement, fuyin, fanyin, or pillar-interaction trigger
- **WHEN** dayun phase weighting is applied
- **THEN** the direct yearly trigger remains present in the returned signals
- **AND** phase weighting does not remove or overwrite it

### Requirement: JinBuHuan Phase Context

The past-events yearly inference SHALL reuse the existing JinBuHuan front/back five-year result as background context.

#### Scenario: Front five years use qian rating

- **GIVEN** a dayun with a non-empty `jin_bu_huan.qian_level` and `qian_desc`
- **WHEN** a year in positions 1 through 5 is evaluated
- **THEN** the year can include dayun phase background derived from `qian_level` and `qian_desc`
- **AND** `hou_desc` is not used as that year's phase background

#### Scenario: Back five years use hou rating

- **GIVEN** a dayun with a non-empty `jin_bu_huan.hou_level` and `hou_desc`
- **WHEN** a year in positions 6 through 10 is evaluated
- **THEN** the year can include dayun phase background derived from `hou_level` and `hou_desc`
- **AND** `qian_desc` is not used as that year's phase background

### Requirement: Plain-Language Narrative Priority

The yearly narrative renderer SHALL treat dayun phase context as supporting tone unless no more specific yearly event exists.

#### Scenario: Specific yearly theme remains dominant

- **GIVEN** a year with both dayun phase background and a specific school, health, relationship, career, money, movement, fuyin, or fanyin signal
- **WHEN** the plain-language yearly narrative is rendered
- **THEN** the specific yearly theme is the primary narrative focus
- **AND** the dayun phase context may appear as secondary context or evidence

#### Scenario: Phase context explains otherwise quiet year

- **GIVEN** a year whose only meaningful signal is dayun phase background
- **WHEN** the plain-language yearly narrative is rendered
- **THEN** the narrative describes the early/late dayun tone in non-technical language
- **AND** the professional basis remains available through evidence summary

### Requirement: Dayun Summary Receives Phase Data

The per-dayun AI summary generation SHALL receive phase-aware yearly signal data.

#### Scenario: AI prompt includes phase context

- **GIVEN** the system is generating an AI summary for one dayun
- **WHEN** it serializes the 10 yearly signal rows for the prompt
- **THEN** each row includes enough data to distinguish early `gan` phase years from late `zhi` phase years
- **AND** the prompt instructs the AI to mention early/late phase differences when the signal data supports a meaningful contrast
