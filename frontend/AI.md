# Frontend — AI Agent Context

React 19 + Vite 8 frontend for the poker platform.

## Tech stack

| Tool            | Version                 |
| --------------- | ----------------------- |
| Runtime/bundler | Vite 8                  |
| Package manager | Bun                     |
| Framework       | React 19                |
| Styling         | Tailwind v4             |
| UI primitives   | shadcn/ui (Radix-based) |
| Icons           | lucide-react            |
| HTTP            | fetch (lib/api.ts)      |
| WebSocket       | native browser WS       |
| Linter          | oxlint                  |
| Formatter       | oxfmt                   |
| Unit tests      | Bun test runner         |
| E2E tests       | Playwright              |

## Source layout

```
src/
  main.tsx              App entry — mounts React, sets up router
  App.tsx               Top-level routing + layout
  app/style.css         Global styles + Tailwind base

  game/
    PokerGameClient.ts       Core state machine — WS consumer → ClientState
    __tests__/
      PokerGameClient.test.ts

  types/
    backend.ts          TypeScript shapes for all Go API responses

  hooks/
    usePokerTable.ts    React hook wrapping PokerGameClient lifecycle

  ui/
    Renderer.ts         Canvas/DOM poker table renderer
    PokerSoundEngine.ts Sound playback mapped to PokerSoundEvent types

  lib/
    api.ts              REST client — typed wrappers around fetch
    api.test.ts
    utils.ts            cn() and other helpers

  contexts/
    AuthContext.tsx     User auth state, login/logout, guest flow

  components/
    ui/                 shadcn primitives (button, dialog, input, label)

e2e/
  lobby-to-table.spec.ts
  multi-user-poker.spec.ts
  layout.regression.spec.ts
```

## Environment variables (build-time)

```
VITE_BACKEND_URL        e.g. http://localhost:4000
VITE_BACKEND_WS_URL     e.g. ws://localhost:4000/socket
```

## Dev commands

```bash
bun install
bun run dev             # Vite dev server on :3000
bun run build
bun run preview

bun run test            # unit tests (PokerGameClient)
bun run test:e2e        # Playwright
bun run test:e2e:headed # Playwright with browser visible

bun run lint            # oxlint
bun run format          # oxfmt
```

## PokerGameClient — state machine

The most important frontend file. Consumes raw `BackendTable` JSON (from WS or REST polling) and produces `ClientState`.

**Types:**

- `BackendTable` — raw Go API shape (snake_case, from `types/backend.ts`)
- `ClientState` — normalized UI shape (camelCase, from `game/PokerGameClient.ts`)
- `ClientPlayer` — per-seat player info
- `PokerSoundEvent` — union type of sound triggers
- `HandResultSummary` — post-hand result for display

**Key methods:**

```ts
client.update(backendTable: BackendTable)  // call on every WS message
client.getState(): ClientState             // read the current derived state
client.onStateChange = () => void          // called after every update
client.onSoundEvent = (e: PokerSoundEvent) => void
client.onHandResultUpdate = (r: HandResultSummary | null) => void
```

**ClientState fields:**

```ts
players: Array<ClientPlayer | null>     // indexed by seat (1-8, 0-indexed array)
actionTo: number | null                  // seat number whose turn it is
street: "PREFLOP"|"FLOP"|"TURN"|"RIVER"|"SHOWDOWN"
pots: Array<{amount: number}>
currentBets: Map<number, number>         // seat → bet this street
board: string[]                          // community cards
winners: Array<{seat, amount}> | null
buttonSeat: number
handEndMode: "fold"|"showdown"|null
manualStartRequired: boolean
```

## API client (lib/api.ts)

All REST calls. Auth state is managed via cookies (`_poker_key`).

Key functions:

- `login(email, password)` → sets cookie
- `register(...)` → creates account + sets cookie
- `guestLogin()` → anonymous session
- `logout()`
- `getMe()` → current user
- `listTables()` → string[]
- `createTable(tableId, withBots)` → BackendTable
- `getTable(tableId)` → BackendTable
- `tableAction(tableId, action, payload)` → BackendTable

## WebSocket usage

`usePokerTable.ts` manages the WS connection lifecycle:

1. Opens WS to `VITE_BACKEND_WS_URL`
2. Sends `phx_join` for `poker:<tableId>`
3. On `table_update` events, calls `client.update(payload)`
4. Handles reconnection on disconnect

Message format mirrors the Phoenix channel protocol:

```ts
[joinRef, msgRef, topic, event, payload];
```

## shadcn/ui components

Located in `src/components/ui/`. These are copy-pasted shadcn components, not a package import. Modify them directly if needed. Uses `cn()` from `lib/utils.ts` for class merging.

## E2E tests

Playwright config in `playwright.config.ts`. Tests expect:

- Frontend running on `http://localhost:3000`
- Backend running on `http://localhost:4000`

Run `make dev` before running E2E tests locally.

## Card notation

Cards are strings like `"Ah"` (Ace of hearts), `"Td"` (Ten of diamonds).
Suits: `h` (hearts), `d` (diamonds), `c` (clubs), `s` (spades).
Ranks: `2–9`, `T`, `J`, `Q`, `K`, `A`.

Empty/unknown hole cards are represented as `""` or `null`.

## Style conventions

- All files: TypeScript (`.ts` / `.tsx`)
- Tailwind utility classes only — no custom CSS except `app/style.css`
- Component files use PascalCase, utility files use camelCase
- API payloads stay `snake_case` to match Go JSON tags
- Internal React state / props use `camelCase`
