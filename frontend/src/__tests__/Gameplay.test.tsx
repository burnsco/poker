import { vi, describe, it, expect, beforeEach } from "vitest";
import { renderWithProviders, screen, waitFor } from "../test/test-utils";
import App from "../App";
import React from "react";

// Mock AuthContext to avoid real API calls and provide a default user
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

// Mock usePokerTable
vi.mock("../hooks/usePokerTable", () => ({
  usePokerTable: (tableId: string) => ({
    backendTable: {
      table_id: tableId,
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
        {
          seat: 5,
          name: "Opponent",
          player_id: "2",
          stack: 4500,
          status: "ACTIVE",
          is_bot: false,
          hole_cards: ["Kh", "Kd"],
        },
      ],
      game_state: "hand_in_progress",
      hand_number: 1,
      connected_clients: 2,
      last_event: "flop_dealt",
      hand_state: {
        status: "in_progress",
        stage: "flop",
        pot: 100,
        community_cards: ["2h", "7d", "Tc"],
        action_log: ["Hand 1 started.", "Flop dealt."],
        acting_seat: 1,
        dealer_seat: 1,
        small_blind_seat: 1,
        big_blind_seat: 5,
        small_blind: 10,
        big_blind: 20,
        current_bet: 0,
        minimum_raise: 20,
        winner_seats: [],
        winner_amounts: {},
        hand_result: null,
      },
    },
    backendState: "Connected",
    backendHealth: { service: "poker", version: "1.0.0" },
    playerId: "1",
    sendAction: vi.fn(),
    tableNotice: null,
    setTableNotice: vi.fn(),
  }),
}));

describe("Gameplay UI", () => {
  beforeEach(() => {
    mockHash("#/tables/default");
  });

  it("renders players in their correct seats", async () => {
    renderWithProviders(<App />);

    // Wait for the players to be rendered in the DOM
    await waitFor(() => {
      expect(screen.getByText("TestUser")).toBeInTheDocument();
      expect(screen.getByText("Opponent")).toBeInTheDocument();
      expect(screen.getByText((_, el) => Boolean(el?.textContent === "5,000"))).toBeInTheDocument();
      expect(screen.getByText((_, el) => Boolean(el?.textContent === "4,500"))).toBeInTheDocument();
    });
  });
});
