"use client";

import { useEffect } from "react";
import { useRouter } from "next/navigation";
import { useAuthStore } from "@/store/auth-store";
import { privateApi } from "@/lib/http";
import { takePostLoginPath } from "@/lib/post-login";

// Backend (routes_auth.go handleGoogleCallback) redirect ke sini bawa
// accessToken/refreshToken di URL FRAGMENT (bukan query string) -- fragment
// gak pernah kekirim ke server, jadi token gak nyangkut di access log manapun.
export default function GoogleCallbackPage() {
  const router = useRouter();
  const setSession = useAuthStore((s) => s.setSession);

  useEffect(() => {
    const params = new URLSearchParams(window.location.hash.slice(1));
    const accessToken = params.get("accessToken");
    const refreshToken = params.get("refreshToken");

    if (!accessToken || !refreshToken) {
      router.replace("/login?error=google_failed");
      return;
    }

    // set token dulu (profil sementara kosong) biar privateApi bisa nempelin
    // Authorization header pas manggil /app/me di bawah.
    setSession({ id: "", email: "", displayName: "", role: "student", accessToken, refreshToken });

    privateApi
      .get("/app/me")
      .then(({ data }) => {
        setSession({
          id: data.id,
          email: data.email,
          displayName: data.displayName,
          role: data.role,
          accessToken,
          refreshToken,
        });
        router.replace(data.nickname ? takePostLoginPath("/belajar") : "/onboarding");
      })
      .catch(() => {
        setSession(null);
        router.replace("/login?error=google_failed");
      });
  }, [router, setSession]);

  return (
    <div className="flex min-h-screen items-center justify-center bg-zinc-50 dark:bg-black">
      <p className="text-zinc-500 dark:text-zinc-500">Menyelesaikan login…</p>
    </div>
  );
}
