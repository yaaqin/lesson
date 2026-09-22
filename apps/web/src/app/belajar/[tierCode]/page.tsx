"use client";

import Link from "next/link";
import { useParams, useRouter } from "next/navigation";
import { useEffect } from "react";
import { useAuthStore } from "@/store/auth-store";
import {
  useBatchesQuery,
  useCategoriesQuery,
  useChallengesForBatchesQuery,
  useMeQuery,
  useTiersQuery,
} from "@/hooks/use-curriculum";

export default function TierBatchPage() {
  const router = useRouter();
  const params = useParams<{ tierCode: string }>();
  const session = useAuthStore((s) => s.session);
  const hasHydrated = useAuthStore((s) => s.hasHydrated);

  useEffect(() => {
    if (!hasHydrated) return;
    if (session === null) router.replace("/login");
  }, [hasHydrated, session, router]);

  const meQuery = useMeQuery();
  const tiersQuery = useTiersQuery();
  const tier = tiersQuery.data?.find((t) => t.code === params.tierCode);
  const usesBatch = tier?.usesBatch ?? true;

  const batchesQuery = useBatchesQuery(params.tierCode);
  const batches = usesBatch ? batchesQuery.data : undefined;
  const challengesQuery = useChallengesForBatchesQuery(batches);

  const categoriesQuery = useCategoriesQuery(!usesBatch ? params.tierCode : "");

  if (!hasHydrated || !session) {
    return (
      <div className="flex min-h-screen items-center justify-center bg-zinc-50 dark:bg-black">
        <p className="text-zinc-500 dark:text-zinc-500">Memuat…</p>
      </div>
    );
  }

  const lives = meQuery.data?.livesRemaining ?? 0;
  const streak = meQuery.data?.currentStreak ?? 0;

  return (
    <div className="flex min-h-screen flex-col bg-zinc-50 font-sans dark:bg-black">
      <header className="flex items-center justify-between px-6 py-5 sm:px-10">
        <Link href="/belajar" className="text-lg font-semibold tracking-tight text-black dark:text-zinc-50">
          MathQuest
        </Link>
        <div className="flex items-center gap-4 text-sm">
          <span className="flex items-center gap-1">
            {Array.from({ length: 3 }).map((_, i) => (
              <span key={i} className={i < lives ? "text-red-500" : "text-zinc-300 dark:text-zinc-700"}>
                ❤️
              </span>
            ))}
          </span>
          <span className="flex items-center gap-1 font-medium text-orange-500">🔥 {streak}</span>
        </div>
      </header>

      <main className="mx-auto flex w-full max-w-3xl flex-1 flex-col gap-8 px-6 py-8">
        <div className="flex flex-col gap-1">
          <Link href="/belajar" className="text-sm font-medium text-blue-600 dark:text-blue-400">
            ← Ganti jenjang
          </Link>
          <h1 className="text-2xl font-semibold tracking-tight text-black dark:text-zinc-50">
            {tier?.name ?? params.tierCode.toUpperCase()}
          </h1>
        </div>

        {usesBatch ? (
          <>
            {challengesQuery.isLoading && (
              <p className="text-sm text-zinc-500 dark:text-zinc-500">Memuat challenge…</p>
            )}

            <div className="flex flex-col gap-3">
              {challengesQuery.data?.map((challenge, index) => (
                <div key={challenge.id} className="flex flex-col gap-3">
                  {(index === 0 ||
                    challenge.batchName !== challengesQuery.data[index - 1].batchName) && (
                    <h2 className="mt-2 text-xs font-semibold tracking-wide text-zinc-500 uppercase first:mt-0 dark:text-zinc-500">
                      {challenge.batchName}
                    </h2>
                  )}
                  <Link
                    href={`/belajar/${params.tierCode}/${challenge.id}`}
                    className="flex items-center gap-4 rounded-2xl border border-black/[.08] bg-white px-5 py-4 transition-colors hover:border-blue-400 dark:border-white/[.145] dark:bg-zinc-900"
                  >
                    <span
                      className={`flex h-9 w-9 shrink-0 items-center justify-center rounded-full text-sm font-semibold text-white ${
                        challenge.isExam ? "bg-amber-500" : "bg-blue-600 dark:bg-blue-500"
                      }`}
                    >
                      {challenge.isExam ? "🏁" : index + 1}
                    </span>
                    <div className="flex flex-1 flex-col">
                      <span className="flex items-center gap-2 font-semibold text-black dark:text-zinc-50">
                        {challenge.name}
                        {challenge.hasEssay && (
                          <span className="rounded-full bg-purple-100 px-2 py-0.5 text-[10px] font-semibold text-purple-700 dark:bg-purple-500/10 dark:text-purple-400">
                            ✏️ Ada Essay
                          </span>
                        )}
                      </span>
                      <span className="text-xs text-zinc-500 dark:text-zinc-500">
                        {challenge.questionCountRequired} soal · lulus minimal{" "}
                        {challenge.passThresholdPercent}% ·{" "}
                        {challenge.isExam
                          ? `${challenge.timeLimitSeconds}s total`
                          : `${challenge.timeLimitSeconds}s / soal`}
                      </span>
                    </div>
                    {challenge.completed ? (
                      <span className="shrink-0 rounded-full bg-green-100 px-3 py-1 text-xs font-medium text-green-700 dark:bg-green-500/10 dark:text-green-400">
                        Selesai ✓
                      </span>
                    ) : (
                      <span className="shrink-0 text-sm font-medium text-blue-600 dark:text-blue-400">
                        Mulai →
                      </span>
                    )}
                  </Link>
                </div>
              ))}
            </div>
          </>
        ) : (
          <>
            {categoriesQuery.isLoading && (
              <p className="text-sm text-zinc-500 dark:text-zinc-500">Memuat kategori…</p>
            )}

            <div className="grid grid-cols-1 gap-3 sm:grid-cols-2">
              {categoriesQuery.data?.map((category) => (
                <Link
                  key={category.id}
                  href={`/belajar/${params.tierCode}/kategori/${category.id}`}
                  className="flex flex-col gap-1 rounded-2xl border border-black/[.08] bg-white px-5 py-4 transition-colors hover:border-blue-400 dark:border-white/[.145] dark:bg-zinc-900"
                >
                  <span className="font-semibold text-black dark:text-zinc-50">{category.name}</span>
                  <span className="text-sm font-medium text-blue-600 dark:text-blue-400">Lihat level →</span>
                </Link>
              ))}
            </div>
          </>
        )}
      </main>
    </div>
  );
}
