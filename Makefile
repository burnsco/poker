COMPOSE := docker compose

.PHONY: up down build logs ps deploy deploy-down frontend-shell backend-shell frontend-install frontend-lint frontend-test frontend-build frontend-check backend-setup backend-compile backend-test backend-check check training-venv training-train training-train-leduc training-train-leduc-50k training-aggregate training-retrain training-train-holdem training-aggregate-holdem training-retrain-holdem

up:
	$(COMPOSE) up --build

down:
	$(COMPOSE) down

deploy:
	$(COMPOSE) -f docker-compose.deploy.yml up --build -d

deploy-down:
	$(COMPOSE) -f docker-compose.deploy.yml down

build:
	$(COMPOSE) build

logs:
	$(COMPOSE) logs -f

ps:
	$(COMPOSE) ps

frontend-shell:
	$(COMPOSE) exec frontend bash

backend-shell:
	$(COMPOSE) exec backend bash

frontend-install:
	cd frontend && bun install

frontend-lint:
	$(COMPOSE) run --rm frontend bun run lint

frontend-test:
	$(COMPOSE) run --rm frontend bun run test

frontend-build:
	$(COMPOSE) run --rm frontend bun run build

frontend-check:
	$(COMPOSE) run --rm frontend bun run lint
	$(COMPOSE) run --rm frontend bun run test
	$(COMPOSE) run --rm frontend bun run build

backend-setup:
	$(COMPOSE) run --rm backend mix deps.get

backend-compile:
	$(COMPOSE) run --rm backend mix compile

backend-test:
	$(COMPOSE) run --rm -e MIX_ENV=test backend mix test

backend-check:
	$(COMPOSE) run --rm -e MIX_ENV=test backend mix compile --warnings-as-errors
	$(COMPOSE) run --rm -e MIX_ENV=test backend mix test

# OpenSpiel training (run from repo root; requires Python 3 and training/.venv)
training-venv:
	cd training && python3 -m venv .venv && .venv/bin/pip install -r requirements.txt

training-train:
	cd training && .venv/bin/python train_cfr.py

training-train-leduc:
	cd training && .venv/bin/python train_cfr.py --game leduc_poker --iterations 10000 -o policies/leduc.json

training-train-leduc-50k:
	cd training && .venv/bin/python train_cfr.py --game leduc_poker --iterations 50000 --exploitability-every 5000 -o policies/leduc.json

training-aggregate:
	cd training && .venv/bin/python aggregate_cfr.py

training-train-holdem:
	cd training && .venv/bin/python train_cfr.py --game holdem --iterations 10000 --exploitability-every 2000 -o policies/holdem.json

training-aggregate-holdem:
	cd training && .venv/bin/python aggregate_cfr.py --policy policies/holdem.json --output policies/cfr_table_holdem.json

# Full retrain + re-aggregate pipeline (run after changing iterations or game)
training-retrain: training-train-leduc-50k training-aggregate

training-retrain-holdem: training-train-holdem training-aggregate-holdem

check: frontend-check backend-check
