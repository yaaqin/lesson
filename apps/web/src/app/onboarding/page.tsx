"use client";

import { useRouter } from "next/navigation";
import { useEffect } from "react";
import { useAuthStore } from "@/store/auth-store";
import { useMeQuery } from "@/hooks/use-curriculum";
import { ProfileForm } from "@/components/profile-form";

// Pertama kali masuk (mis. baru login Google): wajib bikin nickname dulu
// sebelum bisa main -- lihat useRequireNickname di halaman-halaman lain.
export default function OnboardingPage() {
  const router = useRouter();
  const session = useAuthStore((s) => s.session);
  const hasHydrated = useAuthStore((s) => s.hasHydrated);
  const meQuery = useMeQuery();
  const me = meQuery.data;

  useEffect(() => {
    if (!hasHydrated) return;
    if (session === null) router.replace("/login");
  }, [hasHydrated, session, router]);

  useEffect(() => {
    if (me?.nickname) router.replace("/belajar");
  }, [me, router]);

  if (!hasHydrated || !session || !me || me.nickname) {
    return (
      <div className="flex min-h-screen items-center justify-center bg-zinc-50 dark:bg-black">
        <p className="text-zinc-500 dark:text-zinc-500">Memuat…</p>
      </div>
    );
  }

  return (
    <div className="flex min-h-screen flex-col items-center justify-center gap-6 bg-zinc-50 px-4 py-10 font-sans dark:bg-black">
      <div className="flex w-full max-w-md flex-col gap-6 rounded-2xl border border-black/[.08] bg-white p-6 dark:border-white/[.145] dark:bg-zinc-900">
        <div className="flex flex-col gap-1">
          <h1 className="text-xl font-semibold text-black dark:text-zinc-50">Halo! Bikin profil kamu dulu</h1>
          <p className="text-sm text-zinc-500">
            Nickname ini jadi identitas unik kamu di MathQuest (gak boleh sama kayak user lain). Nanti masih bisa
            diganti di halaman profil.
          </p>
        </div>
        <ProfileForm me={me} submitLabel="Simpan & mulai belajar" onSaved={() => router.replace("/belajar")} />
      </div>
    </div>
  );
}
