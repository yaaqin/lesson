"use client";

import { useParams, useRouter } from "next/navigation";
import { useEffect } from "react";
import { useAuthStore } from "@/store/auth-store";
import { useRequireNickname } from "@/hooks/use-profile";
import {
  useTierLeaderboardQuery,
  type TierLeaderboard,
  type TierLeaderboardEntry,
} from "@/hooks/use-curriculum";
import { UserAvatar } from "@/components/user-avatar";
import { BackButton } from "@/components/back-button";

const MEDAL: Record<number, string> = { 1: "🥇", 2: "🥈", 3: "🥉" };

export default function TierRankingPage() {
  const router = useRouter();
  const params = useParams<{ tierCode: string }>();
  const session = useAuthStore((s) => s.session);
  const hasHydrated = useAuthStore((s) => s.hasHydrated);
  useRequireNickname();

  const leaderboardQuery = useTierLeaderboardQuery(params.tierCode);

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
  const title = board ? `Ranking ${board.tierName}` : "Ranking";

  return (
    <div className="flex min-h-screen flex-col bg-zinc-50 font-sans dark:bg-black">
      <header className="flex items-center justify-between px-4 py-5 sm:px-10">
        <BackButton href={`/belajar/${params.tierCode}`} label="Kembali ke jenjang" />
        <span className="text-lg font-semibold tracking-tight text-black dark:text-zinc-50">{title}</span>
        <span className="w-9" />
      </header>

      <main className="mx-auto flex w-full max-w-2xl flex-1 flex-col gap-6 px-4 pb-10 sm:px-6">
        {leaderboardQuery.isLoading && <p className="text-sm text-zinc-500">Memuat ranking…</p>}
        {leaderboardQuery.isError && (
          <p className="text-sm text-red-500">Ranking buat jenjang ini belum ada atau gagal dimuat.</p>
        )}

        {board && (
          <>
            <RulesCard board={board} />
            {board.entries.length === 0 ? (
              <p className="text-sm text-zinc-500">Belum ada yang ngumpulin poin di jenjang ini. Jadilah yang pertama!</p>
            ) : (
              <ol className="flex flex-col gap-2">
                {board.entries.map((e) => (
                  <EntryRow key={e.rank} entry={e} />
                ))}
              </ol>
            )}
            {board.me && (
              <div className="flex flex-col gap-2">
                <h2 className="text-sm font-semibold text-black dark:text-zinc-50">Posisi kamu</h2>
                <ol>
                  <EntryRow entry={board.me} />
                </ol>
              </div>
            )}
          </>
        )}
      </main>
    </div>
  );
}

function RulesCard({ board }: { board: TierLeaderboard }) {
  return (
    <div className="flex flex-col gap-2 rounded-2xl border border-black/[.08] bg-white p-5 text-sm leading-relaxed text-zinc-600 dark:border-white/[.145] dark:bg-zinc-900 dark:text-zinc-400">
      <span className="text-xs font-medium tracking-wide text-zinc-500 uppercase">
        Top {board.entries.length} dari {board.totalRanked.toLocaleString("id-ID")} murid {board.tierName}
      </span>
      <ul className="flex flex-col gap-1">
        <li>⭐ Tiap jawaban benar = 1 poin. Yang dihitung skor terbaikmu di tiap challenge.</li>
        <li>🏁 Ujian cuma nambah poin kalau kamu lulus.</li>
        <li>⏱️ Kalau poinnya sama, yang nyampe duluan ada di atas.</li>
      </ul>
    </div>
  );
}

function EntryRow({ entry }: { entry: TierLeaderboardEntry }) {
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
          {entry.challengesCompleted} challenge
          {entry.examsPassed > 0 && ` · ${entry.examsPassed} ujian lulus`}
        </span>
      </div>
      <div className="flex shrink-0 flex-col items-end">
        <span className="font-semibold text-black dark:text-zinc-50">{entry.points.toLocaleString("id-ID")}</span>
        <span className="text-xs text-zinc-500">poin</span>
      </div>
    </li>
  );
}
