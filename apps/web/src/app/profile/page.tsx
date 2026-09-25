"use client";

import Link from "next/link";
import { useRouter } from "next/navigation";
import { useEffect } from "react";
import { useAuthStore } from "@/store/auth-store";
import { useRequireNickname } from "@/hooks/use-profile";
import { ProfileForm } from "@/components/profile-form";
import { InstallAppButton } from "@/components/install-app";

export default function ProfilePage() {
  const router = useRouter();
  const session = useAuthStore((s) => s.session);
  const hasHydrated = useAuthStore((s) => s.hasHydrated);
  const meQuery = useRequireNickname();
  const me = meQuery.data;

  useEffect(() => {
    if (!hasHydrated) return;
    if (session === null) router.replace("/login");
  }, [hasHydrated, session, router]);

  if (!hasHydrated || !session || !me || !me.nickname) {
    return (
      <div className="flex min-h-screen items-center justify-center bg-zinc-50 dark:bg-black">
        <p className="text-zinc-500 dark:text-zinc-500">Memuat…</p>
      </div>
    );
  }

  return (
    <div className="flex min-h-screen flex-col bg-zinc-50 font-sans dark:bg-black">
      <header className="flex items-center justify-between px-4 py-5 sm:px-10">
        <Link
          href="/belajar"
          className="text-sm font-medium text-zinc-500 hover:text-zinc-700 dark:text-zinc-500 dark:hover:text-zinc-300"
        >
          ← Kembali
        </Link>
        <span className="text-lg font-semibold tracking-tight text-black dark:text-zinc-50">Profil</span>
        <span className="w-16" />
      </header>

      <main className="mx-auto flex w-full max-w-md flex-1 flex-col gap-6 px-4 pb-10">
        <div className="rounded-2xl border border-black/[.08] bg-white p-6 dark:border-white/[.145] dark:bg-zinc-900">
          <ProfileForm me={me} submitLabel="Simpan perubahan" />
        </div>

        <div className="grid grid-cols-3 gap-3 text-center">
          <Stat label="Streak" value={`🔥 ${me.currentStreak}`} />
          <Stat label="Streak terlama" value={`${me.longestStreak}`} />
          <Stat label="Nyawa" value={`❤️ ${me.livesRemaining}`} />
        </div>

        <InstallAppButton />
      </main>
    </div>
  );
}

function Stat({ label, value }: { label: string; value: string }) {
  return (
    <div className="flex flex-col gap-1 rounded-2xl border border-black/[.08] bg-white px-3 py-4 dark:border-white/[.145] dark:bg-zinc-900">
      <span className="text-lg font-semibold text-black dark:text-zinc-50">{value}</span>
      <span className="text-xs text-zinc-500">{label}</span>
    </div>
  );
}
