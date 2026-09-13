import { useMutation } from "@tanstack/react-query";
import axios from "axios";
import { publicApi, privateApi } from "@/lib/http";
import { useAuthStore, type Session } from "@/store/auth-store";
import { DUMMY_ACCOUNTS } from "@/lib/dummy-accounts";
import { getRegistrations } from "@/store/org-registrations-store";

type ApiUser = {
  id: string;
  email: string;
  displayName: string;
  role: "student" | "admin" | "superadmin";
};

type ApiTokenPair = {
  accessToken: string;
  refreshToken: string;
  user: ApiUser;
};

export type LoginResult =
  | { ok: true; session: Session }
  | { ok: false; reason: "invalid" | "pending" | "rejected" };

type LoginVars = { identifier: string; password: string };

// Login coba backend asli (akun admin platform) dulu lewat publicApi + TanStack mutation,
// baru fallback ke akun dummy organisasi (localStorage) kalau gak cocok — backend org
// belum dibangun. Ini juga yang bikin login page dapet isPending/error dari React Query.
export function useLoginMutation() {
  const setSession = useAuthStore((s) => s.setSession);

  return useMutation<LoginResult, Error, LoginVars>({
    mutationFn: async ({ identifier, password }) => {
      const normalizedEmail = identifier.trim().toLowerCase();

      try {
        const { data } = await publicApi.post<ApiTokenPair>("/auth/login", {
          identifier: normalizedEmail,
          password,
        });
        const session: Session = {
          area: "admin",
          email: data.user.email,
          displayName: data.user.displayName,
          platformRole: data.user.role === "student" ? "admin" : data.user.role,
          accessToken: data.accessToken,
          refreshToken: data.refreshToken,
        };
        setSession(session);
        return { ok: true, session };
      } catch (err) {
        // 401 (salah password) atau backend belum jalan (network error) -> coba jalur dummy.
        if (axios.isAxiosError(err) && err.response && err.response.status !== 401) {
          throw err;
        }
      }

      const seedMatch = DUMMY_ACCOUNTS.find(
        (acc) => acc.email.toLowerCase() === normalizedEmail && acc.password === password,
      );
      if (seedMatch) {
        const session: Session = {
          area: "org",
          email: seedMatch.email,
          displayName: seedMatch.displayName,
          orgRole: seedMatch.orgRole,
          organizationId: seedMatch.organizationId,
          organizationName: seedMatch.organizationName,
        };
        setSession(session);
        return { ok: true, session };
      }

      const registration = getRegistrations().find(
        (r) => r.ownerEmail.toLowerCase() === normalizedEmail && r.password === password,
      );
      if (!registration) return { ok: false, reason: "invalid" };
      if (registration.status === "pending") return { ok: false, reason: "pending" };
      if (registration.status === "rejected") return { ok: false, reason: "rejected" };

      const session: Session = {
        area: "org",
        email: registration.ownerEmail,
        displayName: registration.ownerName,
        orgRole: "owner",
        organizationId: registration.id,
        organizationName: registration.organizationName,
      };
      setSession(session);
      return { ok: true, session };
    },
  });
}

export function useLogoutMutation() {
  const setSession = useAuthStore((s) => s.setSession);

  return useMutation({
    mutationFn: async () => {
      const session = useAuthStore.getState().session;
      if (session?.area === "admin") {
        // request ini butuh access token dari sesi yang masih aktif, makanya
        // dipanggil SEBELUM setSession(null) — interceptor privateApi baca token
        // dari store saat ini juga.
        await privateApi.post("/auth/logout").catch(() => {
          // best-effort — sesi lokal tetap dihapus walau request ini gagal/backend mati
        });
      }
      setSession(null);
    },
  });
}
