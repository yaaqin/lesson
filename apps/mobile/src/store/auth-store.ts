import AsyncStorage from "@react-native-async-storage/async-storage";
import { create } from "zustand";
import { createJSONStorage, persist } from "zustand/middleware";

export type Session = {
  id: string;
  email: string;
  displayName: string;
  role: "student" | "admin" | "superadmin";
  accessToken: string;
  refreshToken: string;
};

type AuthState = {
  session: Session | null;
  // persist Zustand baca AsyncStorage async -- layar yang nge-guard route
  // (redirect ke /login) WAJIB nunggu ini true dulu sebelum mutusin redirect,
  // sama persis kayak pola hasHydrated di apps/web.
  hasHydrated: boolean;
  setSession: (session: Session | null) => void;
  setHasHydrated: (value: boolean) => void;
};

// Diakses juga dari luar React (axios interceptor di lib/http.ts) lewat
// useAuthStore.getState()/.setState().
export const useAuthStore = create<AuthState>()(
  persist(
    (set) => ({
      session: null,
      hasHydrated: false,
      setSession: (session) => set({ session }),
      setHasHydrated: (value) => set({ hasHydrated: value }),
    }),
    {
      name: "mathquest_mobile_session_v1",
      storage: createJSONStorage(() => AsyncStorage),
      onRehydrateStorage: () => (state) => {
        state?.setHasHydrated(true);
      },
    },
  ),
);
