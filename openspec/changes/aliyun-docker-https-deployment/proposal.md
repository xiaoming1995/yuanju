## Why

The current Compose setup publishes the frontend and backend directly for local development and has no HTTPS entry point. Production deployment needs a repeatable configuration that serves the web app and miniapp API through trusted HTTPS while preserving local development behavior.

## What Changes

- Split common, local, and production Compose configuration; keep local access on ports 3000 and 9002.
- Add a Caddy container with domain routing, automatic HTTPS, and persistent certificate storage for production.
- Exclude backend environment files from Docker build context.
- Update the deployment guide to match the implemented files and commands.

## Capabilities

### New Capabilities

- `https-compose-deployment`: Production Compose exposes only the HTTPS ingress while retaining the existing app services and local developer workflow.

### Modified Capabilities

None.

## Impact

Compose files, local startup script, Caddy configuration, backend `.dockerignore`, and the Aliyun deployment guide. No application API or schema change.
