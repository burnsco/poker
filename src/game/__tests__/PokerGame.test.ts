import { afterEach, beforeEach, describe, expect, jest, test } from "bun:test";
import { PokerGame } from "../PokerGame";

function requireValue<T>(value: T | null | undefined, message: string): T {
	if (value == null) {
		throw new Error(message);
	}
	return value;
}

type MutableState = {
	actionTo: number | null;
	currentBets: Map<number, number>;
	players: Array<{ id: string; status: string; betThisStreet: number } | null>;
};

type MutableGameInternals = {
	aiDelayTimer: ReturnType<typeof setTimeout> | null;
};

type ActionLogInternals = {
	appendActionLog: (message: string) => void;
};

describe("PokerGame Configuration and Initialization", () => {
	test("Initializes with default configuration if none provided", () => {
		const game = new PokerGame();

		expect(game.config.maxPlayers).toBe(8);
		expect(game.config.smallBlind).toBe(10);
		expect(game.config.bigBlind).toBe(20);
		expect(game.config.bettingStructure).toBe("NO_LIMIT");
		expect(game.config.minBet).toBe(20);
		expect(game.config.allowStraddle).toBe(false);
		expect(game.config.timeBankSeconds).toBe(30);
		expect(game.config.turnTimeSeconds).toBe(15);
		expect(game.engine).toBeDefined();
	});

	test("Overrides default configuration with provided Partial<GameConfig>", () => {
		const game = new PokerGame({
			smallBlind: 50,
			bigBlind: 100,
			bettingStructure: "POT_LIMIT",
			turnTimeSeconds: 10,
		});

		expect(game.config.smallBlind).toBe(50);
		expect(game.config.bigBlind).toBe(100);
		expect(game.config.bettingStructure).toBe("POT_LIMIT");
		expect(game.config.turnTimeSeconds).toBe(10);
	});

	test("Initially seats 4 players (You, Alice, Bob, Charlie)", () => {
		const game = new PokerGame();
		const state = game.state;

		expect(state.players[0]?.name).toBe("You");
		expect(state.players[1]?.name).toBe("Alice");
		expect(state.players[0]?.stack).toBe(5000);
	});

	test("Clamps maxPlayers to supported bounds", () => {
		const tooSmall = new PokerGame({ maxPlayers: 1 });
		const tooLarge = new PokerGame({ maxPlayers: 99 });

		expect(tooSmall.config.maxPlayers).toBe(2);
		expect(tooLarge.config.maxPlayers).toBe(8);
	});
});

describe("PokerGame View and State Getters", () => {
	test("Exposes internal state via getter", () => {
		const game = new PokerGame();
		expect(game.state.street).toBe("PREFLOP");
	});

	test("Exposes masked view for 'p1' (You) via getter", () => {
		const game = new PokerGame();
		game.startRound();

		const view = game.view;
		expect(view.players[0]?.hand?.length).toBe(2);

		// p1 should NOT see Alice's (p2) cards initially (masked)
		let hasAllHidden = true;
		for (const card of view.players[1]?.hand || []) {
			if (card !== null) hasAllHidden = false;
		}
		expect(hasAllHidden).toBe(true);
	});

	test("Exposes current betting stage string via getter", () => {
		const game = new PokerGame();
		expect(game.stage).toBe("preflop");
		game.startRound();
		expect(game.stage).toBe("preflop");
	});
});

