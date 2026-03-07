#!/usr/bin/env python3
"""
Aggregate CFR policy into a compact hand-strength lookup table
that can be embedded in the Elixir poker bot.

Supported policy formats:
  leduc_poker     : 6-card simplified poker (J/Q/K x 2 suits)
  holdem          : 10-card Hold'em-like (T/J/Q/K/A x 2 suits, universal_poker)

Output: a JSON table keyed by (strength, round, has_pair, facing_bet)
where probabilities are {fold, passive, aggressive}.

Usage:
  python aggregate_cfr.py                           # uses policies/leduc.json
  python aggregate_cfr.py --policy policies/holdem.json --output policies/cfr_table.json
"""

import argparse
import json
import re
from pathlib import Path
from collections import defaultdict
from itertools import combinations

DEFAULT_POLICY_PATH = Path(__file__).parent / "policies" / "leduc.json"
DEFAULT_OUTPUT_PATH = Path(__file__).parent / "policies" / "cfr_table.json"

# ---------------------------------------------------------------------------
# Leduc helpers
# ---------------------------------------------------------------------------

def leduc_card_rank(card_idx: int) -> int:
    """Map card 0-5 to rank 0=Jack, 1=Queen, 2=King."""
    return card_idx // 2


def leduc_hand_bucket(private: int) -> str:
    rank = leduc_card_rank(private)
    return {0: "weak", 1: "medium", 2: "strong"}[rank]


def parse_leduc_key(key: str) -> dict | None:
    """
    Parse Leduc infoset key into structured fields.
    Example key: '[Round 1][Private: 4][Round1: 2 ]'
    Returns None if key can't be parsed.
    """
    m = re.search(r'\[Private: (\d+)\]', key)
    if not m:
        return None
    private = int(m.group(1))

    m = re.search(r'\[Round (\d+)\]', key)
    if not m:
        return None
    round_num = int(m.group(1))

    public = None
    m = re.search(r'\[Public: (\d+)\]', key)
    if m:
        public = int(m.group(1))

    # Extract current round's action sequence
    if round_num == 1:
        m = re.search(r'\[Round1: ([^\]]*)\]', key)
    else:
        m = re.search(r'\[Round2: ([^\]]*)\]', key)

    if not m:
        return None
    action_seq = m.group(1).strip()

    # Facing a bet = last action in current round was 2
    actions = [int(a) for a in action_seq.split() if a.strip()]
    facing_bet = bool(actions and actions[-1] == 2)

    # Has pair = private and public share the same rank (round 2 only)
    has_pair = False
    if round_num == 2 and public is not None:
        has_pair = leduc_card_rank(private) == leduc_card_rank(public)

    return {
        "round": round_num,
        "has_pair": has_pair,
        "facing_bet": facing_bet,
        "strength": leduc_hand_bucket(private),
    }


# ---------------------------------------------------------------------------
# Hold'em (universal_poker) helpers
# ---------------------------------------------------------------------------

# Rank chars in ascending order (handles both numeric '234' and face 'TJQKA')
_RANK_ORDER = "23456789TJQKA"


def holdem_card_rank(card_str: str) -> int:
    """Return numeric rank for a card string like '2d', '4c', 'Kh'."""
    return _RANK_ORDER.index(card_str[0])


def parse_packed_cards(s: str) -> list[str]:
    """Parse packed card string '2d2c' -> ['2d', '2c']. Each card is 2 chars."""
    return [s[i:i+2] for i in range(0, len(s), 2)] if s.strip() else []


def holdem_pair_rank(cards: list[str]) -> int | None:
    """Return rank of pair if cards contain a pair, else None."""
    ranks = [holdem_card_rank(c) for c in cards]
    seen: dict[int, int] = {}
    for r in ranks:
        seen[r] = seen.get(r, 0) + 1
    for r, cnt in seen.items():
        if cnt >= 2:
            return r
    return None


def holdem_hand_strength(hole: list[str], board: list[str]) -> tuple[str, bool]:
    """
    Evaluate hand strength bucket and whether a pair exists.

    3-rank game (2=low, 3=mid, 4=high in this training deck):
      strong  – pair of top rank, or trips
      medium  – any other pair, or high-card top rank
      weak    – everything else
    """
    all_cards = hole + board
    top_rank = max(holdem_card_rank(c) for c in all_cards) if all_cards else 0

    if board:
        pair_rank = holdem_pair_rank(all_cards)
        has_pair = pair_rank is not None
        if has_pair:
            strength = "strong" if pair_rank >= top_rank else "medium"
        else:
            best_hole = max(holdem_card_rank(c) for c in hole)
            strength = "medium" if best_hole >= top_rank else "weak"
    else:
        has_pair = holdem_pair_rank(hole) is not None
        ranks = sorted([holdem_card_rank(c) for c in hole], reverse=True)
        if has_pair:
            strength = "strong" if ranks[0] >= top_rank else "medium"
        else:
            strength = "medium" if ranks[0] >= top_rank else "weak"

    return strength, has_pair


