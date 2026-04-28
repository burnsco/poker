# OpenSpiel poker training

This directory sets up [OpenSpiel](https://github.com/google-deepmind/open_spiel) so you can train poker bots with **CFR** (Counterfactual Regret Minimization) and related algorithms. The trained policies can inform or replace the heuristic bots in the main app.

## Quick start

```bash
# From repo root or training/
uv sync

# Train on Leduc poker (default, ~10k iterations)
uv run python train_cfr.py

# Train on Kuhn poker (very fast), export policy
uv run python train_cfr.py --game kuhn_poker --iterations 1000 -o policies/kuhn.json

# More iterations and exploitability logging
uv run python train_cfr.py --game leduc_poker --iterations 50000 --exploitability-every 5000 -o policies/leduc.json
```

## Supported games

| Game              | Description                          | Use case                    |
|------------------|--------------------------------------|-----------------------------|
| **kuhn_poker**   | 2 players, 3 cards, 1 round          | Sanity check, very fast     |
| **leduc_poker**  | 2 players, 6 cards, 2 rounds         | Standard benchmark          |
| **universal_poker** | Configurable (GAMEDEF)            | Custom small poker variants |

Your main app is **Texas Hold'em** (preflop, flop, turn, river, community cards). OpenSpiel does not include full no-limit Hold'em. You can still:

- Use **Leduc** (and similar small games) to train equilibrium-style play and concepts (betting rounds, bluffing, hand strength).
- Use the exported **policy JSON** as a reference (e.g. action probabilities per information state) and approximate similar behavior in your Elixir bots, or feed it into a larger pipeline later.

## Options

- `--game` – `kuhn_poker`, `leduc_poker`, or `universal_poker`
- `--iterations` – CFR iterations (default: 10000)
- `--output` / `-o` – path to write policy JSON
- `--exploitability-every N` – print exploitability every N iterations (0 = off)

## Output

With `-o path.json` you get:

- **exploitability** – distance from Nash (lower is better)
- **policy** – map of information-state string → action id → probability

You can use this from Elixir by loading the JSON and mapping your game’s states (or abstractions) to the nearest OpenSpiel infoset and sampling actions by those probabilities.

## Install from source (if pip wheel fails)

The `open_spiel` pip package is a binary wheel (e.g. x86_64). If it’s not available for your platform:

```bash
# Ubuntu/Debian
sudo apt-get install cmake clang python3-dev
uv pip install --no-binary=:all: open_spiel
```

See [OpenSpiel installation](https://openspiel.readthedocs.io/en/latest/install.html) for full instructions.
