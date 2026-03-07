# ♠️ Poker Platform

A real-time, full-stack poker platform featuring a high-concurrency Elixir/Phoenix backend and a modern React frontend.

## 🚀 Tech Stack

- **Backend**: [Elixir](https://elixir-lang.org/) & [Phoenix Framework](https://www.phoenixframework.org/) (WebSockets, Channels, OTP)
- **Frontend**: [React 19](https://react.dev/), [Vite](https://vitejs.dev/), [TypeScript](https://www.typescriptlang.org/)
- **Bot AI**: [OpenSpiel](https://github.com/google-deepmind/open_spiel) (Counterfactual Regret Minimization)
- **Linting/Formatting**: [Biome](https://biomejs.dev/)
- **Infrastructure**: Docker & Docker Compose

## 🤖 Bot AI

### The short version

The bots aren't scripted. They were trained using machine learning — specifically a game-theory algorithm called **CFR (Counterfactual Regret Minimization)** — and they learned how to play poker by playing millions of hands against themselves until they stopped making obvious mistakes. The result is a strategy that's hard to exploit: they won't keep folding every time you raise, and they won't call you off with trash hands forever either.

At the table you'll find three personality types, each built on the same trained base but with different tendencies layered on top:

- **Tight** — cautious, folds weak and marginal hands, only puts money in with real strength
- **Balanced** — plays close to the mathematically optimal strategy the training produced
- **Aggressive** — raises frequently, bluffs more, and doesn't back down easily

### How the training works

The AI was trained using **[OpenSpiel](https://github.com/google-deepmind/open_spiel)**, a research library from Google DeepMind used for training game-playing AI. The algorithm is **CFR+** (an improved variant of CFR), which works like this:

1. Two virtual players play thousands of simulated poker hands against each other.
2. After every hand, each player looks back and asks: *"What would have happened if I'd played differently at each decision point?"*
3. Decisions that would have done better get reinforced. Decisions that would have done worse get discouraged.
4. Repeat tens of thousands of times. The strategy converges toward a **Nash equilibrium** — a point where neither player can improve their results by changing strategy, assuming the opponent also plays optimally.

This is the same class of algorithm behind **Libratus** and **Pluribus**, the AIs that beat professional poker players in 2017 and 2019.

### Training games

Training was done on two simplified poker variants (full Texas Hold'em has too large a game tree for tabular CFR to solve in reasonable time):

- **Leduc poker** — 6-card deck (J/Q/K x 2 suits), 2 players, 2 rounds. 50,000 iterations, ~5 min.
- **Mini Hold'em** — 6-card deck (2/3/4 x 2 suits), 2 hole cards + 1 board card, 2 rounds. 10,000 iterations, ~10 min.

The Mini Hold'em variant was specifically designed to bridge Leduc and real Hold'em: it's the same deck size as Leduc but introduces **pocket pairs** (two hole cards of the same rank), which meaningfully changes preflop strategy. This produced better action-frequency data for situations like "strong pair preflop facing a raise" that Leduc can't model.

### From training to runtime

The trained policy has thousands of entries (one per unique game situation). That's too large to embed directly. Instead, a post-processing step aggregates it into a compact lookup table keyed by:

- **Hand strength**: weak / medium / strong
- **Street**: preflop / postflop
- **Has pair**: whether the player holds a pair
- **Facing a bet**: whether someone has bet before acting

Each entry stores `{fold%, call%, raise%}` probabilities. At runtime, the bot looks up its situation, reads those probabilities, and randomly samples an action — so two bots in the same spot won't always do the same thing.

The per-personality adjustments (tight/balanced/aggressive) shift those probabilities before sampling: tight bots add extra fold weight, aggressive bots add extra raise weight.

### Retrain

```bash
make training-retrain          # 50,000 CFR+ iterations on Leduc (~5 min)
make training-retrain-holdem   # 10,000 CFR+ iterations on Mini Hold'em (~10 min)
```

## 📂 Project Structure

- `frontend/`: React + Vite client for the poker UI.
- `backend/`: Elixir service hosting table state and real-time game traffic via Phoenix Channels.
- `training/`: OpenSpiel (Python) scripts to train poker bots with CFR; see [training/README.md](training/README.md).

## 🛠️ Getting Started

### Prerequisites

- [Docker](https://www.docker.com/) & [Docker Compose](https://docs.docker.com/compose/)
- OR [Bun](https://bun.sh/) (for local frontend development)
- OR [Elixir](https://elixir-lang.org/) (for local backend development)

### Quick Start (Docker)

```bash
make up
```

### Local Development

**Frontend:**

```bash
cd frontend
bun install
bun run dev
```

**Backend:**

```bash
cd backend
mix deps.get
mix phx.server
```

## 📡 Endpoints

- **Frontend**: `http://localhost:3000`
- **Backend Health**: `http://localhost:4000/api/health`
- **Table API**: `http://localhost:4000/api/tables/default`
- **Phoenix Socket**: `ws://localhost:4000/socket`
- **Topic**: `table:<table_id>`

## 📜 Commands

- `make build`: Build the Docker images.
- `make down`: Stop the services.
- `make logs`: View service logs.
- `make ps`: View running containers.