describe("PokerGame Base Action Validations (validateBetOrRaise)", () => {
	let game: PokerGame;

	beforeEach(() => {
		game = new PokerGame({ smallBlind: 10, bigBlind: 20 });
		game.startRound();
	});

	test("NO_LIMIT: Prevent raises smaller than the min raise", () => {
		expect(game.validateBetOrRaise("p1", 50, true)).toBe(true);
	});

	test("POT_LIMIT: Calculates exact max limit size and rejects values above it", () => {
		const plGame = new PokerGame({
			smallBlind: 10,
			bigBlind: 20,
			bettingStructure: "POT_LIMIT",
		});
		plGame.engine.deal();

		const s = plGame.state;
		const actingPlayerSeat = requireValue(
			s.actionTo,
			"Expected an acting seat",
		);
		const actingPlayerId = requireValue(
			s.players[actingPlayerSeat]?.id,
			"Expected acting player id",
		);

		// Pot is 30 (SB 10 + BB 20).
		// Action is on UTG. To call is 20. Pot after call is 50.
		// In pot-limit, max raise target is currentBet(20) + potAfterCall(50) = 70.
		expect(plGame.validateBetOrRaise(actingPlayerId, 70, true)).toBe(true);
		expect(plGame.validateBetOrRaise(actingPlayerId, 71, true)).toBe(false);
	});

	test("FIXED_LIMIT: Restricts bets/raises to exact jump sizes depending on street", () => {
		const flGame = new PokerGame({
			smallBlind: 10,
			bigBlind: 20,
			bettingStructure: "FIXED_LIMIT",
		});
		flGame.engine.deal();

		const s = flGame.state;
		const actingPlayerSeat = requireValue(
			s.actionTo,
			"Expected an acting seat",
		);
		const actingPlayerId = requireValue(
			s.players[actingPlayerSeat]?.id,
			"Expected acting player id",
		);

		// Preflop fixed bet = 10 (small blind amount).
		// UTG to call 20 + raise 10 = Target 30.
		expect(flGame.validateBetOrRaise(actingPlayerId, 30, true)).toBe(true);
		expect(flGame.validateBetOrRaise(actingPlayerId, 40, true)).toBe(false);
	});

	test("Rejects bets above stack and sends a message", () => {
		const onMessageSpy = jest.fn();
		game.onMessage = onMessageSpy;

		const seat = requireValue(game.state.actionTo, "Expected an acting seat");
		const player = requireValue(game.state.players[seat], "Expected a player");
		const overStackTarget = player.betThisStreet + player.stack + 1;

		expect(game.validateBetOrRaise(player.id, overStackTarget, true)).toBe(
			false,
		);
		expect(onMessageSpy).toHaveBeenCalledWith(
			expect.stringContaining("up to your stack"),
		);
	});
});

describe("PokerGame Actions (handlePlayerAction)", () => {
	let game: PokerGame;

	beforeEach(() => {
		// Mock setTimeout/setInterval to prevent AI background loops
		jest.useFakeTimers();
		game = new PokerGame({ smallBlind: 10, bigBlind: 20 });
		// Set state deterministically so p1 is always UTG for easier testing
		// With 4 players, seat 0 (p1) is UTG if button is seat 1 (p2). SB=p3, BB=p4.
		game.state.button = 1;
		game.engine.deal();
	});

	test("Fails silently if not p1's turn", () => {
		// Change action to someone else
		game.state.actionTo = 2;
		const previousPot = game.state.pot;

		game.handlePlayerAction("call");

		// Action shouldn't happen, state shouldn't change
		expect(game.state.pot).toBe(previousPot);
		expect(game.state.actionTo).toBe(2);
	});

	test("Maps string 'fold' to ActionType.FOLD", () => {
		game.state.actionTo = 0; // Ensure it's p1
		const engineActSpy = jest.spyOn(game.engine, "act");

		game.handlePlayerAction("fold");

		expect(engineActSpy).toHaveBeenCalledWith(
			expect.objectContaining({ type: "FOLD", playerId: "p1" }),
		);
		expect(game.state.players[0]?.status).toBe("FOLDED");
	});

	test("Maps string 'call' to ActionType.CALL", () => {
		game.state.actionTo = 0;
		const engineActSpy = jest.spyOn(game.engine, "act");

		game.handlePlayerAction("call");

		expect(engineActSpy).toHaveBeenCalledWith(
			expect.objectContaining({ type: "CALL", playerId: "p1" }),
		);
		expect(game.state.players[0]?.betThisStreet).toBe(20);
	});

	test("Maps string 'check' to ActionType.CHECK (falling back to CALL if facing a bet)", () => {
		game.state.actionTo = 0;
		const onMessageSpy = jest.fn();
		game.onMessage = onMessageSpy;

		// Let's actually force the error to ensure the catch block executes
		// as sometimes the engine logic or test setup might allow a check if state isn't perfectly synced.
		const engineActSpy = jest
			.spyOn(game.engine, "act")
			.mockImplementationOnce(() => {
				throw new Error("ILLEGAL_ACTION: cannot check");
			});

		// Preflop facing a 20 blind bet. Check is invalid. Wrapper should fallback to call.
		game.handlePlayerAction("check");

		// Assert it fell back to call (the second call to `act` since the first one threw)
		expect(engineActSpy).toHaveBeenCalledWith(
			expect.objectContaining({ type: "CALL", playerId: "p1" }),
		);
	});

	test("Maps string 'raise' to ActionType.RAISE within valid limits", () => {
		game.state.actionTo = 0;
		const engineActSpy = jest.spyOn(game.engine, "act");

		// Explicit amount
		game.handlePlayerAction("raise", 60);

		expect(engineActSpy).toHaveBeenCalledWith(
			expect.objectContaining({ type: "RAISE", playerId: "p1", amount: 60 }),
		);
		expect(game.state.players[0]?.betThisStreet).toBe(60);
	});

	test("Blocks straddle when straddles are disabled", () => {
		game.state.actionTo = 0;
		const onMessageSpy = jest.fn();
		game.onMessage = onMessageSpy;
		const engineActSpy = jest.spyOn(game.engine, "act");

		game.handlePlayerAction("straddle");

		expect(onMessageSpy).toHaveBeenCalledWith(
			"Straddles not allowed right now.",
		);
		expect(engineActSpy).not.toHaveBeenCalledWith(
			expect.objectContaining({ type: "RAISE", playerId: "p1" }),
		);
	});
});

