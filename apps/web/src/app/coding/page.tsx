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
// materi + latihan per module (/coding/course, services/api coursesvc).
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

        <Link
          href="/coding/course"
          className="flex items-center gap-4 rounded-2xl border-2 border-sky-500 bg-white px-5 py-5 transition-colors hover:bg-sky-50 dark:border-sky-400 dark:bg-zinc-900 dark:hover:bg-sky-500/10"
        >
          <span className="text-3xl">📚</span>
          <span className="flex flex-1 flex-col gap-0.5">
            <span className="font-semibold text-black dark:text-zinc-50">Course</span>
            <span className="text-sm text-zinc-500 dark:text-zinc-400">
              Belajar materi step by step: Tipe Variabel → Looping → Asynchronous. Tanpa nyawa.
            </span>
          </span>
          <span className="text-sm font-medium text-sky-600 dark:text-sky-400">Belajar →</span>
        </Link>
      </main>
    </div>
  );
}
