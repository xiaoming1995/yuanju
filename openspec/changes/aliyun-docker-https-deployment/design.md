## Context

The existing frontend Nginx proxies `/api` and `/static` to the backend. Production requires only the `www` hostname to serve both the website and miniapp API; the separate `api` hostname must not route to the backend.

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
- Pass `WEB_DOMAIN` from a server-only `.env.production` using Compose interpolation and a Caddyfile environment placeholder. A missing domain fails configuration before launch.
- Route only the web host to `frontend:80`. The frontend Nginx proxies `/api` and `/static` to `backend:9002`, including miniapp requests that use the same public web origin.
- Exclude `.env` and `.env.*` from `backend/.dockerignore`, so secrets are absent from `COPY . .` build context. Runtime `env_file` remains available to Compose.

## Risks / Trade-offs

- [Production DNS or firewall not ready] → Caddy cannot obtain a public certificate; document prerequisites and verification.
- [Switching the base file removes implicit local published ports] → Update the local helper and document explicit local command.
- [User runs production files without domain variables] → Compose required-variable interpolation fails early.
- [Existing `backend/.env.docker` is reused on a server] → Document server-specific credentials and migration checks.
- [Users can call public `/api` paths on `www`] → Keep backend authentication and authorization on protected routes; removing a hostname does not make the web API private.

## Migration Plan

1. Add local and production overlays plus Caddyfile; remove published ports from the common file.
2. Update local helper and docs; validate both merged Compose configurations.
3. Prepare server-only environment files, DNS, firewall, and backups before deployment.
4. Rollback: stop the production stack, restore the previous release/configuration and retained data volumes; certificates remain in `caddy_data`.

## Open Questions

The actual domains, ECS region, and production credentials are deployment-time inputs and are not available in this repository.
