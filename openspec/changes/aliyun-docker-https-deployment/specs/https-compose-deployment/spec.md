## ADDED Requirements

### Requirement: Environment-specific Compose entry points
The project SHALL provide a common Compose configuration and explicit local and production overlays. The local entry point SHALL publish frontend and backend development ports; the production entry point SHALL publish only HTTP and HTTPS ingress ports.

#### Scenario: Local startup
- **WHEN** the local Compose configuration is selected
- **THEN** frontend port 80 and backend port 9002 are published to the host

#### Scenario: Production startup
- **WHEN** the production Compose configuration is selected
- **THEN** only Caddy ports 80 and 443 are published to the host

### Requirement: HTTPS ingress
The production stack SHALL proxy only the configured web hostname to the frontend service through Caddy, with persistent Caddy certificate data. API requests on the web hostname SHALL reach the backend through the frontend Nginx proxy. The production stack SHALL NOT expose a separate API hostname route.

#### Scenario: Domain routing
- **WHEN** a valid web domain is configured and traffic reaches Caddy
- **THEN** the web hostname is routed to the frontend service and its `/api` paths are proxied to the backend

#### Scenario: Separate API hostname
- **WHEN** a request targets an unrelated `api` hostname
- **THEN** the production Caddy configuration does not route it to the backend

#### Scenario: Missing domains
- **WHEN** a required production domain variable is absent
- **THEN** Compose configuration fails before startup

### Requirement: Reachable certificate issuer
The production Caddy configuration SHALL obtain publicly trusted certificates from the configured ZeroSSL ACME service using a required contact email.

#### Scenario: Missing certificate contact
- **WHEN** the production ACME email variable is absent
- **THEN** Compose configuration fails before startup

#### Scenario: Certificate issuance
- **WHEN** the ZeroSSL services and domain validation are reachable from the server
- **THEN** Caddy can request and renew the web hostname certificate without manual certificate files

### Requirement: Secret-safe backend build context
The backend Docker build context MUST exclude local environment files containing runtime credentials.

#### Scenario: Building backend image
- **WHEN** a backend image is built from the repository directory
- **THEN** `.env` and `.env.*` files are excluded from its build context
