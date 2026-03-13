import { vi, describe, it, expect, beforeEach } from "vitest";
import { renderWithProviders, screen, waitFor, fireEvent } from "../test/test-utils";
import App from "../App";
import React from "react";

// Mock AuthContext
vi.stubEnv("VITE_BACKEND_URL", "http://localhost:4000");

vi.mock("../contexts/AuthContext", () => ({
  useAuth: () => ({
    user: { id: 1, username: "TestUser", balance: 5000 },
    loading: false,
    authPending: false,
    clearAuthError: vi.fn(),
    login: vi.fn(),
    register: vi.fn(),
    logout: vi.fn(),
  }),
  AuthProvider: ({ children }: { children: React.ReactNode }) => <div>{children}</div>,
}));

// Mock window.location.hash
const mockHash = (hash: string) => {
  Object.defineProperty(window, "location", {
    value: { hash },
    writable: true,
  });
};

// Mock UI modules
vi.mock("../ui/Renderer", () => ({
  Renderer: class {
    update() {}
    showMessage() {}
  },
}));

vi.mock("../ui/PokerSoundEngine", () => ({
  PokerSoundEngine: class {
    play() {}
    dispose() {}
  },
}));

// Create a mutable mock for usePokerTable state
const mockSendAction = vi.fn();
let mockTableState: any = {
  table_id: "default",
  players: [
    {
      seat: 1,
      name: "TestUser",
      player_id: "1",
      stack: 5000,
      status: "ACTIVE",
      is_bot: false,
      hole_cards: ["Ah", "Ad"],
    },
  ],
  game_state: "hand_in_progress",
  hand_number: 1,
  connected_clients: 1,
  last_event: "ready",
  hand_state: {
    status: "in_progress",
    stage: "preflop",
    pot: 30,
    current_bet: 0,
    minimum_raise: 20,
    community_cards: [],
    action_log: [],
    acting_seat: 1,
    dealer_seat: 1,
    small_blind_seat: 1,
    big_blind_seat: 5,
    small_blind: 10,
    big_blind: 20,
    winner_seats: [],
    winner_amounts: {},
    hand_result: null,
  },
};

vi.mock("../hooks/usePokerTable", () => ({
  usePokerTable: () => ({
    backendTable: mockTableState,
    backendState: "Connected",
    backendHealth: { service: "poker", version: "1.0.0" },
    playerId: "1",
    sendAction: mockSendAction,
    tableNotice: null,
    setTableNotice: vi.fn(),
  }),
}));

describe("Player Actions", () => {
  beforeEach(() => {
    mockHash("#/tables/default");
    vi.clearAllMocks();
  });

  it("enables actions when it is the players turn", async () => {
    // Override acting_seat to hero
    mockTableState.hand_state.acting_seat = 1;

    renderWithProviders(<App />);

    // In my refactored App.tsx, I should check if "Your turn" message is rendered by React
    // Currently it's still partly in Renderer. But let's check if buttons are enabled.
    await waitFor(() => {
      const foldBtn = screen.getByRole("button", { name: /Fold/i });
      expect(foldBtn).not.toBeDisabled();
    });

    const checkBtn = screen.getByRole("button", { name: /Check/i });
    const betBtn = screen.getByRole("button", { name: /Bet/i });

    expect(checkBtn).not.toBeDisabled();
    expect(betBtn).not.toBeDisabled();
  });

  it("calls sendAction when Fold is clicked", async () => {
    mockTableState.hand_state.acting_seat = 1;

    renderWithProviders(<App />);

    await waitFor(() => screen.getByRole("button", { name: /Fold/i }));
    fireEvent.click(screen.getByRole("button", { name: /Fold/i }));

    expect(mockSendAction).toHaveBeenCalledWith("fold", undefined);
  });
});
