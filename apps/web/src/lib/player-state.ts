"use client";

import { useCallback, useSyncExternalStore } from "react";

const STORAGE_KEY = "mathquest_dummy_player_v1";

type PlayerState = {
  lives: number;
  streak: number;
  completedChallengeIds: string[];
};

const DEFAULT_STATE: PlayerState = {
  lives: 3,
  streak: 0,
  completedChallengeIds: [],
};

let cache: PlayerState | null = null;
const listeners = new Set<() => void>();

function readFromStorage(): PlayerState {
  try {
    const raw = window.localStorage.getItem(STORAGE_KEY);
    if (!raw) return DEFAULT_STATE;
    return { ...DEFAULT_STATE, ...JSON.parse(raw) };
  } catch {
    return DEFAULT_STATE;
  }
}

function subscribe(listener: () => void) {
  listeners.add(listener);
  return () => listeners.delete(listener);
}

// Lazy-load sekali dari localStorage saat pertama diakses di client.
function getSnapshot(): PlayerState {
  if (cache === null) {
    cache = readFromStorage();
  }
  return cache;
}

function getServerSnapshot(): PlayerState {
  return DEFAULT_STATE;
}

function writeState(next: PlayerState) {
  cache = next;
  window.localStorage.setItem(STORAGE_KEY, JSON.stringify(next));
  listeners.forEach((listener) => listener());
}

// State dummy di localStorage buat simulasi sebelum backend Go + Postgres siap.
// Aturan asli nyawa/streak (reset harian, cap mingguan) ada di FSD.md 3.6 & 3.7.
export function usePlayerState() {
  const state = useSyncExternalStore(subscribe, getSnapshot, getServerSnapshot);

  const recordAttempt = useCallback((challengeId: string, passed: boolean) => {
    const prev = getSnapshot();
    writeState({
      lives: passed ? prev.lives : Math.max(0, prev.lives - 1),
      streak: prev.streak + 1,
      completedChallengeIds: passed
        ? Array.from(new Set([...prev.completedChallengeIds, challengeId]))
        : prev.completedChallengeIds,
    });
  }, []);

  return { ...state, recordAttempt };
}
