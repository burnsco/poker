# ♠️ Poker Platform

A real-time, full-stack poker platform featuring a high-concurrency Elixir/Phoenix backend and a modern React frontend.

## 🚀 Tech Stack

- **Backend**: [Elixir](https://elixir-lang.org/) & [Phoenix Framework](https://www.phoenixframework.org/) (WebSockets, Channels, OTP)
- **Frontend**: [React 19](https://react.dev/), [Vite](https://vitejs.dev/), [TypeScript](https://www.typescriptlang.org/)
- **Linting/Formatting**: [Biome](https://biomejs.dev/)
- **Infrastructure**: Docker & Docker Compose

## 📂 Project Structure

- `frontend/`: React + Vite client for the poker UI.
- `backend/`: Elixir service hosting table state and real-time game traffic via Phoenix Channels.

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
