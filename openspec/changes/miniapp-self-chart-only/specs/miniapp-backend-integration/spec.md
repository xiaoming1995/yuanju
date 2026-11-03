## MODIFIED Requirements

### Requirement: Real chart calculation
The miniapp SHALL collect birth inputs only in the authenticated personal-chart configuration flow, SHALL save the calculated chart as the current account's unique personal chart, and SHALL display a full chart result only after the backend confirms that binding. The miniapp MUST NOT offer arbitrary birth-chart queries on Home.

#### Scenario: Successful first configuration
- **WHEN** an authenticated user without a personal chart saves valid birth date, hour and gender in Mine
- **THEN** the backend SHALL create the chart and its personal binding together, and the miniapp SHALL offer its full detail

#### Scenario: Failed configuration
- **WHEN** calculation or saving the binding fails
- **THEN** the miniapp SHALL show an error/retry state, SHALL NOT display a fixed sample as the user's result, and the previous binding SHALL remain unchanged

#### Scenario: Home without query form
- **WHEN** any user opens the miniapp Home tab
- **THEN** the birth-chart form, its heading, history shortcut and date picker SHALL be absent, while the calendar content SHALL remain available

### Requirement: Real account and history
The miniapp SHALL use ordinary user authentication for profile and personal-chart operations. Existing history data SHALL remain stored, but the miniapp MUST NOT provide full detail for a record other than the current account's bound personal chart.

#### Scenario: First WeChat login
- **WHEN** a user taps WeChat login for the first time while miniapp registration is enabled
- **THEN** the miniapp SHALL enter a new account without requiring web credentials
- **THEN** the miniapp SHALL offer optional avatar and nickname selection using the platform's user-initiated controls

#### Scenario: Optional web account binding
- **WHEN** a signed-in miniapp user chooses web account binding in Mine
- **THEN** the miniapp SHALL collect web credentials, explain that records will merge, and use the returned web-user JWT after a successful merge

#### Scenario: Authenticated history
- **WHEN** a logged-in user opens history
- **THEN** the miniapp SHALL show server-owned record summaries and counts, including empty and paginated states, but SHALL offer full detail only for the bound personal chart

#### Scenario: Mine shows only the personal chart
- **WHEN** a logged-in user opens Mine
- **THEN** the miniapp SHALL show the bound personal chart or its configuration state, and SHALL NOT show a separate chart-history list or count card

#### Scenario: Delete chart
- **WHEN** the user confirms deletion of an owned chart
- **THEN** the backend SHALL delete it and the miniapp history summary SHALL refresh; if it was the bound chart, the miniapp SHALL show the unbound configuration state

#### Scenario: Token expired
- **WHEN** a protected request returns 401
- **THEN** the miniapp SHALL clear its session and offer login without showing another user's cached data

## ADDED Requirements

### Requirement: Bound personal-chart detail
The miniapp SHALL load full chart detail through an authenticated server operation that resolves the current user's personal-chart binding and confirms chart ownership. A client-supplied history ID or cached result MUST NOT bypass this check.

#### Scenario: Current binding
- **WHEN** a signed-in user opens chart detail with a valid owned personal-chart binding
- **THEN** the result SHALL contain only the currently bound chart

#### Scenario: Unbound, deleted or foreign chart
- **WHEN** the user has no valid binding, or attempts to open a different chart by an old miniapp navigation path
- **THEN** the full detail SHALL not display, and the miniapp SHALL guide the user to configure or select their personal chart
