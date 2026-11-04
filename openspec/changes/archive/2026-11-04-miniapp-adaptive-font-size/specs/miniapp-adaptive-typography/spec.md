## ADDED Requirements

### Requirement: Readable default typography
The miniapp SHALL present core body text, labels, controls, and explanatory text at a more readable default size across its pages while preserving visual hierarchy.

#### Scenario: Standard size
- **WHEN** a user selects Standard font size
- **THEN** all miniapp pages display the increased baseline text sizes

### Requirement: Follow WeChat font size
The miniapp SHALL use the user's available WeChat font size setting by default, refresh it when pages become visible, and fall back to Standard when the setting is unavailable.

#### Scenario: WeChat font size changes
- **WHEN** the user changes WeChat's font size and returns to the miniapp in Follow WeChat mode
- **THEN** the miniapp updates its text size according to the new setting

#### Scenario: Setting unavailable
- **WHEN** the WeChat font size value cannot be obtained
- **THEN** the miniapp uses Standard size without blocking the page

### Requirement: Manual font size choice
The miniapp SHALL offer Follow WeChat, Standard, Larger, and Extra Large in the Mine page, apply the choice immediately across pages, and persist the mode on the current device independent of login state.

#### Scenario: Choose a fixed size
- **WHEN** a user selects Larger or Extra Large in Mine
- **THEN** the chosen size is shown immediately and remains after restarting the miniapp

#### Scenario: Resume following WeChat
- **WHEN** a user changes the setting back to Follow WeChat
- **THEN** the miniapp uses the current WeChat font size again

### Requirement: Large text remains usable
The miniapp SHALL keep essential labels, values, and actions readable and operable at Extra Large size on narrow mobile screens.

#### Scenario: Dense pages
- **WHEN** Extra Large size is active on a narrow screen
- **THEN** calendar dates, chart values, settings, and primary actions do not overlap or clip essential text
