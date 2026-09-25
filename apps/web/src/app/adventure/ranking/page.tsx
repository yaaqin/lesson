"use client";

import Link from "next/link";
import { useRouter } from "next/navigation";
import { useEffect } from "react";
import { useAuthStore } from "@/store/auth-store";
import { useRequireNickname } from "@/hooks/use-profile";
import {
  useAdventureLeaderboardQuery,
  type AdventureLeaderboard,
  type AdventureLeaderboardEntry,
} from "@/hooks/use-adventure";
import { UserAvatar } from "@/components/user-avatar";
import { formatDuration, formatTimePercent } from "@/lib/duration";

const MEDAL: Record<number, string> = { 1: "🥇", 2: "🥈", 3: "🥉" };

export default function AdventureRankingPage() {
  const router = useRouter();
  const session = useAuthStore((s) => s.session);
  const hasHydrated = useAuthStore((s) => s.hasHydrated);
  useRequireNickname();

  const leaderboardQuery = useAdventureLeaderboardQuery();

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

  const board = leaderboardQuery.data;

  return (
    <div className="flex min-h-screen flex-col bg-zinc-50 font-sans dark:bg-black">
      <header className="flex items-center justify-between px-4 py-5 sm:px-10">
        <Link
          href="/adventure"
          className="text-sm font-medium text-zinc-500 hover:text-zinc-700 dark:text-zinc-500 dark:hover:text-zinc-300"
        >
          ← Adventure
        </Link>
        <span className="text-lg font-semibold tracking-tight text-black dark:text-zinc-50">Ranking</span>
        <span className="w-20" />
      </header>

      <main className="mx-auto flex w-full max-w-2xl flex-1 flex-col gap-6 px-4 pb-10 sm:px-6">
        {leaderboardQuery.isLoading && <p className="text-sm text-zinc-500">Memuat ranking…</p>}
        {leaderboardQuery.isError && <p className="text-sm text-red-500">Gagal memuat ranking.</p>}

        {board && (
          <>
            <RulesCard board={board} />
            {board.entries.length === 0 ? (
              <p className="text-sm text-zinc-500">Belum ada yang lulus checkpoint. Jadilah yang pertama!</p>
            ) : (
              <ol className="flex flex-col gap-2">
                {board.entries.map((e) => (
                  <EntryRow key={e.rank} entry={e} totalCheckpoints={board.totalCheckpoints} />
                ))}
              </ol>
            )}
            {board.me && (
              <div className="flex flex-col gap-2">
                <h2 className="text-sm font-semibold text-black dark:text-zinc-50">Posisi kamu</h2>
                <ol>
                  <EntryRow entry={board.me} totalCheckpoints={board.totalCheckpoints} />
                </ol>
              </div>
            )}
          </>
        )}
      </main>
    </div>
  );
}

function RulesCard({ board }: { board: AdventureLeaderboard }) {
  return (
    <div className="flex flex-col gap-2 rounded-2xl border border-black/[.08] bg-white p-5 text-sm leading-relaxed text-zinc-600 dark:border-white/[.145] dark:bg-zinc-900 dark:text-zinc-400">
      <span className="text-xs font-medium tracking-wide text-zinc-500 uppercase">
        Top {board.entries.length} dari {board.totalRanked.toLocaleString("id-ID")} petualang
      </span>
      <ul className="flex flex-col gap-1">
        <li>🗺️ Diurutin dari checkpoint terjauh yang udah lulus.</li>
        <li>
          ⚡ Kalau checkpoint-nya sama, diadu persentase waktu: waktu yang kamu pakai dibagi total waktu yang
          disediain soal-soalnya. Makin kecil makin bagus.
        </li>
        <li>❤️ Tiap sisa jatah salah yang gak kepake motong waktumu {board.bonusSecondsPerFail} detik.</li>
      </ul>
    </div>
  );
}

function EntryRow({ entry, totalCheckpoints }: { entry: AdventureLeaderboardEntry; totalCheckpoints: number }) {
  return (
    <li
      className={`flex items-center gap-3 rounded-2xl border bg-white p-3 sm:p-4 dark:bg-zinc-900 ${
        entry.isMe
          ? "border-blue-500 ring-1 ring-blue-500 dark:border-blue-400 dark:ring-blue-400"
          : "border-black/[.08] dark:border-white/[.145]"
      }`}
    >
      <span className="w-9 shrink-0 text-center text-base font-semibold text-zinc-500">
        {MEDAL[entry.rank] ?? `#${entry.rank}`}
      </span>
      <UserAvatar avatar={entry.avatar} size="sm" />
      <div className="flex min-w-0 flex-1 flex-col">
        <span className="truncate font-semibold text-black dark:text-zinc-50">
          {entry.nickname}
          {entry.isMe && <span className="font-normal text-blue-600 dark:text-blue-400"> (kamu)</span>}
        </span>
        <span className="text-xs text-zinc-500">
          {entry.completed ? "Tamat 🏆" : `Checkpoint ${entry.checkpointsCleared}/${totalCheckpoints}`} ·{" "}
          {entry.questionsCleared.toLocaleString("id-ID")} soal
        </span>
      </div>
      <div
        className="flex shrink-0 flex-col items-end"
        title={`${formatDuration(entry.adjustedSeconds)} dari ${formatDuration(entry.timeAllowedSeconds)} yang disediain (udah dipotong bonus sisa jatah)`}
      >
        <span className="font-semibold text-black dark:text-zinc-50">{formatTimePercent(entry.timePercent)}</span>
        <span className="text-xs text-zinc-500">{formatDuration(entry.adjustedSeconds)}</span>
      </div>
    </li>
  );
}
