# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

**缘聚 (Yuanju)** is a Chinese numerology (八字/Bazi) analysis web platform. It combines algorithmic calculations (via `lunar-go`) with AI-powered natural language interpretation (DeepSeek/OpenAI). The codebase has a Go+Gin backend, React+TypeScript frontend, and PostgreSQL database — all orchestrated via Docker Compose.

---

## Commands

### Frontend (`cd frontend`)
```bash
npm run dev      # Dev server on port 5200, proxies /api → localhost:9002
npm run build    # TypeScript check + Vite production build
npm run lint     # ESLint
npm run preview  # Preview production build
```

### Backend (`cd backend`)
```bash
go run ./cmd/api       # Run API server (port 9002 by default)
go test ./...          # Run all tests
go test ./pkg/bazi/... # Run bazi-specific tests
```

### Full Stack
```bash
docker-compose up -d   # Start PostgreSQL + Redis + Backend + Frontend
docker-compose down
docker-compose logs -f
```

---

## Architecture

```
frontend (React 19 + TS + Vite)
  └── src/lib/api.ts, adminApi.ts    # Axios clients
      src/contexts/                  # AuthContext, AdminAuthContext
      src/pages/                     # User pages + src/pages/admin/
      src/components/                # Shared UI components

backend (Go + Gin)
  └── cmd/api/main.go                # Entry point, route registration
      configs/config.go              # Env config struct (reads .env)
      internal/handler/              # HTTP handlers (auth, bazi, admin)
      internal/middleware/           # JWT auth (user vs admin are separate)
      internal/model/                # Data structs
      internal/repository/           # All SQL — no SQL outside this layer
      internal/service/              # Business logic, AI client
      pkg/bazi/                      # Numerology algorithm engine
        engine.go, shishen.go, shensha.go, tiaohou.go, tiaohou_dict.go
        liuyue.go                    # 流月 (monthly luck) calculations
        jin_bu_huan_dict.go          # 进不换 lookup table (Tiaohou Gan/Zhi exact match)
      pkg/crypto/crypto.go           # AES-256-GCM for API key storage
      pkg/database/database.go       # PostgreSQL connection + DDL migrations
      pkg/seed/seed.go               # Seeds LLM providers from .env on startup
```

**Data flow (core feature):**
1. User submits birth data → `POST /api/bazi/calculate` → `pkg/bazi/engine.go` (lunar-go)
2. User requests AI report → `POST /api/bazi/report/:chart_id` → prompt template from DB → DeepSeek API → cached in `ai_reports` table
3. Admin manages LLM providers via `/api/admin/llm-providers` — active provider is used by the AI client at runtime

**Auth:** Two independent JWT systems — user token (`yj_token`) and admin token (`yj_admin_token`). They use separate secrets (`JWT_SECRET` vs `ADMIN_JWT_SECRET`) and separate middleware.

---

## Key Conventions

### Backend
- **No ORM.** All DB access via `lib/pq` directly in `internal/repository/`. Never write SQL in handlers or services.
- **All DDL migrations** are in `pkg/database/database.go`. Add new tables/columns there.
- **API keys** must be encrypted with `pkg/crypto` (AES-256-GCM) before storing. Never store plaintext keys.
- **Config** is read exclusively via `configs/config.go` `Config` struct — no direct `os.Getenv` elsewhere.
- Error response: `{"error": "message"}` — Success response: `{"data": ...}` or direct object.
- **500-line file limit.** Split files that exceed this.

### Frontend
- **No UI frameworks.** Pure CSS Variables only — no Tailwind, Ant Design, MUI.
- Admin routes use `adminApi.ts` + `AdminAuthContext`. User routes use `api.ts` + `AuthContext`.
- Routes are defined in `App.tsx`. Check there before adding new pages.

### Before Writing Any Code
Per `ENGINEERING.md`: always scan existing code first.
- New backend feature → check `internal/handler/`, `internal/service/`, `internal/repository/`
- New bazi algorithm → check `pkg/bazi/` (engine.go, shishen.go, shensha.go, tiaohou.go)
- New frontend component → check `src/components/`, `src/lib/`
- DB changes → only modify `pkg/database/database.go`

### OpenSpec Change Management
New features use the OpenSpec workflow:
- `/opsx-propose` → propose change
- `/opsx-apply` → implement tasks
- `/opsx-archive` → archive after completion
- Changes live in `openspec/changes/<change-name>/`
