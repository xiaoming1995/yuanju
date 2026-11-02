## Context

The existing Compose file publishes backend port 9002 and frontend port 3000 for local use. The frontend Nginx already proxies `/api` and `/static` to the backend. Production needs a single public ingress for the website plus a separate HTTPS API host for the miniapp. The backend environment file contains secrets and is currently in the Docker build context.

## Goals / Non-Goals

**Goals:**
- Keep the existing local startup helper and port behavior.
- Add production HTTPS with automatic certificate renewal and no direct published application/database ports.
- Keep production domains configurable without committing them.
- Keep backend secrets out of Docker build context.

**Non-Goals:**
- Provision ECS, DNS, ICP filing, or production credentials.
- Deploy to a real server or publish the miniapp.
- Alter application routes or authentication behavior.

## Decisions

- Move host port bindings out of the base Compose file into an explicit `docker-compose.local.yml`. The local helper selects both files. This prevents Compose's additive port merge from accidentally exposing ports in production.
- Add `docker-compose.prod.yml` with a Caddy service. It alone publishes 80/443, mounts `./caddy` as configuration, and persists `/data` and `/config` in named volumes.
- Pass `WEB_DOMAIN` and `API_DOMAIN` from a server-only `.env.production` using Compose interpolation and then Caddyfile environment placeholders. This makes missing domains fail configuration before launch.
- Route web host to `frontend:80` and API host to `backend:9002`. The web host retains the existing same-origin Nginx proxy and SSE handling.
- Exclude `.env` and `.env.*` from `backend/.dockerignore`, so secrets are absent from `COPY . .` build context. Runtime `env_file` remains available to Compose.

## Risks / Trade-offs

- [Production DNS or firewall not ready] → Caddy cannot obtain a public certificate; document prerequisites and verification.
- [Switching the base file removes implicit local published ports] → Update the local helper and document explicit local command.
- [User runs production files without domain variables] → Compose required-variable interpolation fails early.
- [Existing `backend/.env.docker` is reused on a server] → Document server-specific credentials and migration checks.

## Migration Plan

1. Add local and production overlays plus Caddyfile; remove published ports from the common file.
2. Update local helper and docs; validate both merged Compose configurations.
3. Prepare server-only environment files, DNS, firewall, and backups before deployment.
4. Rollback: stop the production stack, restore the previous release/configuration and retained data volumes; certificates remain in `caddy_data`.

## Open Questions

The actual domains, ECS region, and production credentials are deployment-time inputs and are not available in this repository.
