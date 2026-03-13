import { vi, describe, it, expect, beforeEach } from "vitest";
import { renderWithProviders, screen, waitFor } from "../test/test-utils";
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
          status: "READY",
          is_bot: false,
          hole_cards: ["Ah", "Kd"],
          contributed_this_hand: 100,
        },
        {
          seat: 2,
          name: "Alice",
          player_id: "2",
          stack: 4500,
          status: "READY",
          is_bot: true,
          hole_cards: ["As", "Kh"],
          contributed_this_hand: 100,
        },
      ],
      game_state: "waiting_for_hand",
      hand_number: 1,
      connected_clients: 2,
      last_event: "hand_complete",
      hand_state: {
        status: "complete",
        stage: "showdown",
        pot: 200,
        current_bet: 0,
        minimum_raise: 20,
        community_cards: ["2h", "3d", "4c", "5s", "Jd"],
        action_log: ["Showdown.", "TestUser shows Ah Kd", "Alice shows As Kh", "Split pot."],
        acting_seat: null,
        dealer_seat: 1,
        small_blind: 10,
        big_blind: 20,
        winner_seats: [1, 2],
        winner_amounts: { "1": 100, "2": 100 },
        hand_result: {
          heading: "Split pot between TestUser and Alice",
          lines: ["Split pot of 200 awarded."],
          hero_outcome: "split",
        },
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

describe("Poker Edge Cases", () => {
  beforeEach(() => {
    mockHash("#/tables/default");
  });

  it("renders split pot winners correctly", async () => {
    renderWithProviders(<App />);

    await waitFor(() => {
      expect(screen.getByText(/Split pot between TestUser and Alice/)).toBeInTheDocument();
    });

    const winnerLabels = screen.getAllByText("Winner");
    expect(winnerLabels.length).toBe(2);
  });
});
