# Poker Platform — AI Agent Context

Real-time Texas Hold'em poker platform. Go/Gin backend + React/Vite frontend in a monorepo.

## Repo layout

```
backend/          Go backend (Gin, Gorilla WebSocket, GORM)
  cmd/server/     Entry point — main.go
  internal/
    api/          HTTP handlers + route registration
    auth/         JWT middleware
    db/           GORM init, migrations
    game/         Core game engine (table, betting, bots, presence)
    models/       Shared structs (User, Player, TableState, HandState)
    ws/           WebSocket hub + client lifecycle
  data/           CFR bot data (bot_profiles.json, cfr_table.json)
frontend/         React 19 + Vite + Tailwind v4
  src/
    game/         PokerGameClient.ts — state machine bridging WS → UI
    types/        backend.ts — Go response shapes
    hooks/        usePokerTable.ts
    ui/           Renderer.ts, PokerSoundEngine.ts
    lib/          api.ts — REST client
    contexts/     AuthContext.tsx
    components/   shadcn/ui primitives
  e2e/            Playwright tests
load/             k6 stress test scripts + profiles
.plans/           Active work notes
```

## Quick commands

```bash
# Full stack (Docker)
make up

# Local dev (preferred)
make dev          # starts db+backend via Docker, then bun dev for frontend

# Backend only
cd backend && go run cmd/server/main.go

# Frontend only
cd frontend && bun install && bun run dev

# Tests
make backend-check      # fmt + lint + test
make frontend-check     # lint + test + build
cd backend && go test ./...
cd frontend && bun run test

# Stress testing
make stress-stack-up
make stress-low         # → make stress-medium/high/extreme/insane
```

## Ports

| Service  | Port |
|----------|------|
| Backend  | 4000 |
| Frontend | 3000 |
| Postgres | 5432 (Docker) |

## Environment

Copy `.env.example` → `.env`. Required vars:
- `JWT_SECRET` — signs player session cookies
- `SECRET_KEY_BASE` — legacy var, kept for compat
- `VITE_BACKEND_URL` / `VITE_BACKEND_WS_URL` — frontend build-time

## Architecture

### WebSocket protocol

The client opens a single WS to `/socket/websocket`. Messages are JSON arrays:
```
[join_ref, msg_ref, topic, event, payload]
```

Topics follow `poker:<table_id>`. Events include:
- `phx_join` — subscribe to a table
- `table_action` — game action (fold/check/call/bet/raise)
- `sit_down`, `stand_up`, `start_hand`, `add_chips`
- `table_update` — server push of full `TableState`

### Identity

Players authenticate via the `_poker_key` cookie (JWT). **The backend always overwrites `player_id` from the session — never trust client-supplied identity.**

### Game state flow

```
Table (mutex-guarded)
  └─ BettingEngine    — action processing, pot management
  └─ ShowdownResolver — hand comparison, winner determination
  └─ PresenceTracker  — disconnect/reconnect, ghost seat cleanup
  └─ BotEngine        — CFR-based bot decisions
```

`TableRegistry` is a global singleton. Pre-seeded tables:
- `"default"` — bot warmup table (all bots)
- `"human-table"` — empty table for human players

### Frontend state

`PokerGameClient` is a state machine that consumes raw `BackendTable` responses and produces the typed `ClientState` that React components read. It also emits `PokerSoundEvent` for the sound engine.

## Key files for common tasks

| Task | Files |
|------|-------|
| Add a new API endpoint | `backend/internal/api/handlers.go` + `RegisterRoutes` |
| Change game rules / blind amounts | `backend/internal/game/table.go` constants block |
| Bot behavior | `backend/internal/game/bot_engine.go`, `data/bot_profiles.json` |
| WS message handling | `backend/internal/ws/hub.go` |
| Frontend WS state | `frontend/src/game/PokerGameClient.ts` |
| REST API calls | `frontend/src/lib/api.ts` |
| DB schema | `backend/internal/models/models.go`, `backend/internal/db/db.go` |
| Auth | `backend/internal/auth/auth.go` |

## Known issues / active work

See `poker_audit.md` for the full security audit (March 2026).
See `.plans/backend-optimize.md` for in-progress optimizations.

Critical items fixed:
- WS identity spoofing
- Cross-player action impersonation
- Hole card server-side masking via `GetStateFor(viewerID)`

Remaining concerns:
- Test coverage in `game/` package is ~38% — needs integration tests
- `betting_engine_old.go` is dead code — should be removed
- Rate limiting on table actions is disabled during stress testing (`TABLE_ACTION_RATE_LIMIT_DISABLED`)

## Conventions

- Go package name matches directory name
- All game mutations happen inside `Table.mu` (RWMutex)
- Frontend uses `snake_case` for all API/WS payloads to match Go JSON tags
- No ORM migrations are auto-run in tests — use `db_test_helper.go`
- Bun is the JS runtime and package manager (not Node/npm)
- Linter: oxlint (frontend), golangci-lint (backend)
