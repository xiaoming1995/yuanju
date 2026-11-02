## ADDED Requirements

### Requirement: Environment-specific Compose entry points
The project SHALL provide a common Compose configuration and explicit local and production overlays. The local entry point SHALL publish frontend and backend development ports; the production entry point SHALL publish only HTTP and HTTPS ingress ports.

#### Scenario: Local startup
- **WHEN** the local Compose configuration is selected
- **THEN** frontend port 3000 and backend port 9002 are published to the host

#### Scenario: Production startup
- **WHEN** the production Compose configuration is selected
- **THEN** only Caddy ports 80 and 443 are published to the host

### Requirement: HTTPS ingress
The production stack SHALL proxy the web domain to the frontend service and the API domain to the backend service through Caddy, with persistent Caddy certificate data.

#### Scenario: Domain routing
- **WHEN** valid web and API domains are configured and traffic reaches Caddy
- **THEN** each domain is routed to its corresponding service over the Compose network

#### Scenario: Missing domains
- **WHEN** a required production domain variable is absent
- **THEN** Compose configuration fails before startup

### Requirement: Secret-safe backend build context
The backend Docker build context MUST exclude local environment files containing runtime credentials.

#### Scenario: Building backend image
- **WHEN** a backend image is built from the repository directory
- **THEN** `.env` and `.env.*` files are excluded from its build context