def parse_holdem_key(key: str) -> dict | None:
    """
    Parse universal_poker infoset key into structured fields.

    Actual OpenSpiel key format:
      '[Round 0][Player: 0][Pot: 4][Money: N N][Private: 2d2c][Public: ][Sequences: ]'
      '[Round 1][Player: 0][Pot: 4][Money: N N][Private: 2d2c][Public: 4c][Sequences: cc|]'
      '[Round 1][Player: 0][Pot: 12][Money: N N][Private: 2d2c][Public: 4c][Sequences: cc|cr]'
    Cards are packed (no spaces): '2d2c' = two cards '2d' and '2c'.
    Sequences are pipe-separated per round: 'cc|cr' means round0='cc', round1='cr'.
    """
    m = re.search(r'\[Round (\d+)\]', key)
    if not m:
        return None
    round_num = int(m.group(1))  # 0=preflop, 1=postflop

    m = re.search(r'\[Private: ([^\]]*)\]', key)
    if not m:
        return None
    hole_cards = parse_packed_cards(m.group(1).strip())
    if not hole_cards:
        return None

    m = re.search(r'\[Public: ([^\]]*)\]', key)
    board_raw = m.group(1).strip() if m else ""
    board_cards = parse_packed_cards(board_raw)

    # '[Sequences: cc|cr]' — rounds separated by '|'
    m = re.search(r'\[Sequences: ([^\]]*)\]', key)
    if not m:
        return None
    parts = m.group(1).split('|')
    current_seq = parts[round_num] if round_num < len(parts) else ""

    # Facing a bet = last action char was 'r' (raise/bet)
    facing_bet = bool(current_seq and current_seq[-1] == 'r')

    round_label = "preflop" if round_num == 0 else "postflop"
    strength, has_pair = holdem_hand_strength(hole_cards, board_cards)

    return {
        "round": round_num,
        "round_label": round_label,
        "has_pair": has_pair,
        "facing_bet": facing_bet,
        "strength": strength,
    }


# ---------------------------------------------------------------------------
# Action probability extraction (game-format-aware)
# ---------------------------------------------------------------------------

def extract_action_probs(action_probs: dict, game: str) -> tuple[float, float, float]:
    """
    Return (fold, passive, aggressive) from an action_probs dict.

    Leduc/Kuhn:      keys are string ints "0","1","2"  (0=fold, 1=passive, 2=aggr)
    universal_poker: keys are string ints "0","1","2"  (0=fold, 1=call, 2=raise)
                     but action set may be smaller (no fold when check is available)
    """
    if game in ("leduc_poker", "kuhn_poker", "holdem", "universal_poker"):
        fold_p = action_probs.get("0", 0.0)
        passive_p = action_probs.get("1", 0.0)
        aggressive_p = action_probs.get("2", 0.0)
        return fold_p, passive_p, aggressive_p
    return 0.0, 0.5, 0.5


# ---------------------------------------------------------------------------
# Generic aggregation
# ---------------------------------------------------------------------------

def aggregate(policy: dict, game: str) -> dict:
    """
    Aggregate policy into:
      key: (strength, round_label, has_pair, facing_bet)
      value: averaged {fold, passive, aggressive} probabilities
    """
    buckets: dict = defaultdict(lambda: {"fold": [], "passive": [], "aggressive": []})

    parse_key = parse_leduc_key if game in ("leduc_poker", "kuhn_poker") else parse_holdem_key

    for infoset_key, action_probs in policy.items():
        parsed = parse_key(infoset_key)
        if parsed is None:
            continue

        round_label = parsed.get("round_label") or ("preflop" if parsed["round"] == 1 else "postflop")
        bucket_key = (
            parsed["strength"],
            round_label,
            parsed["has_pair"],
            parsed["facing_bet"],
        )

        fold_p, passive_p, aggressive_p = extract_action_probs(action_probs, game)
        buckets[bucket_key]["fold"].append(fold_p)
        buckets[bucket_key]["passive"].append(passive_p)
        buckets[bucket_key]["aggressive"].append(aggressive_p)

    result = {}
    for (strength, round_label, has_pair, facing_bet), vals in sorted(buckets.items()):
        def avg(lst):
            return round(sum(lst) / len(lst), 4) if lst else 0.0

        key = f"{strength}_{round_label}_{'pair' if has_pair else 'nopair'}_{'bet' if facing_bet else 'nobet'}"
        result[key] = {
            "fold": avg(vals["fold"]),
            "passive": avg(vals["passive"]),
            "aggressive": avg(vals["aggressive"]),
            "samples": len(vals["fold"]),
        }

    return result


def main():
    parser = argparse.ArgumentParser(description="Aggregate CFR policy into hand-strength lookup table")
    parser.add_argument("--policy", "-p", type=Path, default=DEFAULT_POLICY_PATH,
                        help="Path to policy JSON produced by train_cfr.py")
    parser.add_argument("--output", "-o", type=Path, default=DEFAULT_OUTPUT_PATH,
                        help="Output path for aggregated cfr_table.json")
    args = parser.parse_args()

    with open(args.policy) as f:
        data = json.load(f)

    game = data.get("game", "leduc_poker")
    policy = data["policy"]
    print(f"Loaded {len(policy)} infosets from {args.policy} (game={game})")

    table = aggregate(policy, game)
    print(f"\nAggregated into {len(table)} buckets:\n")

    for key, v in sorted(table.items()):
        print(f"  {key:45s}  fold={v['fold']:.3f}  pass={v['passive']:.3f}  aggr={v['aggressive']:.3f}  (n={v['samples']})")

    out = {
        "source": f"{game}_cfr",
        "exploitability": data.get("exploitability"),
        "iterations": data.get("iterations"),
        "table": table,
    }
    args.output.parent.mkdir(parents=True, exist_ok=True)
    with open(args.output, "w") as f:
        json.dump(out, f, indent=2)
    print(f"\nWritten to {args.output}")


if __name__ == "__main__":
    main()
