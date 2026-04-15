# Poker App Review & Optimization Plan

> Generated 2026-04-15 by Opus review. All prior critical issues from March 2026 review confirmed fixed.

---

## Phase 1: Security Fixes (Priority: CRITICAL)

### 1.1 Strip deck and hole cards from broadcast state
- **File:** `backend/lib/poker_backend/table.ex` (line 2029-2034)
- **File:** `backend/lib/poker_backend_web/channels/table_channel.ex` (line 40-43)
- **Problem:** `broadcast_state/1` sends the ENTIRE table state (including `deck` and all players' `hole_cards`) over PubSub to every connected client. The TableChannel pushes this raw state via `handle_info`. A client sniffing WebSocket messages can see every player's cards and the remaining deck order.
- **Fix:** Filter the state per-socket in the channel's `handle_info`. Strip `deck` entirely. For each player, only reveal `hole_cards` if:
  - The player is the socket's own player (matched via `socket.assigns.player_id`)
  - The player is a bot (`is_bot == true`)
  - The player has `show_cards == true`
  - Otherwise replace hole_cards with `[nil, nil]`
- **Implementation:**
  ```elixir
  # table_channel.ex - replace handle_info
  def handle_info({:table_event, %{type: "table_state", state: state} = payload}, socket) do
    filtered = filter_state_for_player(state, socket.assigns.player_id)
    push(socket, "table_event", %{payload | state: filtered})
    {:noreply, socket}
  end

  # Also filter the join reply in join/3

  defp filter_state_for_player(state, viewer_player_id) do
    filtered_players =
      Enum.map(state.players, fn player ->
        if should_reveal_cards?(player, viewer_player_id) do
          player
        else
          %{player | hole_cards: [nil, nil]}
        end
      end)

    state
    |> Map.put(:players, filtered_players)
    |> update_in([:hand_state], &Map.delete(&1, :deck))
  end

  defp should_reveal_cards?(player, viewer_player_id) do
    player.is_bot or
      player.show_cards or
      to_string(player.player_id) == to_string(viewer_player_id)
  end
  ```
- **Also:** The HTTP `show` endpoint (`table_controller.ex:14-19`) returns raw state too. Add similar filtering there, or strip the deck at minimum.

### 1.2 Fix balance validation and registration cast
- **File:** `backend/lib/poker_backend/accounts/user.ex` (line 22)
- **File:** `backend/lib/poker_backend/accounts.ex` (lines 73-77)
- **Problem 1:** `registration_changeset` casts `:balance` from user input. A registration request payload could set an arbitrary starting balance (e.g., `{"user": {"balance": 999999}}`).
- **Problem 2:** `update_user_balance` accepts any value with no validation. Negative balances possible.
- **Fix:**
  - Remove `:balance` from the cast list in `registration_changeset` (line 22). The schema default of 5000 will be used.
  - Add validation in `update_user_balance`:
    ```elixir
    def update_user_balance(user, balance) when is_integer(balance) and balance >= 0 do
      user
      |> Ecto.Changeset.cast(%{balance: balance}, [:balance])
      |> Ecto.Changeset.validate_number(:balance, greater_than_or_equal_to: 0)
      |> Repo.update()
    end
    ```

### 1.3 Add authentication to WebSocket channel actions
- **File:** `backend/lib/poker_backend_web/channels/table_channel.ex` (lines 30-33)
- **Problem:** `handle_in("action", ...)` forwards actions directly to the GenServer without any auth check. The HTTP path requires authentication, but a client could bypass this by sending actions through the WebSocket.
- **Fix options (pick one):**
  - **Option A (recommended):** Remove the `handle_in("action", ...)` handler entirely. The frontend only uses HTTP for actions anyway.
  - **Option B:** Add auth check in the handler that validates the action's player_id matches socket.assigns.player_id.

### 1.4 Add table_id validation on WebSocket join
- **File:** `backend/lib/poker_backend_web/channels/table_channel.ex` (line 5)
- **Problem:** HTTP path validates table IDs with regex `^[a-zA-Z0-9_\-]+$` and max 64 chars. WebSocket join accepts any string, allowing creation of arbitrary table GenServers (unbounded memory).
- **Fix:** Add the same validation at the top of `join/3`:
  ```elixir
  @max_table_id_length 64
  @table_id_pattern ~r/^[a-zA-Z0-9_\-]+$/

  def join("table:" <> table_id, params, socket) do
    if byte_size(table_id) > 0 and
       byte_size(table_id) <= @max_table_id_length and
       Regex.match?(@table_id_pattern, table_id) do
      # ... existing join logic
    else
      {:error, %{reason: "invalid_table_id"}}
    end
  end
  ```

---

## Phase 2: Data Integrity Fixes (Priority: HIGH)

### 2.1 Fix balance race condition with atomic updates
- **File:** `backend/lib/poker_backend/accounts.ex` (lines 73-77)
- **File:** `backend/lib/poker_backend/table.ex` (lines 1965-2009)
- **Problem:** `update_user_balance` does a simple `Repo.update()` to SET the balance. If two tables conclude hands simultaneously for the same player, the last write wins (lost update). Since `persist_human_balances` runs in a `Task`, this is a real race.
- **Fix:** Use an atomic increment/decrement approach instead of setting the absolute value. Change `persist_player_balance` to compute the delta (new_stack - old_balance) and use `Repo.update_all`:
  ```elixir
  # In accounts.ex, add:
  def set_user_balance(user_id, new_balance) when is_integer(user_id) and is_integer(new_balance) do
    from(u in User, where: u.id == ^user_id)
    |> Repo.update_all(set: [balance: max(new_balance, 0), updated_at: DateTime.utc_now()])
    |> case do
      {1, _} -> :ok
      {0, _} -> {:error, :not_found}
    end
  end
  ```
  Note: The current approach of setting absolute balance is inherently racy when a player is at multiple tables. A fully correct solution would track per-table stakes separately. For now, the atomic set is acceptable since multi-tabling is rare.

### 2.2 Add refill rate limiting
- **File:** `backend/lib/poker_backend_web/controllers/user_session_json_controller.ex`
- **Problem:** `refill_user_balance` has no cooldown. A script could call `/api/users/refill` repeatedly during a hand to always maintain max balance.
- **Fix:** Add a cooldown check. Either:
  - Track `last_refill_at` on the user schema and reject if < 10 minutes ago
  - Or use a simple ETS-based rate limit similar to the existing table action rate limiter

---

## Phase 3: Code Quality & Performance (Priority: MEDIUM)

### 3.1 Deduplicate `getBackendUrl()`
- **Files:**
  - `frontend/src/App.tsx` (lines 19-26)
  - `frontend/src/hooks/usePhoenixTable.ts` (lines 13-20)
  - `frontend/src/contexts/AuthContext.tsx` (lines 27-38)
  - Also `getBaseUrl()` at `usePhoenixTable.ts:41-44` is a simpler duplicate
- **Fix:** Extract to `frontend/src/lib/config.ts`:
  ```typescript
  export const BACKEND_URL = getBackendUrl();
  export const WEBSOCKET_BASE = getWebSocketBase();
  ```
  Import from there in all three files. Remove `getBaseUrl()` (dead code after consolidation).

### 3.2 Add bulk table summary endpoint to fix N+1 lobby polling
- **File:** `backend/lib/poker_backend_web/controllers/table_controller.ex`
- **File:** `frontend/src/App.tsx` (lines 267-313)
- **Problem:** Lobby calls `fetchActiveTables()` then `fetchTableState(tableId)` for EVERY table. With 10 tables, that's 11 HTTP requests every 8 seconds per lobby viewer.
- **Fix:**
  - Add `GET /api/tables/summary` endpoint that returns all active table states in one response
  - Update `refreshGames` to use the new endpoint

### 3.3 Remove console.log statements from production code
- **File:** `frontend/src/contexts/AuthContext.tsx` (lines 104, 120-121, 123)
- **Fix:** Remove `console.log("AuthProvider: Attempting registration...")` and similar debug statements. Keep `console.error` for actual errors.

### 3.4 Fix unused variable
- **File:** `frontend/src/App.tsx` (line 220)
- **Problem:** `const suffix = parts.pop()` assigns but never uses `suffix`.
- **Fix:** Change to `parts.pop()` (discard return value).

### 3.5 Memoize `loadPlayerIdentity()`
- **File:** `frontend/src/hooks/usePhoenixTable.ts` (line 87)
- **Problem:** Called on every render (reads localStorage each time).
- **Fix:** Wrap in `useMemo` or move into the effect.

---

## Phase 4: Architecture Improvements (Priority: LOW / FUTURE)

### 4.1 Split App.tsx into separate components
- `frontend/src/App.tsx` is 1712 lines with `LobbyScreen`, `TableScreen`, and `App` all in one file.
- Extract to `LobbyScreen.tsx`, `TableScreen.tsx`, and shared components.
- The bet-sizing logic (lines 915-1001) deserves its own hook (`useBetSlider`).

### 4.2 Migrate Renderer to React components
- `Renderer.ts` uses imperative DOM manipulation (`document.getElementById`, `.innerHTML`) which bypasses React's rendering. This creates a parallel rendering path that is fragile.
- Long-term: migrate Renderer logic into React components.

### 4.3 Add WebSocket rate limiting
- `table_channel.ex` `handle_in` has no rate limiting. The HTTP path has ETS-based rate limiting but WebSocket is open.
- Add per-socket rate limiting in the channel.

### 4.4 Enable WebSocket origin check in production
- `endpoint.ex` has `check_origin: false`. Enable in production config.

### 4.5 Hand evaluator optimization
- `hand_evaluator.ex` generates C(7,5)=21 combinations per evaluation via recursive list comprehension.
- If CPU becomes a bottleneck, replace with a lookup-table evaluator.

---

## Implementation Order

| Step | Items | Est. Effort |
|------|-------|-------------|
| 1 | 1.2 (balance validation) | 10 min |
| 2 | 1.4 (table_id validation on WS) | 10 min |
| 3 | 1.3 (remove WS action handler) | 5 min |
| 4 | 1.1 (filter broadcast state) | 30 min |
| 5 | 2.1 (balance race condition) | 15 min |
| 6 | 2.2 (refill rate limiting) | 15 min |
| 7 | 3.1 (deduplicate getBackendUrl) | 10 min |
| 8 | 3.3 + 3.4 + 3.5 (cleanup) | 10 min |
| 9 | 3.2 (bulk summary endpoint) | 20 min |
