"use client";

import Link from "next/link";
import { useRouter } from "next/navigation";
import { useEffect } from "react";
import { useAuthStore } from "@/store/auth-store";
import { useLogoutMutation } from "@/hooks/use-auth";
import { useTiersQuery } from "@/hooks/use-curriculum";
import { useRequireNickname } from "@/hooks/use-profile";
import { UserAvatar } from "@/components/user-avatar";
import { InstallAppBanner } from "@/components/install-app";

const TIER_ICON: Record<string, string> = {
  sd: "➕",
  smp: "📐",
  smk: "📊",
  kampus: "🎓",
  umum: "🧩",
};

export default function TierPickerPage() {
  const router = useRouter();
  const session = useAuthStore((s) => s.session);
  const hasHydrated = useAuthStore((s) => s.hasHydrated);
  const logoutMutation = useLogoutMutation();
  const tiersQuery = useTiersQuery();
  const me = useRequireNickname().data;

  useEffect(() => {
    if (!hasHydrated) return;
    if (session === null) {
      router.replace("/login");
    }
  }, [hasHydrated, session, router]);

  if (!hasHydrated || !session) {
    return (
      <div className="flex min-h-screen items-center justify-center bg-zinc-50 dark:bg-black">
        <p className="text-zinc-500 dark:text-zinc-500">Memuat…</p>
      </div>
    );
  }

  return (
    <div className="flex min-h-screen flex-col bg-zinc-50 font-sans dark:bg-black">
      <header className="flex items-center justify-between px-6 py-5 sm:px-10">
        <Link href="/" className="text-lg font-semibold tracking-tight text-black dark:text-zinc-50">
          MathQuest
        </Link>
        <div className="flex items-center gap-2">
          {me?.nickname && (
            <Link
              href="/profile"
              className="flex items-center gap-2 rounded-full py-1 pr-3 pl-1 transition-colors hover:bg-black/[.04] dark:hover:bg-[#1a1a1a]"
            >
              <UserAvatar avatar={me.avatar} size="sm" />
              <span className="hidden text-sm font-medium text-zinc-700 sm:inline dark:text-zinc-300">@{me.nickname}</span>
            </Link>
          )}
          <button
            type="button"
            onClick={async () => {
              await logoutMutation.mutateAsync();
              router.push("/login");
            }}
            className="rounded-full border border-black/[.08] px-4 py-1.5 text-sm font-medium text-zinc-600 transition-colors hover:bg-black/[.04] dark:border-white/[.145] dark:text-zinc-400 dark:hover:bg-[#1a1a1a]"
          >
            Keluar
          </button>
        </div>
      </header>

      <main className="mx-auto flex w-full max-w-2xl flex-1 flex-col gap-6 px-6 py-8">
        <div className="flex flex-col gap-1">
          <h1 className="text-2xl font-semibold tracking-tight text-black dark:text-zinc-50">
            Pilih Jenjang
          </h1>
          <p className="text-sm text-zinc-500 dark:text-zinc-500">
            Halo, {me?.nickname ? `@${me.nickname}` : session.displayName}. Mau latihan yang mana hari ini?
          </p>
        </div>

        <InstallAppBanner />

        <Link
          href="/adventure"
          className="flex items-center gap-4 rounded-2xl border-2 border-blue-600 bg-white px-5 py-5 transition-colors hover:bg-blue-50 dark:border-blue-500 dark:bg-zinc-900 dark:hover:bg-blue-500/10"
        >
          <span className="text-3xl">🗺️</span>
          <span className="flex flex-1 flex-col gap-0.5">
            <span className="font-semibold text-black dark:text-zinc-50">Adventure</span>
            <span className="text-sm text-zinc-500 dark:text-zinc-400">
              4.000 soal campuran SD sampai Kampus, checkpoint tiap 100 soal.
            </span>
          </span>
          <span className="text-sm font-medium text-blue-600 dark:text-blue-400">Main →</span>
        </Link>

        <Link
          href="/multiplayer"
          className="flex items-center gap-4 rounded-2xl border-2 border-violet-500 bg-white px-5 py-5 transition-colors hover:bg-violet-50 dark:border-violet-400 dark:bg-zinc-900 dark:hover:bg-violet-500/10"
        >
          <span className="text-3xl">⚔️</span>
          <span className="flex flex-1 flex-col gap-0.5">
            <span className="font-semibold text-black dark:text-zinc-50">Multiplayer</span>
            <span className="text-sm text-zinc-500 dark:text-zinc-400">
              Main bareng sampai 10 orang, pakai kode room.
            </span>
          </span>
          <span className="text-sm font-medium text-violet-600 dark:text-violet-400">Gabung →</span>
        </Link>

        {tiersQuery.isLoading && (
          <p className="text-sm text-zinc-500 dark:text-zinc-500">Memuat jenjang…</p>
        )}

        <div className="grid grid-cols-1 gap-4 sm:grid-cols-3">
          {tiersQuery.data?.map((tier) => (
            <Link
              key={tier.id}
              href={`/belajar/${tier.code}`}
              className="flex flex-col items-center gap-2 rounded-2xl border border-black/[.08] bg-white px-4 py-8 text-center transition-colors hover:border-blue-400 dark:border-white/[.145] dark:bg-zinc-900"
            >
              <span className="text-3xl">{TIER_ICON[tier.code] ?? "🔢"}</span>
              <span className="font-semibold text-black dark:text-zinc-50">{tier.name}</span>
            </Link>
          ))}
        </div>
      </main>
    </div>
  );
}
