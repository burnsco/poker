# Poker Platform — Agent Context
> Tool-agnostic version of CLAUDE.md. Works with Claude Code, Codex, Gemini CLI, or any agent that reads project context files.

## What this is

A real-time Texas Hold'em platform. Go backend + React frontend. Docker-first local dev.

## One-sentence descriptions

| Component | Description |
|-----------|-------------|
| `backend/` | Go/Gin HTTP + WebSocket server. Manages all game state in-process. |
| `frontend/` | React 19 + Vite SPA. Canvas renderer + WebSocket client. |
| `backend/internal/game/` | Core poker engine: table, betting, showdown, bots, presence. |
| `backend/internal/ws/` | WebSocket hub — one connection per player, Phoenix-protocol messages. |
| `backend/internal/api/` | REST handlers for auth, table CRUD, and actions. |
| `frontend/src/game/PokerGameClient.ts` | State machine: transforms raw backend JSON into typed UI state. |

## Stack at a glance

**Backend:** Go 1.25, Gin, Gorilla WebSocket, GORM, PostgreSQL, JWT cookies  
**Frontend:** React 19, Vite 8, Tailwind v4, Bun, shadcn/ui, Playwright  
**Infra:** Docker Compose, GitHub Actions CI, k6 stress tests

## Run it

```bash
make up          # full stack via Docker
make dev         # Docker for db+backend, bun dev for frontend

# Tests
cd backend && go test ./...
cd frontend && bun run test
make backend-check    # fmt + lint + test
make frontend-check   # lint + test + build
```

## Ports

- Backend: `:4000`
- Frontend: `:3000`
- WS endpoint: `ws://localhost:4000/socket/websocket`

## Domain model

```
TableRegistry (singleton)
  └─ Table (one per table_id, mutex-guarded)
       ├─ TableState (models.TableState)
       │    ├─ []Player (8 seats)
       │    └─ HandState (current hand)
       ├─ BettingEngine   — action processing
       ├─ ShowdownResolver — winner determination
       ├─ PresenceTracker  — connection lifecycle
       └─ BotEngine        — CFR-based AI players
```

Pre-seeded tables:
- `"default"` — all bots (warmup / demo)
- `"human-table"` — empty, for human players

## REST API summary

```
GET  /api/health
GET  /api/tables
POST /api/tables                    body: {table_id, with_bots}  [auth]
GET  /api/tables/:table_id
POST /api/tables/:table_id/actions  body: {action, ...}          [auth]

POST /api/users/register
POST /api/users/log-in
POST /api/users/guest
DELETE /api/users/log-out
GET  /api/users/me                  [auth]
```

Auth is cookie-based. Cookie name: `_poker_key` (JWT).

## WebSocket protocol

Phoenix channel protocol over a single connection:
```
[join_ref, msg_ref, topic, event, payload]
```

Topic format: `poker:<table_id>`

Key events:
- `phx_join` — subscribe; server replies with full table state
- `table_update` — server push after any state change
- `table_action` — player action `{action, amount?, seat?}`
- `sit_down`, `stand_up`, `start_hand`, `add_chips`

**Security invariant:** The server always overwrites `player_id` from the JWT session. Never inject or trust a client-supplied `player_id`.

## Game state lifecycle

```
waiting_for_hand
  → (start_hand or auto-timer)
hand_in_progress
  preflop → flop → turn → river → showdown
  → (hand complete)
waiting_for_hand
```

Player status values: `ACTIVE`, `FOLDED`, `ALL_IN`, `SITTING_OUT`

## Key files by task

| Task | File(s) |
|------|---------|
| New REST endpoint | `backend/internal/api/handlers.go` |
| New game action | `backend/internal/game/betting_engine.go` + `ws/hub.go` |
| Game constants | `backend/internal/game/table.go` top of file |
| Bot AI / personalities | `backend/internal/game/bot_engine.go`, `backend/data/bot_profiles.json` |
| DB schema | `backend/internal/models/models.go` |
| Auth logic | `backend/internal/auth/auth.go` |
| Frontend WS state | `frontend/src/game/PokerGameClient.ts` |
| REST client (frontend) | `frontend/src/lib/api.ts` |
| Auth UI state | `frontend/src/contexts/AuthContext.tsx` |

## Card notation

`"Ah"` = Ace of hearts. Ranks: `2–9, T, J, Q, K, A`. Suits: `h d c s`.  
Masked / unknown cards are `""` or `null`.

## Conventions

- Go: package = directory name, RWMutex on all Table writes, no I/O inside lock
- Frontend: `snake_case` payloads match Go JSON tags; internal React state is `camelCase`
- Tests: backend uses in-memory SQLite (`db_test_helper.go`), never call `db.InitDB()`
- JS runtime: Bun only (not Node/npm)
- Dead code: `backend/internal/game/betting_engine_old.go` — ignore, do not reference

## Current known issues

See `poker_audit.md` for full security audit.

High priority remaining work:
1. Game package test coverage is ~38% — needs integration tests for full hand simulation
2. `betting_engine_old.go` should be deleted
3. Rate limiting on actions is stripped during stress tests — verify it's re-enabled in prod
4. WS reconnection handling in `usePokerTable.ts` needs exponential backoff
