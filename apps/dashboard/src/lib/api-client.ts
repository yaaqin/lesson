const API_BASE_URL = process.env.NEXT_PUBLIC_API_BASE_URL ?? "http://localhost:9801/api/v1";

export type ApiUser = {
  id: string;
  email: string;
  displayName: string;
  role: "student" | "admin" | "superadmin";
};

export type ApiTokenPair = {
  accessToken: string;
  refreshToken: string;
  user: ApiUser;
};

export type ApiLoginResult =
  | { ok: true; data: ApiTokenPair }
  | { ok: false; reason: "invalid_credentials" | "network_error" | "unknown" };

export async function apiLogin(identifier: string, password: string): Promise<ApiLoginResult> {
  try {
    const res = await fetch(`${API_BASE_URL}/auth/login`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ identifier, password }),
    });

    if (res.status === 401) {
      return { ok: false, reason: "invalid_credentials" };
    }
    if (!res.ok) {
      return { ok: false, reason: "unknown" };
    }

    const data = (await res.json()) as ApiTokenPair;
    return { ok: true, data };
  } catch {
    // API belum jalan / gak keraih dari browser
    return { ok: false, reason: "network_error" };
  }
}

export function apiLogout(accessToken: string): void {
  fetch(`${API_BASE_URL}/auth/logout`, {
    method: "POST",
    headers: { Authorization: `Bearer ${accessToken}` },
  }).catch(() => {
    // best-effort — sesi lokal tetap dihapus walau request ini gagal
  });
}
