"use client";

import Link from "next/link";
import { useParams, useRouter } from "next/navigation";
import { useEffect, useState } from "react";
import { useAuthStore } from "@/store/auth-store";
import { useLogoutMutation } from "@/hooks/use-auth";
import {
  useAdminCategoriesQuery,
  useUpdateChallengeTimingMutation,
  type AdminChallenge,
} from "@/hooks/use-admin-curriculum";
import { PLATFORM_ROLE_LABEL } from "@/lib/dummy-accounts";
import { BackButton } from "@/components/back-button";

export default function AdminKategoriPage() {
  const router = useRouter();
  const params = useParams<{ tierCode: string }>();
  const session = useAuthStore((s) => s.session);
  const hasHydrated = useAuthStore((s) => s.hasHydrated);
  const logoutMutation = useLogoutMutation();
  const categoriesQuery = useAdminCategoriesQuery(params.tierCode);

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
          <div className="-ml-2 flex items-center gap-1">
            <BackButton href="/admin/kurikulum" label="Kembali ke Kurikulum" />
            <h1 className="text-2xl font-semibold tracking-tight text-black dark:text-zinc-50">Kategori — Tier {params.tierCode}</h1>
          </div>
          <p className="text-sm text-zinc-500 dark:text-zinc-500">
            Challenge di tier ini nempel ke kategori (bukan batch), jadi ditampilin di halaman
            terpisah. Klik &ldquo;Kelola Soal&rdquo; buat CRUD soal puzzle-nya (addition grid /
            cryptarithm).
          </p>
        </div>

        {categoriesQuery.isLoading && (
          <p className="text-sm text-zinc-500 dark:text-zinc-500">Memuat kategori…</p>
        )}

        {categoriesQuery.data?.length === 0 && (
          <p className="rounded-2xl border border-dashed border-black/[.08] px-5 py-6 text-center text-sm text-zinc-500 dark:border-white/[.145]">
            Tier ini belum punya kategori.
          </p>
        )}

        <div className="flex flex-col gap-6">
          {categoriesQuery.data?.map((category) => (
            <div key={category.id} className="flex flex-col gap-2">
              <h2 className="text-sm font-semibold tracking-wide text-zinc-500 uppercase dark:text-zinc-500">
                {category.name}
              </h2>
              <div className="flex flex-col gap-2 rounded-2xl border border-black/[.08] bg-white p-5 dark:border-white/[.145] dark:bg-zinc-900">
                <div className="flex flex-col divide-y divide-black/[.06] dark:divide-white/[.08]">
                  {category.challenges.map((challenge) => (
                    <CategoryChallengeRow
                      key={challenge.id}
                      tierCode={params.tierCode}
                      challenge={challenge}
                    />
                  ))}
                </div>
              </div>
            </div>
          ))}
        </div>
      </main>
    </div>
  );
}

function CategoryChallengeRow({
  tierCode,
  challenge,
}: {
  tierCode: string;
  challenge: AdminChallenge;
}) {
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
        <span className="text-sm font-medium text-black dark:text-zinc-50">{challenge.name}</span>
        <span className="text-xs text-zinc-500 dark:text-zinc-500">
          {challenge.questionBankSize} soal di bank · {challenge.questionCountRequired} diambil ·
          lulus {challenge.passThresholdPercent}%
        </span>
        <Link
          href={`/admin/kurikulum/kategori/${tierCode}/${challenge.id}`}
          className="w-fit text-xs font-medium text-blue-600 dark:text-blue-400"
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
