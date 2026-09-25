"use client";

import Link from "next/link";
import { useRouter } from "next/navigation";
import { useEffect, useState } from "react";
import { useAuthStore } from "@/store/auth-store";
import { useRequireNickname } from "@/hooks/use-profile";
import {
  errorCode,
  useCreateRoomMutation,
  useMultiplayerOptionsQuery,
  useRoomQuotaQuery,
  type MultiplayerFormat,
  type MultiplayerMode,
  type MultiplayerSourceBatch,
} from "@/hooks/use-multiplayer";
import { BackButton } from "@/components/back-button";
import { FORMAT_LABEL, MODE_INFO, TIER_BADGE } from "@/lib/multiplayer";

const CREATE_ERROR_TEXT: Record<string, string> = {
  no_room_quota: "Kesempatan bikin room kamu udah habis.",
  not_enough_questions: "Soal di racikan ini gak cukup. Tambah batch lain atau kurangin jumlah soal.",
  already_hosting: "Game kamu yang sebelumnya masih jalan. Tunggu selesai dulu ya.",
  invalid_settings: "Pengaturan room gak valid.",
  invalid_source: "Pilihan batch gak valid, coba muat ulang halaman.",
};

export default function CreateRoomPage() {
  const router = useRouter();
  const session = useAuthStore((s) => s.session);
  const hasHydrated = useAuthStore((s) => s.hasHydrated);
  const me = useRequireNickname().data;
  const optionsQuery = useMultiplayerOptionsQuery();
  const quotaQuery = useRoomQuotaQuery();
  const createMutation = useCreateRoomMutation();

  const [mode, setMode] = useState<MultiplayerMode>("classic");
  const [questionCount, setQuestionCount] = useState(10);
  const [format, setFormat] = useState<MultiplayerFormat>("mc");
  const [seconds, setSeconds] = useState(20);
  const [batchIds, setBatchIds] = useState<string[]>([]);
  const [hostPlays, setHostPlays] = useState(true);
  const [showFastest, setShowFastest] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (!hasHydrated) return;
    if (session === null) router.replace("/login");
  }, [hasHydrated, session, router]);

  const quota = quotaQuery.data;
  if (!hasHydrated || !session || !me || !quota) {
    return <Centered text={quotaQuery.isError ? "Gagal memuat sisa kesempatan." : "Memuat…"} />;
  }
  if (!quota.isPremium && quota.roomQuota === 0) {
    return (
      <Centered text="Kesempatan bikin room kamu udah habis. Minta tambahan ke admin dulu ya.">
        <Link href="/multiplayer" className="rounded-full bg-foreground px-6 py-3 text-sm font-semibold text-background">
          Minta tambahan kesempatan
        </Link>
      </Centered>
    );
  }

  const options = optionsQuery.data;
  const countOptions = options ? (mode === "race" ? options.raceQuestionCounts : options.classicQuestionCounts) : [];
  const usable = (b: MultiplayerSourceBatch) => (format === "mc" ? b.mcCount : b.questionCount);
  const allBatches = options?.tiers.flatMap((t) => t.batches) ?? [];
  const selectedUsable = allBatches.filter((b) => batchIds.includes(b.id)).reduce((sum, b) => sum + usable(b), 0);

  const pickMode = (next: MultiplayerMode) => {
    setMode(next);
    const counts = next === "race" ? options?.raceQuestionCounts : options?.classicQuestionCounts;
    if (counts && !counts.includes(questionCount)) {
      // Pindah mode: ambil jumlah terdekat di daftar mode baru (10 -> 5/15 dst).
      setQuestionCount(counts.reduce((best, n) => (Math.abs(n - questionCount) < Math.abs(best - questionCount) ? n : best)));
    }
  };

  const pickFormat = (next: MultiplayerFormat) => {
    setFormat(next);
    // Batch yang gak punya soal PG sama sekali gak bisa dipake format PG.
    if (next === "mc") setBatchIds((ids) => ids.filter((id) => (allBatches.find((b) => b.id === id)?.mcCount ?? 0) > 0));
  };

  const toggleBatch = (id: string) =>
    setBatchIds((ids) => (ids.includes(id) ? ids.filter((x) => x !== id) : [...ids, id]));

  const toggleTier = (batches: MultiplayerSourceBatch[]) => {
    const ids = batches.filter((b) => usable(b) > 0).map((b) => b.id);
    const allOn = ids.every((id) => batchIds.includes(id));
    setBatchIds((cur) => (allOn ? cur.filter((id) => !ids.includes(id)) : [...new Set([...cur, ...ids])]));
  };

  const submit = () => {
    setError(null);
    createMutation.mutate(
      { mode, questionCount, format, secondsPerQuestion: seconds, batchIds, hostPlays, showFastest },
      {
        onSuccess: ({ code }) => router.push(`/multiplayer/${code}`),
        onError: (err) => setError(CREATE_ERROR_TEXT[errorCode(err) ?? ""] ?? "Gagal bikin room, coba lagi."),
      },
    );
  };

  return (
    <div className="flex min-h-screen flex-col bg-zinc-50 font-sans dark:bg-black">
      <header className="flex items-center justify-between px-4 py-5 sm:px-10">
        <BackButton href="/multiplayer" label="Kembali" />
        <span className="text-lg font-semibold tracking-tight text-black dark:text-zinc-50">Bikin Room</span>
        <span className="w-9" />
      </header>

      <main className="mx-auto flex w-full max-w-xl flex-1 flex-col gap-6 px-4 pb-32">
        <Section title="Mode">
          <div className="grid grid-cols-1 gap-3 sm:grid-cols-2">
            {(Object.keys(MODE_INFO) as MultiplayerMode[]).map((m) => (
              <button
                key={m}
                type="button"
                onClick={() => pickMode(m)}
                className={`flex flex-col gap-1 rounded-2xl border-2 p-4 text-left transition-colors ${
                  mode === m
                    ? "border-violet-500 bg-violet-50 dark:bg-violet-500/10"
                    : "border-black/[.08] bg-white dark:border-white/[.145] dark:bg-zinc-900"
                }`}
              >
                <span className="font-semibold text-black dark:text-zinc-50">
                  {MODE_INFO[m].icon} {MODE_INFO[m].label}
                </span>
                <span className="text-xs text-zinc-500 dark:text-zinc-400">{MODE_INFO[m].desc}</span>
              </button>
            ))}
          </div>
        </Section>

        <Section title="Jumlah soal" hint={mode === "race" ? "Adu cepat pakai jumlah ganjil biar jarang seri." : undefined}>
          <Chips values={countOptions} selected={questionCount} onPick={setQuestionCount} />
        </Section>

        {mode === "race" && (
          <Section
            title="Tampilkan yang paling cepat"
            hint={
              showFastest
                ? "Tiap soal, nama yang jawab benar paling duluan dipamerin sebentar."
                : "Pemenang tiap soal dirahasiain. Hasil akhir baru keluar setelah kamu klik \"Tampilkan hasil\"."
            }
          >
            <Chips
              values={[true, false]}
              selected={showFastest}
              onPick={setShowFastest}
              label={(v) => (v ? "⚡ Tampilkan" : "🤫 Rahasiain")}
            />
          </Section>
        )}

        <Section title="Bentuk soal" hint={format === "mixed" ? "Separuh pilihan ganda, separuh isian." : undefined}>
          <Chips
            values={Object.keys(FORMAT_LABEL) as MultiplayerFormat[]}
            selected={format}
            onPick={pickFormat}
            label={(f) => FORMAT_LABEL[f]}
          />
        </Section>

        <Section title="Waktu per soal">
          <Chips values={options?.secondsPerQuestionOptions ?? []} selected={seconds} onPick={setSeconds} label={(s) => `${s} detik`} />
        </Section>

        <Section title="Racikan soal" hint="Pilih satu atau lebih batch. Soal diambil acak dan dibagi rata dari semua batch yang dipilih.">
          {optionsQuery.isLoading && <p className="text-sm text-zinc-500">Memuat bank soal…</p>}
          {optionsQuery.isError && <p className="text-sm text-red-500">Gagal memuat bank soal.</p>}
          <div className="flex flex-col gap-4">
            {options?.tiers.map((tier) => {
              const pickable = tier.batches.filter((b) => usable(b) > 0);
              const allOn = pickable.length > 0 && pickable.every((b) => batchIds.includes(b.id));
              return (
                <div
                  key={tier.tierCode}
                  className="flex flex-col gap-2 rounded-2xl border border-black/[.08] bg-white p-4 dark:border-white/[.145] dark:bg-zinc-900"
                >
                  <div className="flex items-center justify-between">
                    <span className={`rounded-full px-3 py-1 text-xs font-semibold ${TIER_BADGE[tier.tierCode] ?? ""}`}>
                      {tier.tierName}
                    </span>
                    <button
                      type="button"
                      onClick={() => toggleTier(tier.batches)}
                      disabled={pickable.length === 0}
                      className="text-xs font-medium text-violet-600 disabled:opacity-40 dark:text-violet-400"
                    >
                      {allOn ? "Batal semua" : "Pilih semua"}
                    </button>
                  </div>
                  {tier.batches.map((b) => {
                    const count = usable(b);
                    const on = batchIds.includes(b.id);
                    return (
                      <label
                        key={b.id}
                        className={`flex items-center gap-3 rounded-xl px-2 py-2 ${count === 0 ? "opacity-40" : "cursor-pointer hover:bg-black/[.03] dark:hover:bg-white/[.04]"}`}
                      >
                        <input
                          type="checkbox"
                          checked={on}
                          disabled={count === 0}
                          onChange={() => toggleBatch(b.id)}
                          className="h-4 w-4 accent-violet-600"
                        />
                        <span className="flex-1 text-sm text-black dark:text-zinc-50">{b.name}</span>
                        <span className="text-xs text-zinc-400 dark:text-zinc-600">{count} soal</span>
                      </label>
                    );
                  })}
                </div>
              );
            })}
          </div>
        </Section>

        <Section title="Kamu di room ini">
          <Chips
            values={[true, false]}
            selected={hostPlays}
            onPick={setHostPlays}
            label={(v) => (v ? "🎮 Ikut main" : "👀 Cuma nonton")}
          />
        </Section>
      </main>

      <div className="fixed inset-x-0 bottom-0 border-t border-black/[.08] bg-white/95 px-4 py-3 backdrop-blur dark:border-white/[.145] dark:bg-black/90">
        <div className="mx-auto flex w-full max-w-xl flex-col gap-2">
          {error && <p className="text-center text-sm text-red-500">{error}</p>}
          {!quota.isPremium && !error && (
            <p className="text-center text-xs text-zinc-500">
              Bikin room ini makai 1 dari {quota.roomQuota} kesempatanmu.
            </p>
          )}
          {batchIds.length > 0 && selectedUsable < questionCount && !error && (
            <p className="text-center text-xs text-amber-600 dark:text-amber-400">
              Bank soal terpilih cuma {selectedUsable}, butuh {questionCount}. Tambah batch lagi.
            </p>
          )}
          <button
            type="button"
            onClick={submit}
            disabled={batchIds.length === 0 || selectedUsable < questionCount || createMutation.isPending}
            className="rounded-full bg-violet-600 px-6 py-3 text-sm font-semibold text-white transition-colors hover:bg-violet-700 disabled:opacity-40"
          >
            {createMutation.isPending
              ? "Menyiapkan room…"
              : `Bikin room · ${MODE_INFO[mode].label} · ${questionCount} soal`}
          </button>
        </div>
      </div>
    </div>
  );
}

