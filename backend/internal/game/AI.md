# Game Engine — AI Agent Context

Deep-dive context for the `backend/internal/game/` package.

## Files

| File | Responsibility |
|------|---------------|
| `table.go` | `Table` struct, `NewTable`, `GetState`/`GetStateFor`, `Subscribe`/`Unsubscribe`, `broadcast` |
| `state_manager.go` | `initialState`, `buildEmptySeat`, `buildBotPlayer`, `initialHandState` |
| `registry.go` | `TableRegistry` singleton, `CreateTable`, `GetTable`, `ListActiveTables` |
| `betting_engine.go` | `ProcessAction` — all betting logic, pot management, action validation |
| `betting_engine_old.go` | **DEAD CODE — do not use or reference** |
| `showdown_resolver.go` | `ResolveShowdown` — compares hands, assigns winnings, splits side pots |
| `hand_evaluator.go` | 7-card best hand detection, hand ranking |
| `bot_engine.go` | CFR-based bot decisions, personality profiles |
| `presence_tracker.go` | Disconnect detection, ghost seat cleanup, reconnect logic |
| `utils.go` | `NewDeck`, `ShuffleDeck`, `rotateSeat`, misc helpers |

## Table concurrency model

```go
type Table struct {
    mu          sync.RWMutex
    state       models.TableState
    BroadcastChan chan json.RawMessage   // buffered 100
    autoTimer   *time.Timer
    timerSeq    int
    subscribers map[chan struct{}]bool
    // subsystems
    bettingEngine    *BettingEngine
    showdownResolver *ShowdownResolver
    presenceTracker  *PresenceTracker
}
```

**Rules:**
- All reads: `t.mu.RLock()` / `t.mu.RUnlock()`
- All writes: `t.mu.Lock()` / `t.mu.Unlock()`
- Never hold the lock while sending to `BroadcastChan` or doing I/O
- Subsystems receive a `*models.TableState` pointer and a `log func(string)` — they must only be called when the table lock is held

## State flow

```
NewTable(id, withBots)
  → initialState()       builds TableState with 8 seats
  → scheduleAutoProgress() arms timer for next hand

Table.HandleAction(event, playerID, payload)
  → bettingEngine.ProcessAction(...)
    → returns (showdown bool, err error)
  → if showdown: showdownResolver.ResolveShowdown()
  → t.broadcast()        signal subscribers + push to BroadcastChan

Table.HandlePlayerJoin / HandlePlayerLeave
  → presenceTracker manages seat lifecycle
  → t.broadcast()
```

## BettingEngine — action flow

```go
ProcessAction(action, playerID, payload, validateAuth)
```

Actions:
- `fold` — sets player.Status = "FOLDED", advances action
- `check` — only valid when toCall == 0
- `call` — calls current bet (or goes all-in if stack < toCall)
- `bet` — opens betting, sets CurrentBet
- `raise` — increases CurrentBet, enforces MinimumRaise

After each action, `advanceAfterAction` checks:
1. Is there only one active player? → end hand (fold win)
2. Has action completed the street? → advance stage or trigger showdown
3. Otherwise → move `ActingSeat` to next active player

## Hand stages

```
preflop  → deal hole cards, post blinds, betting round
flop     → deal 3 community cards, betting round
turn     → deal 1 community card, betting round  
river    → deal 1 community card, betting round
showdown → reveal cards, ResolveShowdown, award pots
```

Between hands: `HandState.Status = "waiting"`, `GameState = "waiting_for_hand"`

## Side pots

`ShowdownResolver` handles side pots when players are all-in.

`SidePot` struct:
```go
type SidePot struct {
    Amount        int
    EligibleSeats []int
    WinnerSeats   []int
    WinnerAmounts map[string]int
}
```

Build side pots from `ContributedThisHand` on each player.

## Hand evaluator

`hand_evaluator.go` takes `[]string` (7 cards: 2 hole + 5 board) and returns a ranked value. Higher is better.

Hand ranks (highest first):
1. Royal Flush
2. Straight Flush
3. Four of a Kind
4. Full House
5. Flush
6. Straight
7. Three of a Kind
8. Two Pair
9. One Pair
10. High Card

Benchmark: ~137 μs per evaluation.

## Bot engine

Bots are backed by `data/cfr_table.json` (CFR-trained policy) and profiled by `data/bot_profiles.json`.

Bot profiles define personality biases (aggression, bluff frequency, etc.). The engine selects actions via the CFR policy adjusted by personality weights.

Bot action timing: `BotDelay = 450ms` (mimics human reaction).

## PresenceTracker

Tracks connected/disconnected state per player.

- Player disconnects → `disconnected_at` timestamp set, timer started
- After `DisconnectedHumanDelay` (30s): seat treated as abandoned, action auto-folded
- Bot seats are never tracked for presence

## GameState values

```
"waiting_for_hand"    Between hands, waiting to start
"hand_in_progress"    Active hand underway
```

## Constants

```go
SmallBlind             = 10
BigBlind               = 20
StartingStack          = 5000
HandDelay              = 5 * time.Second      // pause between hands
BotDelay               = 450 * time.Millisecond
DisconnectedHumanDelay = 30 * time.Second
LogLimit               = 48                    // max action log entries
Seats                  = []int{1, 2, 3, 4, 5, 6, 7, 8}
```

## Test files

```
hand_evaluator_test.go      Unit tests + benchmarks for hand ranking
betting_engine_test.go      Action validation, pot math
table_test.go               Table lifecycle
sidepot_test.go             Side pot scenarios
state_transitions_test.go   Stage progression
concurrency_test.go         Race condition checks
bot_test.go                 Bot decision smoke tests
```

Run: `cd backend && go test ./internal/game/... -v`

## Common gotchas

- `ActingSeat` is a `*int` (nullable pointer) — always nil-check before dereferencing
- Player indexes in `[]Player` are 0-based but seat numbers are 1-based (seat 1 = index 0)
- `HoleCards` for non-viewers must be masked at the Table level via `GetStateFor`, never rely on client-side masking
- `WinnerAmounts` keys are stringified seat numbers (`"1"`, `"2"`, etc.) matching JSON serialization of `map[string]int`
