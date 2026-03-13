I want you to add high-quality tests and improve coverage for my Golang backend for an online poker game.

Your goal is not just to increase the coverage number, but to create meaningful tests that verify real game correctness, edge cases, and regression safety.

Project context:

- Backend language: Go
- App type: online poker game backend
- Focus: reliable game logic, backend correctness, and prevention of subtle poker rule bugs
- Prioritize maintainable tests over clever or overly complex ones

What I want you to do:

1. Analyze the codebase first

- Identify the core game logic packages and files
- Identify what is currently tested and what is untested
- Point out weak areas where bugs are most likely
- Summarize the current testing structure before making changes

2. Add meaningful automated tests
   Focus especially on:

- Deck creation and shuffling correctness
- Card dealing logic
- Hand evaluation correctness
- Winner determination and tie-breaking
- Betting round logic
- Blind posting
- Turn order progression
- Fold / check / call / bet / raise flows
- Pot calculation
- Side pot handling
- All-in scenarios
- Round transitions (preflop, flop, turn, river, showdown)
- Player elimination / disconnected player handling if applicable
- State validation and illegal move rejection
- Game reset / next hand initialization

3. Cover poker-specific edge cases
   I care a lot about tricky correctness issues. Add tests for:

- Exact tie situations
- Split pots
- Multiple side pots
- All players folding except one
- Minimum raise rules
- Insufficient chips for full call or raise
- Dealer / small blind / big blind rotation
- Heads-up rules if supported
- Showdown with multiple remaining players
- Ace-low straight handling
- Correct ranking between similar hands
- Duplicate card prevention
- Empty deck / invalid state protection

4. Test quality requirements

- Prefer table-driven tests where appropriate
- Keep tests readable and organized
- Avoid brittle tests tied too closely to implementation details
- Use deterministic setups when possible instead of random behavior
- Mock or isolate external dependencies where needed
- Add helper builders/factories for repeated test setup if useful
- Do not add pointless tests just for coverage inflation

5. Coverage and structure

- Run coverage and identify which important paths remain uncovered
- Improve coverage in the most valuable backend areas first
- Organize tests by package and feature clearly
- Refactor minor code only if necessary to make it more testable
- If refactoring, keep production behavior unchanged

6. Deliverables
   After making changes, provide:

- A summary of what tests were added
- What important scenarios are now covered
- Any remaining risky untested areas
- Coverage before and after
- Recommendations for next testing priorities

Implementation rules:

- Use idiomatic Go testing practices
- Use the standard testing package unless there is already a strong existing framework in the repo
- Reuse existing patterns if the codebase already has testing conventions
- Prefer small focused tests plus a few higher-level game flow tests
- Add regression tests for any bug-prone logic you discover

Important:
Do not just write shallow happy-path tests. I want robust tests that would actually catch real poker backend bugs.
Start by inspecting the current backend structure and then implement the most valuable tests first.
