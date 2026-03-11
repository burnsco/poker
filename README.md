# Poker Platform

A real-time poker platform with a Phoenix backend, React frontend, and CFR-trained bot personalities. The project combines table state, live play, and training tooling in one repo.

## Repo layout

- `frontend/`: React + Vite client
- `backend/`: Elixir/Phoenix realtime backend
- `training/`: OpenSpiel-based bot training scripts

## Highlights

- Realtime table play over Phoenix channels
- Multiple bot styles derived from CFR training output
- Docker-first local development flow

## Quick start

```bash
make up
```

Local development:

```bash
cd frontend && bun install && bun run dev
cd backend && mix deps.get && mix phx.server
```

Retrain bots:

```bash
make training-retrain
make training-retrain-holdem
```