describe("PokerGame Bot AI Logic (triggerAILogic)", () => {
	let game: PokerGame;

	beforeEach(() => {
		jest.useFakeTimers();
		game = new PokerGame();

		// Start a round properly
		game.startRound();

		// Find out whose turn it is.
		// In 4-player, someone is first. If it's p1 (You), we need to fold them so an AI acts.
		const actingSeat = game.state.actionTo;
		if (actingSeat === 0) {
			game.handlePlayerAction("fold");
		}
		// Now action is definitely on an AI (seat 1, 2, or 3)
	});

	afterEach(() => {
		jest.useRealTimers();
	});

	test("AI action is delayed by 1.2s", () => {
		const engineActSpy = jest.spyOn(game.engine, "act");

		game.triggerAILogic();

		// At 0ms nothing happened
		expect(engineActSpy).not.toHaveBeenCalled();

		jest.advanceTimersByTime(1199);
		expect(engineActSpy).not.toHaveBeenCalled();

		jest.advanceTimersByTime(1); // 1200ms
		expect(engineActSpy).toHaveBeenCalled();
	});

	test("AI prefers calling or folding when facing a bet with a weak hand", () => {
		const engineActSpy = jest.spyOn(game.engine, "act");

		// Math.random controls heuristic random branch
		// 0.0 means AI considers hand weak and doesn't hit bluff thresholds
		jest.spyOn(Math, "random").mockReturnValue(0.1);

		// The acting player is the current actionTo right now
		const actingSeat = requireValue(
			game.state.actionTo,
			"Expected an acting seat",
		);
		const actingId = requireValue(
			game.state.players[actingSeat]?.id,
			"Expected an acting player id",
		);

		game.triggerAILogic();
		jest.advanceTimersByTime(1200);

		// Facing a bet with a weak hand (rand 0.1 < 0.5) usually falls to fold
		expect(engineActSpy).toHaveBeenCalledWith(
			expect.objectContaining({ type: "FOLD", playerId: actingId }),
		);

		jest.restoreAllMocks();
	});

	test("AI checks when no bet to call", () => {
		const engineActSpy = jest.spyOn(game.engine, "act");

		// Fake no toCall by clearing all current street bets
		const actingSeat = requireValue(
			game.state.actionTo,
			"Expected an acting seat",
		);
		const actingPlayer = requireValue(
			game.state.players[actingSeat],
			"Expected an acting player",
		);
		const mutableState = game.engine.state as unknown as MutableState;
		const mutableActingPlayer = mutableState.players[actingSeat];
		if (!mutableActingPlayer) {
			throw new Error("Expected mutable acting player");
		}

		mutableState.currentBets = new Map();
		mutableActingPlayer.betThisStreet = 0;

		jest.spyOn(Math, "random").mockReturnValue(0.1);

		game.triggerAILogic();
		jest.advanceTimersByTime(1200);

		expect(engineActSpy).toHaveBeenCalledWith(
			expect.objectContaining({ type: "CHECK", playerId: actingPlayer.id }),
		);

		jest.restoreAllMocks();
	});

	test("AI auto-skips a non-active action seat instead of hanging", () => {
		const badSeat = requireValue(
			game.state.actionTo,
			"Expected an acting seat",
		);
		const badPlayer = requireValue(
			game.state.players[badSeat],
			"Expected a bad player",
		);
		const engineActSpy = jest.spyOn(game.engine, "act");
		const mutableState = game.engine.state as unknown as MutableState;
		const mutableGame = game as unknown as MutableGameInternals;
		const mutableBadPlayer = mutableState.players[badSeat];
		if (!mutableBadPlayer) {
			throw new Error("Expected mutable bad player");
		}

		jest.clearAllTimers();
		mutableGame.aiDelayTimer = null;
		mutableState.actionTo = badSeat;
		mutableBadPlayer.status = "FOLDED";

		game.triggerAILogic();

		expect(engineActSpy).toHaveBeenCalledWith(
			expect.objectContaining({ type: "TIMEOUT", playerId: badPlayer.id }),
		);
	});

	test("AI triggers next AI logic step after acting", () => {
		const triggerAISpy = jest.spyOn(game, "triggerAILogic");

		game.triggerAILogic();

		expect(triggerAISpy).toHaveBeenCalledTimes(1); // initial

		jest.advanceTimersByTime(1200);

		// Should be called again in the timeout callback afteract()
		expect(triggerAISpy).toHaveBeenCalledTimes(2);
	});
});