function Centered({ text, children }: { text: string; children?: React.ReactNode }) {
  return (
    <div className="flex min-h-screen flex-col items-center justify-center gap-4 bg-zinc-50 px-6 text-center dark:bg-black">
      <p className="text-zinc-500 dark:text-zinc-500">{text}</p>
      {children}
    </div>
  );
}

function Section({ title, hint, children }: { title: string; hint?: string; children: React.ReactNode }) {
  return (
    <section className="flex flex-col gap-2">
      <span className="text-sm font-semibold text-zinc-700 dark:text-zinc-300">{title}</span>
      {hint && <span className="-mt-1 text-xs text-zinc-500 dark:text-zinc-500">{hint}</span>}
      {children}
    </section>
  );
}

function Chips<T extends string | number | boolean>({
  values,
  selected,
  onPick,
  label = (v) => String(v),
}: {
  values: T[];
  selected: T;
  onPick: (value: T) => void;
  label?: (value: T) => string;
}) {
  return (
    <div className="flex flex-wrap gap-2">
      {values.map((v) => (
        <button
          key={String(v)}
          type="button"
          onClick={() => onPick(v)}
          className={`rounded-full border-2 px-4 py-2 text-sm font-medium transition-colors ${
            v === selected
              ? "border-violet-500 bg-violet-50 text-violet-700 dark:bg-violet-500/10 dark:text-violet-300"
              : "border-black/[.08] bg-white text-zinc-700 dark:border-white/[.145] dark:bg-zinc-900 dark:text-zinc-300"
          }`}
        >
          {label(v)}
        </button>
      ))}
    </div>
  );
}
