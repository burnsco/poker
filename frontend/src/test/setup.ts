import { vi, beforeAll, afterEach, afterAll } from "vitest";
import { server } from "./server";
import "@testing-library/jest-dom";
import { cleanup } from "@testing-library/react";
import { usePokerStore } from "../store/usePokerStore";

// Mock sound assets globally
vi.mock("../assets/sounds/the-sound-of-card-cards-being-laid-out-to-play-poker.mp3?import", () => ({
  default: "mock-sound",
}));
vi.mock("../assets/sounds/poker-chips-stacking-shuffling-01.mp3?import", () => ({
  default: "mock-sound",
}));
vi.mock("../assets/sounds/cards-dealing-01.mp3?import", () => ({ default: "mock-sound" }));
vi.mock("../assets/sounds/poker-chips-bet-01.mp3?import", () => ({ default: "mock-sound" }));

// Establish API mocking before all tests.
beforeAll(() => {
  server.listen({ onUnhandledRequest: "error" });
});

vi.stubEnv("VITE_BACKEND_URL", "http://localhost:4000");

// Reset any request handlers that we may add during the tests,
// so they don't affect other tests.
afterEach(() => {
  cleanup();
  server.resetHandlers();
  usePokerStore.getState().reset();
});

// Clean up after the tests are finished.
afterAll(() => server.close());
