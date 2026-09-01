## ADDED Requirements

### Requirement: Dayun summary uses authoritative natal-strength context
The Dayun timeline SHALL receive the result's natal Fuyi strength context and SHALL not independently infer a Day Master strength from five-element distribution data.

#### Scenario: User switches between Dayun cards
- **WHEN** a user selects different Dayun cards for the same natal chart
- **THEN** the summary's natal-strength context remains unchanged while the stem and branch interpretations update for the selected Dayun.

#### Scenario: Dayun branch carries an adverse element at 长生
- **WHEN** the selected Dayun branch is natal-Fuyi adverse and at 长生
- **THEN** the ordinary summary SHALL communicate that the adverse branch influence is more apparent without describing the natal chart as body-weak.
