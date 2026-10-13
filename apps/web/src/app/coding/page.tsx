"use client";

import Link from "next/link";
import { useRouter } from "next/navigation";
import { useEffect } from "react";
import { useAuthStore } from "@/store/auth-store";
import { useRequireNickname } from "@/hooks/use-profile";
import { UserAvatar } from "@/components/user-avatar";
import { BackButton } from "@/components/back-button";
import { clearPostLoginPath, rememberPostLoginPath } from "@/lib/post-login";

// Hub pelajaran Coding Dasar -- dipisah dari /belajar (khusus matematika).
// Kuis = tier "coding" (latihan soal PG kayak jenjang matematika), Course =
// mode belajar materi, konsepnya nyusul.
export default function CodingHubPage() {
  const router = useRouter();
  const session = useAuthStore((s) => s.session);
  const hasHydrated = useAuthStore((s) => s.hasHydrated);
  const me = useRequireNickname().data;

  useEffect(() => {
    if (!hasHydrated) return;
    if (session === null) {
      rememberPostLoginPath("/coding");
      router.replace("/login");
    }
  }, [hasHydrated, session, router]);

  useEffect(() => clearPostLoginPath(), []);

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
        <Link href="/coding" className="text-lg font-semibold tracking-tight text-black dark:text-zinc-50">
          MathQuest
        </Link>
        {me?.nickname && (
          <Link
            href="/profile"
            className="flex items-center gap-2 rounded-full py-1 pr-3 pl-1 transition-colors hover:bg-black/[.04] dark:hover:bg-[#1a1a1a]"
          >
            <UserAvatar avatar={me.avatar} size="sm" />
            <span className="hidden text-sm font-medium text-zinc-700 sm:inline dark:text-zinc-300">@{me.nickname}</span>
          </Link>
        )}
      </header>

      <main className="mx-auto flex w-full max-w-2xl flex-1 flex-col gap-6 px-6 py-8">
        <div className="-ml-2 flex items-center gap-1">
          <BackButton href="/" label="Ganti pelajaran" />
          <div className="flex flex-col gap-1">
            <h1 className="text-2xl font-semibold tracking-tight text-black dark:text-zinc-50">💻 Coding Dasar</h1>
            <p className="text-sm text-zinc-500 dark:text-zinc-500">JavaScript dari nol. Mau belajar pakai cara apa?</p>
          </div>
        </div>

        <Link
          href="/belajar/coding"
          className="flex items-center gap-4 rounded-2xl border-2 border-emerald-500 bg-white px-5 py-5 transition-colors hover:bg-emerald-50 dark:border-emerald-400 dark:bg-zinc-900 dark:hover:bg-emerald-500/10"
        >
          <span className="text-3xl">📝</span>
          <span className="flex flex-1 flex-col gap-0.5">
            <span className="font-semibold text-black dark:text-zinc-50">Kuis</span>
            <span className="text-sm text-zinc-500 dark:text-zinc-400">
              Tebak output kode: variabel, if, loop, array, fungsi. 4 batch + ujian.
            </span>
          </span>
          <span className="text-sm font-medium text-emerald-600 dark:text-emerald-400">Mulai →</span>
        </Link>

        <div className="relative flex items-center gap-4 rounded-2xl border-2 border-dashed border-black/[.08] bg-white/50 px-5 py-5 text-zinc-400 dark:border-white/[.145] dark:bg-zinc-900/40 dark:text-zinc-600">
          <span className="absolute -top-2.5 right-4 rounded-full bg-zinc-200 px-2 py-0.5 text-[10px] font-semibold tracking-wide text-zinc-600 uppercase dark:bg-zinc-700 dark:text-zinc-300">
            Soon
          </span>
          <span className="text-3xl grayscale">📚</span>
          <span className="flex flex-1 flex-col gap-0.5">
            <span className="font-semibold">Course</span>
            <span className="text-sm">Belajar materi step by step. Segera hadir.</span>
          </span>
        </div>
      </main>
    </div>
  );
}
