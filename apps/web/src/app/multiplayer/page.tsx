"use client";

import Link from "next/link";
import { useRouter } from "next/navigation";
import { useEffect, useState } from "react";
import { useAuthStore } from "@/store/auth-store";
import { useRequireNickname } from "@/hooks/use-profile";
import { BackButton } from "@/components/back-button";
import { PremiumBadge } from "@/components/premium-badge";
import { MODE_INFO, normalizeRoomCode } from "@/lib/multiplayer";

export default function MultiplayerHomePage() {
  const router = useRouter();
  const session = useAuthStore((s) => s.session);
  const hasHydrated = useAuthStore((s) => s.hasHydrated);
  const me = useRequireNickname().data;
  const [code, setCode] = useState("");

  useEffect(() => {
    if (!hasHydrated) return;
    if (session === null) router.replace("/login");
  }, [hasHydrated, session, router]);

  if (!hasHydrated || !session) {
    return (
      <div className="flex min-h-screen items-center justify-center bg-zinc-50 dark:bg-black">
        <p className="text-zinc-500 dark:text-zinc-500">Memuat…</p>
      </div>
    );
  }

  const canJoin = code.length === 6;

  return (
    <div className="flex min-h-screen flex-col bg-zinc-50 font-sans dark:bg-black">
      <header className="flex items-center justify-between px-4 py-5 sm:px-10">
        <BackButton href="/belajar" label="Kembali ke jenjang" />
        <span className="text-lg font-semibold tracking-tight text-black dark:text-zinc-50">Multiplayer</span>
        <span className="w-9" />
      </header>

      <main className="mx-auto flex w-full max-w-md flex-1 flex-col gap-6 px-4 pb-10">
        <form
          onSubmit={(e) => {
            e.preventDefault();
            if (canJoin) router.push(`/multiplayer/${code}`);
          }}
          className="flex flex-col gap-3 rounded-2xl border border-black/[.08] bg-white p-5 dark:border-white/[.145] dark:bg-zinc-900"
        >
          <label htmlFor="room-code" className="font-semibold text-black dark:text-zinc-50">
            Gabung room
          </label>
          <input
            id="room-code"
            value={code}
            onChange={(e) => setCode(normalizeRoomCode(e.target.value))}
            placeholder="KODE ROOM"
            autoComplete="off"
            autoCapitalize="characters"
            inputMode="text"
            className="rounded-xl border-2 border-black/[.08] bg-zinc-50 px-4 py-3 text-center font-mono text-2xl font-semibold tracking-[0.3em] text-black uppercase outline-none focus:border-violet-500 dark:border-white/[.145] dark:bg-black dark:text-zinc-50"
          />
          <button
            type="submit"
            disabled={!canJoin}
            className="rounded-full bg-violet-600 px-6 py-3 text-sm font-semibold text-white transition-colors hover:bg-violet-700 disabled:opacity-40"
          >
            Gabung
          </button>
        </form>

        <div className="flex flex-col gap-3 rounded-2xl border border-black/[.08] bg-white p-5 dark:border-white/[.145] dark:bg-zinc-900">
          <div className="flex items-center justify-between gap-2">
            <span className="font-semibold text-black dark:text-zinc-50">Bikin room</span>
            <PremiumBadge size="xs" />
          </div>
          {me?.isPremium ? (
            <>
              <p className="text-sm text-zinc-500 dark:text-zinc-400">
                Racik soalnya sendiri, bagiin kodenya, terus tantang temen-temenmu.
              </p>
              <Link
                href="/multiplayer/buat"
                className="rounded-full bg-foreground px-6 py-3 text-center text-sm font-semibold text-background"
              >
                Bikin room baru
              </Link>
            </>
          ) : (
            <p className="text-sm text-zinc-500 dark:text-zinc-400">
              Bikin room cuma bisa buat user Premium. Kamu tetap bisa gabung ke room orang lain pakai kodenya.
            </p>
          )}
        </div>

        <div className="flex flex-col gap-3">
          <span className="text-sm font-semibold text-zinc-500 dark:text-zinc-500">Mode permainan</span>
          {Object.values(MODE_INFO).map((mode) => (
            <div
              key={mode.label}
              className="flex gap-3 rounded-2xl border border-black/[.08] bg-white p-4 dark:border-white/[.145] dark:bg-zinc-900"
            >
              <span className="text-2xl">{mode.icon}</span>
              <span className="flex flex-col gap-0.5">
                <span className="font-semibold text-black dark:text-zinc-50">{mode.label}</span>
                <span className="text-sm text-zinc-500 dark:text-zinc-400">{mode.desc}</span>
              </span>
            </div>
          ))}
          <p className="text-xs text-zinc-400 dark:text-zinc-600">
            Multiplayer gak motong nyawa harian. Maksimal 10 pemain per room.
          </p>
        </div>
      </main>
    </div>
  );
}
