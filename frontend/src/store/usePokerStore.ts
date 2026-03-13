import { create } from "zustand";
import type { BackendTable, BackendHealth } from "../types/backend";

interface PokerState {
  table: BackendTable | null;
  health: BackendHealth | null;
  connectionState: string;

  // Actions
  setTable: (table: BackendTable | null) => void;
  setHealth: (health: BackendHealth | null) => void;
  setConnectionState: (state: string) => void;
  reset: () => void;
}

const initialState = {
  table: null,
  health: null,
  connectionState: "Initializing...",
};

export const usePokerStore = create<PokerState>((set) => ({
  ...initialState,

  setTable: (table) => set({ table }),
  setHealth: (health) => set({ health }),
  setConnectionState: (connectionState) => set({ connectionState }),
  reset: () => set(initialState),
}));
