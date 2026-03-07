#!/usr/bin/env python3
"""
Train a poker bot using OpenSpiel's CFR (Counterfactual Regret Minimization).

Supported games:
  - kuhn_poker   : 2-player, 3 cards, 1 round (fast, good for testing)
  - leduc_poker  : 2-player, 6 cards, 2 rounds (common benchmark)
  - holdem       : 2-player, 10 cards (T-A × 2 suits), 3 rounds, 2 hole cards
  - universal_poker : configurable via GAMEDEF (see --gamedef or code)

Usage:
  python train_cfr.py --game kuhn_poker --iterations 1000
  python train_cfr.py --game leduc_poker --iterations 10000 --output policy.json
  python train_cfr.py --game holdem --iterations 10000 --output policies/holdem.json
"""

import argparse
import json
import sys
from pathlib import Path

try:
    import pyspiel
    from open_spiel.python.algorithms import cfr
    from open_spiel.python.algorithms import exploitability
except ImportError as e:
    print("OpenSpiel not installed. Run: pip install -r requirements.txt", file=sys.stderr)
    raise SystemExit(1) from e


def load_game(name: str, players: int = 2, gamedef: str | None = None):
    """Load an OpenSpiel game by name."""
    if name == "kuhn_poker":
        return pyspiel.load_game("kuhn_poker", {"players": players})
    if name == "leduc_poker":
        return pyspiel.load_game("leduc_poker", {"players": players})
    if name == "holdem":
        # Hold'em-like: 3 ranks (J,Q,K) x 2 suits = 6 cards (same deck size as Leduc),
        # but 2 hole cards + 1 board card across 2 rounds.
        # Key difference from Leduc: 2 hole cards enables pocket pairs preflop.
        # Same card count keeps the game tree tractable (~minutes to train).
        return pyspiel.load_game("universal_poker", {
            "betting": "limit",
            "numPlayers": 2,
            "numRounds": 2,
            "blind": "1 2",
            "raiseSize": "2 4",
            "firstPlayer": "1 1",
            "maxRaises": "2 2",
            "numSuits": 2,
            "numRanks": 3,
            "numHoleCards": 2,
            "numBoardCards": "0 1",
        })
    if name == "universal_poker":
        if gamedef is None:
            raise ValueError("universal_poker requires a --gamedef argument")
        return pyspiel.load_game("universal_poker", {"gamedef": gamedef})
    raise ValueError(f"Unknown game: {name}. Use kuhn_poker, leduc_poker, holdem, or universal_poker.")


def policy_to_serializable(policy):
    """Extract (infoset_key -> action_probabilities) for export."""
    # TabularPolicy.to_dict() returns {infostate_key: [(action_id, prob), ...]}
    raw = policy.to_dict()
    return {k: dict(actions_probs) for k, actions_probs in raw.items()}


def main():
    parser = argparse.ArgumentParser(description="Train poker bot with CFR")
    parser.add_argument("--game", default="leduc_poker", choices=["kuhn_poker", "leduc_poker", "holdem", "universal_poker"])
    parser.add_argument("--players", type=int, default=2)
    parser.add_argument("--iterations", type=int, default=10_000)
    parser.add_argument("--output", "-o", type=Path, default=None, help="Write policy JSON here")
    parser.add_argument("--exploitability-every", type=int, default=0, help="Print exploitability every N iterations (0=off)")
    args = parser.parse_args()

    game = load_game(args.game, players=args.players)
    cfr_solver = cfr.CFRPlusSolver(game)

    print(f"Training CFR on {args.game} for {args.iterations} iterations...")
    for i in range(args.iterations):
        cfr_solver.evaluate_and_update_policy()
        if (i + 1) % 1000 == 0:
            print(f"  iter {i + 1} / {args.iterations}", flush=True)
        if args.exploitability_every and (i + 1) % args.exploitability_every == 0:
            conv = exploitability.exploitability(game, cfr_solver.average_policy())
            print(f"    exploitability = {conv:.6f}", flush=True)

    avg_policy = cfr_solver.average_policy()
    conv = exploitability.exploitability(game, avg_policy)
    print(f"Done. Final exploitability = {conv:.6f}")

    if args.output:
        data = {
            "game": args.game,
            "iterations": args.iterations,
            "exploitability": conv,
            "policy": policy_to_serializable(avg_policy),
        }
        args.output.parent.mkdir(parents=True, exist_ok=True)
        with open(args.output, "w") as f:
            json.dump(data, f, indent=2)
        print(f"Policy written to {args.output}")

    return 0


if __name__ == "__main__":
    raise SystemExit(main())
