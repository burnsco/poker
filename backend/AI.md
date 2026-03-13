# Backend — AI Agent Context

Go/Gin backend for the poker platform. Entry point: `cmd/server/main.go`.

## Tech stack

| Layer | Library |
|-------|---------|
| HTTP framework | `github.com/gin-gonic/gin` |
| WebSocket | `github.com/gorilla/websocket` |
| ORM | `gorm.io/gorm` |
| Auth | `github.com/golang-jwt/jwt/v5` |
| DB drivers | postgres + sqlite (sqlite used in tests) |
| Test assertions | `github.com/stretchr/testify` |

Go module: `poker-backend`

## Package map

```
cmd/server/main.go      Gin setup, CORS, route wiring, server start

internal/api/
  handlers.go           All HTTP handlers + RegisterRoutes
  handlers_test.go      HTTP handler tests

internal/auth/
  auth.go               JWT sign/validate, AuthMiddleware() gin middleware

internal/db/
  db.go                 InitDB() — GORM auto-migrate, connection setup
  db_test_helper.go     In-memory SQLite setup for tests

internal/game/
  registry.go           TableRegistry singleton, CreateTable/GetTable
  table.go              Table struct, NewTable, GetState, GetStateFor, Subscribe
  state_manager.go      initialState, seat builders
  betting_engine.go     ProcessAction — fold/check/call/bet/raise logic
  betting_engine_old.go DEAD CODE — do not touch or reference
  showdown_resolver.go  ResolveShowdown, side pot splitting
  hand_evaluator.go     7-card best hand evaluation
  bot_engine.go         CFR-based bot decision making
  presence_tracker.go   Disconnect/reconnect/ghost-seat cleanup
  utils.go              Deck helpers, card utilities

internal/models/
  models.go             User, Player, HandState, TableState, SidePot, HandResult

internal/ws/
  hub.go                WebSocket upgrader, Client, HandleWebSocket, message loop
```

## REST endpoints

```
GET  /api/health
GET  /api/tables
POST /api/tables                   [auth required] — {table_id, with_bots}
GET  /api/tables/:table_id
POST /api/tables/:table_id/actions [auth required] — {action, ...}

POST /api/users/register
POST /api/users/log-in
POST /api/users/guest
DELETE /api/users/log-out
GET  /api/users/me                 [auth required]
```

## WebSocket contract

**Endpoint:** `GET /socket/websocket` (auth cookie required before upgrade)

**Message format:** `[join_ref, msg_ref, topic, event, payload_object]`

**Topics:** `poker:<table_id>` e.g. `poker:default`

**Client → Server events:**
- `phx_join` — join topic, payload `{token?, player_id?}` (player_id ignored, session used)
- `table_action` — `{action: "fold"|"check"|"call"|"bet"|"raise", amount?, seat?}`
- `sit_down` — `{seat: number, name: string}`
- `stand_up`
- `start_hand`
- `add_chips` — `{amount: number}`
- `request_bots`

**Server → Client events:**
- `phx_reply` — ack/nack for join
- `table_update` — full `TableState` (hole cards masked per viewer)
- `new_msg` — chat/log messages

## Security rules (critical)

1. **Never trust client-supplied `player_id`** — always overwrite from JWT session in `hub.go`
2. **Always use `GetStateFor(viewerID)`** — never broadcast raw `GetState()` directly
3. WS connections without a valid `_poker_key` cookie must be rejected before upgrade

## Game state machine

`GameState` field on `TableState`:
- `waiting_for_hand` → `hand_in_progress` → `waiting_for_hand`

`HandState.Stage` during a hand:
- `preflop` → `flop` → `turn` → `river` → `showdown`

`HandState.Status`:
- `waiting` (between hands)
- `active` (hand running)
- `complete`

Player `Status` values:
- `ACTIVE`, `FOLDED`, `ALL_IN`, `SITTING_OUT`

## Testing

```bash
cd backend
go test ./...
go test ./internal/game/... -v
go test ./internal/api/... -v
go test -bench=. ./internal/game/...
```

Tests use in-memory SQLite via `db_test_helper.go`. Do not call `db.InitDB()` in tests.

Coverage is low (~38% in game/). Priority areas for new tests:
1. Full hand simulation (table lifecycle)
2. Side pot scenarios
3. Bot decision determinism with seeded rand
4. Presence tracker disconnect flows

## Constants (game/table.go)

```go
SmallBlind             = 10
BigBlind               = 20
StartingStack          = 5000
HandDelay              = 5 * time.Second
BotDelay               = 450 * time.Millisecond
DisconnectedHumanDelay = 30 * time.Second
Seats                  = []int{1, 2, 3, 4, 5, 6, 7, 8}
```

## Adding a new table action

1. Add handler in `betting_engine.go` `ProcessAction` switch
2. Add WS event routing in `ws/hub.go` `handleAction`
3. Add REST action support in `api/handlers.go` `TableAction`
4. Document new event in the WS contract above

## Concurrency model

- Every `Table` has a `sync.RWMutex` (`t.mu`)
- Read ops use `RLock` / `RUnlock`
- Write ops use `Lock` / `Unlock`
- Never hold the lock while doing I/O or calling `BroadcastChan`
- `BroadcastChan` is buffered (100) — don't block on it
