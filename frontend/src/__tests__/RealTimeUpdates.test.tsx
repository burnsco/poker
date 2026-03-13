import { vi, describe, it, expect, beforeEach } from "vitest";
import { renderWithProviders, screen, waitFor } from "../test/test-utils";
import App from "../App";
import React, { useState, useEffect } from "react";

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

// Create a stateful mock for usePokerTable
let updateCallback: (table: any) => void = () => {};

vi.mock("../hooks/usePokerTable", () => ({
  usePokerTable: (tableId: string) => {
    const [table, setTable] = useState<any>({
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
    });

    useEffect(() => {
      updateCallback = setTable;
    }, []);

    return {
      backendTable: table,
      backendState: "Connected",
      backendHealth: { service: "poker", version: "1.0.0" },
      playerId: "1",
      sendAction: vi.fn(),
      tableNotice: null,
      setTableNotice: vi.fn(),
    };
  },
}));

describe("Real-Time Updates", () => {
  beforeEach(() => {
    mockHash("#/tables/default");
  });

  it("updates table state when receiving a WebSocket event", async () => {
    renderWithProviders(<App />);

    // Wait for initial load
    await waitFor(() => {
      expect(screen.getByText((_, el) => !!el?.textContent?.includes("30"))).toBeInTheDocument();
    });

    // Simulate WebSocket event
    updateCallback({
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
      last_event: "flop_dealt",
      hand_state: {
        status: "in_progress",
        stage: "flop",
        pot: 150,
        community_cards: ["2h", "7d", "Tc"],
        action_log: ["Flop dealt."],
        acting_seat: 1,
        dealer_seat: 1,
        small_blind: 10,
        big_blind: 20,
      },
    });

    await waitFor(() => {
      expect(screen.getByText((_, el) => !!el?.textContent?.includes("150"))).toBeInTheDocument();
      expect(screen.getByText("2h")).toBeInTheDocument();
    });
  });
});
