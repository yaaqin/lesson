"use client";

import Link from "next/link";
import { useRouter } from "next/navigation";
import { useEffect, useState } from "react";
import { useAuthStore } from "@/store/auth-store";
import { useLogoutMutation } from "@/hooks/use-auth";
import {
  useAdminCurriculumQuery,
  useUpdateChallengeTimingMutation,
  type AdminChallenge,
} from "@/hooks/use-admin-curriculum";
import { PLATFORM_ROLE_LABEL } from "@/lib/dummy-accounts";

export default function AdminKurikulumPage() {
  const router = useRouter();
  const session = useAuthStore((s) => s.session);
  const hasHydrated = useAuthStore((s) => s.hasHydrated);
  const logoutMutation = useLogoutMutation();
  const curriculumQuery = useAdminCurriculumQuery();

  useEffect(() => {
    if (!hasHydrated) return;
    if (session === null) {
      router.replace("/login");
      return;
    }
    if (session.area !== "admin") {
      router.replace("/org");
    }
  }, [hasHydrated, session, router]);

  if (!hasHydrated || !session || session.area !== "admin") {
    return (
      <div className="flex min-h-screen items-center justify-center bg-zinc-50 dark:bg-black">
        <p className="text-zinc-500 dark:text-zinc-500">Memuat…</p>
      </div>
    );
  }

  return (
    <div className="flex min-h-screen flex-col bg-zinc-50 font-sans dark:bg-black">
      <header className="flex items-center justify-between px-6 py-5 sm:px-10">
        <div className="flex flex-col">
          <span className="text-lg font-semibold tracking-tight text-black dark:text-zinc-50">
            MathQuest Admin
          </span>
          <span className="text-xs text-zinc-500 dark:text-zinc-500">
            {session.displayName} · {PLATFORM_ROLE_LABEL[session.platformRole]}
          </span>
        </div>
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
      </header>

      <main className="mx-auto flex w-full max-w-3xl flex-1 flex-col gap-6 px-6 py-8">
        <div className="flex flex-col gap-1">
          <Link href="/admin" className="text-sm font-medium text-blue-600 dark:text-blue-400">
            ← Kembali
          </Link>
          <h1 className="text-2xl font-semibold tracking-tight text-black dark:text-zinc-50">
            Kurikulum
          </h1>
          <p className="text-sm text-zinc-500 dark:text-zinc-500">
            Waktu challenge biasa = detik per soal. Waktu ujian (🏁) = total detik buat semua
            soal.
          </p>
        </div>

        {curriculumQuery.isLoading && (
          <p className="text-sm text-zinc-500 dark:text-zinc-500">Memuat kurikulum…</p>
        )}

        <div className="flex flex-col gap-6">
          {curriculumQuery.data?.map((tier) => (
            <div key={tier.id} className="flex flex-col gap-3">
              <h2 className="text-sm font-semibold tracking-wide text-zinc-500 uppercase dark:text-zinc-500">
                {tier.name}
              </h2>
              {tier.batches.length === 0 && (
                <Link
                  href={`/admin/kurikulum/kategori/${tier.code}`}
                  className="flex flex-col gap-1 rounded-2xl border border-black/[.08] bg-white p-5 transition-colors hover:border-blue-400 dark:border-white/[.145] dark:bg-zinc-900"
                >
                  <span className="text-sm font-medium text-black dark:text-zinc-50">
                    Tier ini pakai kategori, bukan batch →
                  </span>
                  <span className="text-xs text-zinc-500 dark:text-zinc-500">
                    Lihat kategori & soal-nya di halaman terpisah.
                  </span>
                </Link>
              )}
              {tier.batches.map((batch) => (
                <div
                  key={batch.id}
                  className="flex flex-col gap-2 rounded-2xl border border-black/[.08] bg-white p-5 dark:border-white/[.145] dark:bg-zinc-900"
                >
                  <span className="text-sm font-medium text-zinc-600 dark:text-zinc-400">
                    {batch.name}
                  </span>
                  <div className="flex flex-col divide-y divide-black/[.06] dark:divide-white/[.08]">
                    {batch.challenges.map((challenge) => (
                      <ChallengeTimingRow key={challenge.id} challenge={challenge} />
                    ))}
                  </div>
                </div>
              ))}
            </div>
          ))}
        </div>
      </main>
    </div>
  );
}

function ChallengeTimingRow({ challenge }: { challenge: AdminChallenge }) {
  const [value, setValue] = useState(String(challenge.timeLimitSeconds));
  const [saved, setSaved] = useState(false);
  const updateMutation = useUpdateChallengeTimingMutation();

  const dirty = Number(value) !== challenge.timeLimitSeconds;

  const handleSave = () => {
    const parsed = Number(value);
    if (!Number.isFinite(parsed) || parsed <= 0) return;
    updateMutation.mutate(
      { challengeId: challenge.id, timeLimitSeconds: parsed },
      {
        onSuccess: () => {
          setSaved(true);
          window.setTimeout(() => setSaved(false), 1500);
        },
      },
    );
  };

  return (
    <div className="flex items-center gap-3 py-3">
      <div className="flex flex-1 flex-col">
        <span className="flex items-center gap-1.5 text-sm font-medium text-black dark:text-zinc-50">
          {challenge.isExam && <span>🏁</span>}
          {challenge.name}
        </span>
        <span className="text-xs text-zinc-500 dark:text-zinc-500">
          {challenge.questionBankSize} soal di bank · {challenge.questionCountRequired} diambil ·
          lulus {challenge.passThresholdPercent}%
        </span>
        <Link
          href={`/admin/kurikulum/${challenge.id}`}
          className="text-xs font-medium text-blue-600 dark:text-blue-400"
        >
          Kelola Soal →
        </Link>
      </div>

      <div className="flex items-center gap-2">
        <input
          type="number"
          min={1}
          value={value}
          onChange={(e) => setValue(e.target.value)}
          className="w-20 rounded-lg border border-black/[.08] bg-transparent px-2 py-1.5 text-right text-sm text-black outline-none focus:border-blue-500 dark:border-white/[.145] dark:text-zinc-50"
        />
        <span className="text-xs text-zinc-500 dark:text-zinc-500">
          {challenge.isExam ? "detik total" : "detik/soal"}
        </span>
        <button
          type="button"
          onClick={handleSave}
          disabled={!dirty || updateMutation.isPending}
          className="rounded-lg bg-foreground px-3 py-1.5 text-xs font-medium text-background disabled:opacity-40"
        >
          {saved ? "Tersimpan ✓" : updateMutation.isPending ? "…" : "Simpan"}
        </button>
      </div>
    </div>
  );
}
