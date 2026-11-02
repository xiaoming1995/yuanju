## 1. Compose environments

- [x] 1.1 Move published application ports from the common Compose file to an explicit local overlay.
- [x] 1.2 Update the local startup helper to select common and local Compose files.
- [x] 1.3 Add the production Caddy Compose overlay and domain-based Caddyfile.

## 2. Build safety and documentation

- [x] 2.1 Exclude backend environment files from Docker build context.
- [x] 2.2 Update the Aliyun deployment guide and local miniapp startup instructions to match the actual implementation.

## 3. Verification

- [x] 3.1 Validate local and production merged Compose configurations and exposed ports; review Caddyfile routing syntax.