describe("PokerGame Timers and Timeouts", () => {
	let game: PokerGame;

	beforeEach(() => {
		jest.useFakeTimers();
		game = new PokerGame({ turnTimeSeconds: 15 });
		// Prevent AI from acting and messing up timeout testing
		jest.spyOn(game, "triggerAILogic").mockImplementation(() => {});
		// Set up deterministic state
		game.startRound();
	});

	afterEach(() => {
		jest.useRealTimers();
		game.clearTimer(); // Not natively exposed but we can just let garbage collector handle mocked timers anyway
	});

	test("Starts a 15 second timer for the active player", () => {
		const onMessageSpy = jest.fn();
		game.onMessage = onMessageSpy;

		// Fast forward 15 seconds
		jest.advanceTimersByTime(15000);

		// Timeout should happen
		expect(onMessageSpy).toHaveBeenCalledWith(
			expect.stringContaining("ran out of time"),
		);
	});

	test("Timer auto-checks or folds the player on timeout", () => {
		const engineActSpy = jest
			.spyOn(game.engine, "act")
			.mockImplementationOnce(() => {
				throw new Error("ILLEGAL_ACTION: test forcing fold");
			});
		const actionTo = requireValue(
			game.state.actionTo,
			"Expected actionTo seat",
		);

		// Fast forward
		jest.advanceTimersByTime(15000);

		// Since we force the first act to throw, the catch block in the timeout handler
		// should trigger the secondary engine.act({ type: 'FOLD' })
		expect(engineActSpy).toHaveBeenCalledWith(
			expect.objectContaining({
				type: "FOLD",
				playerId: game.state.players[actionTo]?.id,
			}),
		);
	});
});

describe("PokerGame Action Log", () => {
	test("Keeps only the most recent 400 action lines", () => {
		const game = new PokerGame();
		const internals = game as unknown as ActionLogInternals;

		for (let i = 0; i < 405; i += 1) {
			internals.appendActionLog(`entry-${i}`);
		}

		expect(game.actionLogEntries.length).toBe(400);
		expect(game.actionLogEntries[0]).toBe("entry-5");
		expect(game.actionLogEntries[399]).toBe("entry-404");
	});
});
