import { create } from "zustand";
import { createJSONStorage, persist } from "zustand/middleware";

type PuzzlePrefsState = {
  keyboardPosition: "left" | "right";
  setKeyboardPosition: (position: "left" | "right") => void;
};

// Preferensi tampilan keypad puzzle (kotak addition_grid) -- posisi kiri/kanan
// disimpan per browser, gak perlu ke server.
export const usePuzzlePrefsStore = create<PuzzlePrefsState>()(
  persist(
    (set) => ({
      keyboardPosition: "right",
      setKeyboardPosition: (keyboardPosition) => set({ keyboardPosition }),
    }),
    {
      name: "mathquest_web_puzzle_prefs_v1",
      storage: createJSONStorage(() => localStorage),
    },
  ),
);
