import axios, { type InternalAxiosRequestConfig } from "axios";
import { useAuthStore, type Session } from "@/store/auth-store";

const API_BASE_URL = process.env.NEXT_PUBLIC_API_BASE_URL ?? "http://localhost:9801/api/v1";

// publicApi: endpoint yang gak butuh login (login, register, refresh).
export const publicApi = axios.create({ baseURL: API_BASE_URL });

// privateApi: endpoint yang butuh access token. Token dipasang otomatis lewat
// interceptor, dan kalau backend balikin 401 (access token kedaluwarsa),
// interceptor nyoba refresh sekali lalu ulang request aslinya.
export const privateApi = axios.create({ baseURL: API_BASE_URL });

privateApi.interceptors.request.use((config) => {
  const session = useAuthStore.getState().session;
  if (session?.area === "admin") {
    config.headers.set("Authorization", `Bearer ${session.accessToken}`);
  }
  return config;
});

type RetryableConfig = InternalAxiosRequestConfig & { _retried?: boolean };

// Beberapa request bisa kena 401 bersamaan (mis. beberapa query jalan paralel) —
// refreshPromise dipakai bareng biar cuma ada 1 call /auth/refresh yang jalan.
let refreshPromise: Promise<string | null> | null = null;

async function refreshAccessToken(): Promise<string | null> {
  const session = useAuthStore.getState().session;
  if (session?.area !== "admin") return null;

  try {
    const { data } = await publicApi.post<{ accessToken: string; refreshToken: string }>(
      "/auth/refresh",
      { refreshToken: session.refreshToken },
    );
    const nextSession: Session = {
      ...session,
      accessToken: data.accessToken,
      refreshToken: data.refreshToken,
    };
    useAuthStore.getState().setSession(nextSession);
    return data.accessToken;
  } catch {
    // refresh token juga invalid (kedaluwarsa / kepakai sesi lain) -> paksa logout
    useAuthStore.getState().setSession(null);
    return null;
  }
}

privateApi.interceptors.response.use(
  (response) => response,
  async (error) => {
    const config = error.config as RetryableConfig | undefined;
    if (error.response?.status !== 401 || !config || config._retried) {
      return Promise.reject(error);
    }
    config._retried = true;

    if (!refreshPromise) {
      refreshPromise = refreshAccessToken().finally(() => {
        refreshPromise = null;
      });
    }

    const newAccessToken = await refreshPromise;
    if (!newAccessToken) {
      return Promise.reject(error);
    }

    config.headers.set("Authorization", `Bearer ${newAccessToken}`);
    return privateApi(config);
  },
);
