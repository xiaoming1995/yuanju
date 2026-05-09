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
go run ./cmd/api                           # Run API server (port 9002 by default)
go test ./...                              # Run all tests
go test ./pkg/bazi/...                     # Run bazi-specific tests
go test ./pkg/bazi/... -run TestFuncName   # Run a single test
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
        DayunTimeline, LiuYueDrawer, TiaohouCard, WuxingRadar,
        YongshenBadge, MingpanAvatar, BottomNav, ParticleBackground

backend (Go + Gin)
  └── cmd/api/main.go                # Entry point, route registration
      configs/config.go              # Env config struct (reads .env)
      internal/handler/              # HTTP handlers
        auth_handler.go, bazi_handler.go
        admin_handler.go, admin_prompt.go
        celebrity_handler.go, shensha_handler.go
        algo_config_handler.go, algo_tiaohou_handler.go
      internal/middleware/           # JWT auth (user vs admin are separate)
      internal/model/                # Data structs
      internal/repository/           # All SQL — no SQL outside this layer
        repository.go, admin_repository.go, prompt_repository.go
        celebrity_repository.go, liunian_repository.go
        shensha_repo.go, algo_config_repository.go, algo_tiaohou_repository.go
      internal/service/              # Business logic, AI client
        ai_client.go, auth_service.go, report_service.go
        celebrity_service.go, algo_config_service.go
      pkg/bazi/                      # Numerology algorithm engine
        engine.go                    # Core 四柱 calculation (lunar-go)
        shishen.go                   # 十神 (Ten Gods) classification
        shensha.go, shensha_dayun.go # 神煞 (Spirit Stars)
        tiaohou.go, tiaohou_dict.go  # 调候用神 (Seasonal Adjustment)
        yongshen.go                  # 用神推算入口（t0 调候字典优先，t1 月令权重扶抑 fallback）
        liuyue.go                    # 流月 (monthly luck) calculations
        jin_bu_huan_dict.go          # 进不换 lookup table
        algo_config.go               # Algorithmic configuration management
        event_signals.go             # 过往年份事件信号引擎（神煞/用神基底/三会三刑/伏吟反吟/空亡/大运合化/加权身强弱）
      pkg/crypto/crypto.go           # AES-256-GCM for API key storage
      pkg/database/database.go       # PostgreSQL connection + DDL migrations
      pkg/seed/seed.go               # Seeds LLM providers from .env on startup
```

**Data flow (core features):**
1. User submits birth data → `POST /api/bazi/calculate` → `pkg/bazi/engine.go` (lunar-go)
2. User requests AI report → `POST /api/bazi/report/:chart_id` → prompt template from DB → DeepSeek API → cached in `ai_reports` table
3. Streaming report → `POST /api/bazi/report-stream/:chart_id` → SSE (Server-Sent Events), handler in `bazi_handler.go:GenerateReportStream`
4. Liunian (流年) report → `POST /api/bazi/liunian-report/:chart_id` → AI analysis for a specific year
5. Admin manages LLM providers, prompts, algo config, and tiaohou rules via `/api/admin/` routes

**Auth:** Two independent JWT systems — user token (`yj_token`) and admin token (`yj_admin_token`). They use separate secrets (`JWT_SECRET` vs `ADMIN_JWT_SECRET`) and separate middleware. `OptionalAuth()` middleware allows unauthenticated access while passing user context when a token is present.

---

## Key Conventions

### Backend
- **No ORM.** All DB access via `lib/pq` directly in `internal/repository/`. Never write SQL in handlers or services.
- **All DDL migrations** are in `pkg/database/database.go`. Add new tables/columns there.
- **API keys** must be encrypted with `pkg/crypto` (AES-256-GCM) before storing. Never store plaintext keys.
- **Config** is read exclusively via `configs/config.go` `Config` struct — no direct `os.Getenv` elsewhere.
- Error response: `{"error": "message"}` — Success response: `{"data": ...}` or direct object.
- **500-line file limit.** Split files that exceed this.
- **Algo config** (`algo_config.go`) is loaded on startup via `service.LoadAlgoConfig()` and provides runtime-tunable algorithm parameters managed through the admin UI.
- **流年信号语义按年龄分段** (`event_signals.go::YoungAgeCutoff = 18`)：起运 ~ 18 岁前自动启用读书期语义重映射，财/官/印/食伤/比劫/桃花/合冲日支 等事件输出 `学业_*` / `性格_*` Type；18 岁及以后回归原 财运/事业/婚恋 语义。`GenerateDayunSummariesStream` 按该段大运中 age<18 占比给 AI prompt 注入读书期/跨界期提示词。

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
- `/opsx:propose` → propose change
- `/opsx:apply` → implement tasks
- `/opsx:archive` → archive after completion
- Changes live in `openspec/changes/<change-name>/`
