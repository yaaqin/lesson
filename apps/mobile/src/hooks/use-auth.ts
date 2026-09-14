import { useMutation } from "@tanstack/react-query";
import { publicApi, privateApi } from "@/lib/http";
import { useAuthStore, type Session } from "@/store/auth-store";

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

type LoginVars = { identifier: string; password: string };

export function useLoginMutation() {
  const setSession = useAuthStore((s) => s.setSession);

  return useMutation({
    mutationFn: async ({ identifier, password }: LoginVars) => {
      const { data } = await publicApi.post<ApiTokenPair>("/auth/login", {
        identifier,
        password,
      });
      const session: Session = {
        id: data.user.id,
        email: data.user.email,
        displayName: data.user.displayName,
        role: data.user.role,
        accessToken: data.accessToken,
        refreshToken: data.refreshToken,
      };
      setSession(session);
      return session;
    },
  });
}

export function useLogoutMutation() {
  const setSession = useAuthStore((s) => s.setSession);

  return useMutation({
    mutationFn: async () => {
      await privateApi.post("/auth/logout").catch(() => {
        // best-effort — sesi lokal tetap dihapus walau request ini gagal/backend mati
      });
      setSession(null);
    },
  });
}
