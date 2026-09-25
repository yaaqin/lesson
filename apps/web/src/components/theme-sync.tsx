"use client";

import { useEffect } from "react";
import { useQuery } from "@tanstack/react-query";
import { privateApi } from "@/lib/http";
import { useAuthStore } from "@/store/auth-store";
import type { MeInfo } from "@/hooks/use-curriculum";
import { applyTheme, readStoredTheme, storeTheme } from "@/lib/theme";

// Jaga tema tetap sinkron: ngikutin perubahan tema sistem (kalau pilihannya
// "system"), dan pas login di device baru, pakai pilihan yang kesimpen di akun.
export function ThemeSync() {
  const session = useAuthStore((s) => s.session);
  const meQuery = useQuery({
    queryKey: ["me"],
    queryFn: async () => (await privateApi.get<MeInfo>("/app/me")).data,
    enabled: !!session,
    staleTime: 5_000,
  });
  const serverTheme = meQuery.data?.themePreference;

  useEffect(() => {
    if (serverTheme && serverTheme !== readStoredTheme()) {
      storeTheme(serverTheme);
      applyTheme(serverTheme);
    }
  }, [serverTheme]);

  useEffect(() => {
    applyTheme(readStoredTheme());
    const mq = window.matchMedia("(prefers-color-scheme: dark)");
    const onChange = () => {
      if (readStoredTheme() === "system") applyTheme("system");
    };
    mq.addEventListener("change", onChange);
    return () => mq.removeEventListener("change", onChange);
  }, []);

  return null;
}
