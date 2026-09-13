import { create } from "zustand";
import { createJSONStorage, persist } from "zustand/middleware";
import type { OrgRole, PlatformRole } from "@/lib/dummy-accounts";

export type Session =
  | {
      area: "admin";
      email: string;
      displayName: string;
      platformRole: PlatformRole;
      accessToken: string;
      refreshToken: string;
    }
  | {
      area: "org";
      email: string;
      displayName: string;
      orgRole: OrgRole;
      organizationId: string;
      organizationName: string;
    };

type AuthState = {
  session: Session | null;
  // localStorage dibaca async oleh persist middleware — sebelum ini true, `session`
  // masih nilai awal (null) walau sebenarnya ada sesi tersimpan. Halaman yang nge-guard
  // route (redirect ke /login kalau session null) WAJIB nunggu hasHydrated dulu,
  // supaya reload halaman gak keliru nge-bounce user yang sebenarnya masih login.
  hasHydrated: boolean;
  setSession: (session: Session | null) => void;
  setHasHydrated: (value: boolean) => void;
};

// Zustand store diakses juga dari luar React (axios interceptor di lib/http.ts)
// lewat useAuthStore.getState()/.setState(), makanya bukan cuma dipakai via hook.
export const useAuthStore = create<AuthState>()(
  persist(
    (set) => ({
      session: null,
      hasHydrated: false,
      setSession: (session) => set({ session }),
      setHasHydrated: (value) => set({ hasHydrated: value }),
    }),
    {
      name: "mathquest_dashboard_session_v1",
      storage: createJSONStorage(() => localStorage),
      onRehydrateStorage: () => (state) => {
        state?.setHasHydrated(true);
      },
    },
  ),
);
