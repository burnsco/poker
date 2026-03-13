I want you to add high-quality frontend tests and improve test coverage for my online poker game frontend.

Frontend stack:
- React
- Vite
- Tailwind CSS
- shadcn/ui

Potential architecture to use or improve:
- Zustand for client state
- TanStack Query for server state / API caching
- WebSockets for real-time game updates

Important:
If this project does NOT already use Zustand, TanStack Query, or WebSockets where appropriate, first evaluate whether they should be added. If they would materially improve the architecture for a real-time poker app, explain why and implement them carefully where justified. Do not force them in blindly, but do not ignore them either. This is a real-time multiplayer poker frontend, so I want the architecture to be solid, scalable, and testable.

Your goal is NOT to inflate coverage with shallow tests. Your goal is to create meaningful, maintainable tests that verify real user behavior, UI correctness, state transitions, real-time update handling, and regression safety.

Core priorities:
- Test real gameplay UI behavior
- Catch regressions in critical poker interactions
- Verify client state and server state behavior
- Verify real-time update handling
- Ensure core poker screens behave correctly under realistic scenarios
- Keep tests readable, maintainable, and actually useful

What I want you to do:

1. Audit the frontend architecture first
- Inspect the current frontend structure
- Identify pages, major components, hooks, stores, context providers, query usage, websocket/realtime handling, and utilities
- Identify what testing setup already exists
- Identify what is currently tested and what is missing
- Identify high-risk bug areas
- Identify whether the app already uses:
  - Zustand
  - TanStack Query
  - WebSockets or another real-time transport
- If one or more of these are missing, explicitly assess whether they should be added for this poker frontend
- Summarize your findings before making major changes

2. Improve the frontend architecture where needed
If the current setup is weak for a real-time poker game, improve it pragmatically.

Specifically:
- If local UI/game state is scattered or hard to test, consider introducing Zustand
- If API data fetching or caching is ad hoc, consider introducing TanStack Query
- If real-time updates are missing or poorly structured, evaluate adding or improving WebSocket handling
- If you add any of these, do it cleanly and minimally, without unnecessary rewrites
- Keep architecture changes focused on improving correctness, maintainability, and testability
- Explain any architectural changes clearly

3. Set up or improve the testing stack
Use the best fit for a Vite + React app:
- Vitest
- React Testing Library
- @testing-library/user-event
- Mock Service Worker for API mocking where needed

If missing, add them properly.
Also add any necessary test utilities for:
- providers
- router
- Zustand store setup/reset
- TanStack Query test client
- WebSocket mocking
- reusable render helpers

4. Add high-value tests for the most important frontend behavior

Focus especially on:

Game table / main gameplay screen
- Correct rendering of:
  - seats
  - players
  - chip stacks
  - hole cards
  - community cards
  - pot and side pots
  - dealer button
  - small blind / big blind markers
  - turn indicators
  - folded / all-in / disconnected / sitting out states
- Correct rendering across phases:
  - waiting
  - preflop
  - flop
  - turn
  - river
  - showdown
- Correct winner and result display
- Correct hidden vs revealed card behavior

Player action controls
- Fold / check / call / bet / raise actions render correctly
- Available actions change correctly based on game state
- Invalid actions are disabled or blocked
- Bet / raise controls enforce min/max constraints
- It is impossible in the UI to take actions when it is not the player’s turn
- Action UI updates immediately and correctly after state changes

Lobby / room / table setup flows if present
- Create room / create table
- Join room / join table
- Seat selection
- Buy-in validation
- Ready state toggles
- Leave table flow
- Waiting for players UI
- Loading, empty, and error states
- Reconnect handling if present

Real-time behavior
- UI updates correctly when websocket events arrive
- Pot, stacks, community cards, action prompts, and player statuses stay in sync with live updates
- Reconnection flow works correctly if supported
- Duplicate or out-of-order events do not corrupt the visible UI
- Stale state is cleared properly between hands or rounds

5. Test Zustand, TanStack Query, and WebSocket logic directly where appropriate

If Zustand is used or added:
- Test store logic
- Test derived selectors
- Test state transitions
- Test reset behavior between hands / games
- Ensure stores do not leak state between tests

If TanStack Query is used or added:
- Test query loading, success, and error states
- Test invalidation/refetch behavior where relevant
- Test optimistic UI only if it actually exists
- Use a proper isolated test QueryClient setup

If WebSocket logic is used or added:
- Test event handling
- Test malformed/unknown event handling
- Test reconnect behavior if implemented
- Test that UI reacts correctly to incoming events
- Keep websocket mocks realistic but simple

6. Cover poker-specific frontend edge cases
I care about subtle correctness and UI-state bugs. Add tests for:
- Split pot display
- Multiple winners
- Side pot display
- Heads-up layout if supported
- Dealer / blind marker rotation
- Community cards revealing street by street
- Showdown with multiple players
- Folded players remaining visible but inactive
- All-in players shown correctly
- Short-stack / low-chip edge states
- Empty seat vs occupied seat rendering
- Spectator vs seated player behavior if supported
- Long player names or large chip values not breaking critical UI logic
- Action controls disappearing or disabling correctly when state changes
- End-of-hand reset state displaying correctly for the next hand

7. Test hooks and utilities where they matter
Add focused tests for:
- custom hooks managing game state or action eligibility
- timer / countdown hooks if applicable
- connection/reconnection hooks if applicable
- utility functions for:
  - chip formatting
  - action labels
  - player position labels
  - card formatting
  - determining available actions
  - derived display state

Do not waste time testing trivial wrappers or purely decorative components.

8. Testing standards
- Prefer behavior-focused tests over implementation-detail tests
- Use React Testing Library idioms
- Query by role, label, and accessible text where possible
- Avoid brittle snapshots
- Use snapshots only when they provide real value and remain stable
- Build reusable render/test helpers
- Keep tests organized by feature
- Use table-driven patterns where useful
- Keep tests deterministic
- Mock only network and realtime boundaries, not the entire app
- Refactor lightly if necessary to improve testability, but do not change intended behavior

9. Coverage priorities
Prioritize the most valuable areas first:
- main game table
- action controls
- state/store logic
- real-time event handling
- lobby/join/buy-in flows
- loading/error/reconnect states

Do not waste effort chasing 100% coverage on low-value files.

10. Deliverables
After making changes, provide:
- Summary of the existing architecture you found
- Whether Zustand, TanStack Query, and WebSockets were already present
- If any were missing, whether you added them and why
- Summary of test setup and utilities added
- List of tests added
- Important user flows and edge cases now covered
- Remaining risky or under-tested areas
- Coverage before and after
- Recommended next testing priorities

Important:
Do not write shallow happy-path-only tests.
Do not just chase the coverage metric.
I want robust tests that would actually catch real regressions in a real-time online poker frontend.

Start by auditing the current architecture, explicitly evaluating whether Zustand, TanStack Query, and WebSockets are present and appropriate, then implement the highest-value architectural improvements and tests first.