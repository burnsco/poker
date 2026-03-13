import { http, HttpResponse } from "msw";

export const handlers = [
  http.get("*/api/health", () => {
    return HttpResponse.json({ status: "ok", version: "1.0.0" });
  }),

  http.get("*/api/tables", () => {
    return HttpResponse.json({ data: ["default"] });
  }),

  http.get("*/api/tables/:tableId", ({ params }) => {
    const { tableId } = params;
    return HttpResponse.json({
      table_id: tableId,
      players: [],
      game_state: "waiting_for_players",
      hand_number: 0,
      connected_clients: 0,
      last_event: "initialized",
      hand_state: {
        status: "waiting",
        stage: "preflop",
        hand_number: 0,
        pot: 0,
        current_bet: 0,
        minimum_raise: 20,
        acting_seat: null,
        dealer_seat: 1,
        small_blind_seat: 1,
        big_blind_seat: 2,
        community_cards: [],
        action_log: [],
        action_log_seq: 0,
        last_action: null,
        acted_seats: [],
        winner_seats: [],
        winner_amounts: {},
        hand_result: null,
      },
    });
  }),
];
