## ADDED Requirements

### Requirement: Admin Creates Ordinary Users
The system SHALL allow authenticated administrators to create ordinary user accounts from the Admin backend.

#### Scenario: Admin creates user successfully
- **WHEN** an authenticated administrator submits a unique email, an optional nickname, and an initial password of at least 8 characters
- **THEN** the system SHALL create an ordinary user account
- **THEN** the system SHALL store the password as a bcrypt hash
- **THEN** the system SHALL return the created user's basic information without exposing the password hash

#### Scenario: Admin-created user can login
- **WHEN** a user created by an administrator submits the configured email and initial password to the ordinary user login endpoint
- **THEN** the system SHALL authenticate the user and return an ordinary user JWT

#### Scenario: Admin creates duplicate email
- **WHEN** an authenticated administrator submits an email that already belongs to an ordinary user
- **THEN** the system SHALL reject the request with a conflict error

#### Scenario: Non-admin cannot create user through Admin endpoint
- **WHEN** a request without a valid Admin JWT calls the Admin user creation endpoint
- **THEN** the system SHALL reject the request with an authentication error

### Requirement: Public Registration Setting
The system SHALL persist a site-level setting that controls whether ordinary users may self-register through the public registration endpoint.

#### Scenario: Registration setting defaults enabled
- **WHEN** the system starts after the migration is applied and no explicit registration setting exists
- **THEN** the system SHALL treat public registration as enabled

#### Scenario: Admin reads registration setting
- **WHEN** an authenticated administrator requests the registration setting
- **THEN** the system SHALL return whether public registration is currently enabled

#### Scenario: Admin updates registration setting
- **WHEN** an authenticated administrator updates the registration setting to enabled or disabled
- **THEN** the system SHALL persist the new value
- **THEN** later public registration attempts SHALL use the updated value without requiring a server restart

#### Scenario: Non-admin cannot update registration setting
- **WHEN** a request without a valid Admin JWT attempts to update the registration setting
- **THEN** the system SHALL reject the request with an authentication error

### Requirement: Public Registration Respects Setting
The ordinary user self-registration endpoint SHALL enforce the public registration setting on the backend.

#### Scenario: Public registration enabled
- **WHEN** public registration is enabled and a visitor submits a valid unique email and password to the ordinary registration endpoint
- **THEN** the system SHALL create the account and return an ordinary user JWT

#### Scenario: Public registration disabled
- **WHEN** public registration is disabled and a visitor submits a registration request to the ordinary registration endpoint
- **THEN** the system SHALL reject the request with a forbidden error
- **THEN** the system SHALL NOT create a user account

#### Scenario: Admin user creation remains available when registration disabled
- **WHEN** public registration is disabled and an authenticated administrator creates an ordinary user through the Admin endpoint
- **THEN** the system SHALL create the ordinary user account

### Requirement: Frontend Reflects Registration Availability
The frontend SHALL reflect public registration availability in ordinary user-facing registration entrypoints.

#### Scenario: Registration enabled shows entrypoints
- **WHEN** public registration is enabled
- **THEN** the unauthenticated navigation SHALL show a registration entrypoint
- **THEN** the registration page SHALL show the registration form

#### Scenario: Registration disabled hides ordinary entrypoints
- **WHEN** public registration is disabled
- **THEN** unauthenticated public pages SHALL avoid presenting ordinary registration as an available action
- **THEN** the registration page SHALL show a clear unavailable-state message instead of submitting the registration form

### Requirement: Admin Users Page Supports Registration Operations
The Admin users page SHALL expose controls for creating ordinary users and managing public registration availability.

#### Scenario: Admin sees user creation control
- **WHEN** an authenticated administrator opens the Admin users page
- **THEN** the page SHALL provide a control to create an ordinary user with email, optional nickname, and initial password

#### Scenario: Admin creates user from users page
- **WHEN** an authenticated administrator submits a valid create-user form from the Admin users page
- **THEN** the page SHALL call the Admin user creation endpoint
- **THEN** the page SHALL refresh or update the user list to include the newly created user

#### Scenario: Admin toggles public registration from users page
- **WHEN** an authenticated administrator changes the public registration toggle on the Admin users page
- **THEN** the page SHALL persist the change through the Admin registration setting endpoint
- **THEN** the page SHALL show the saved setting state
