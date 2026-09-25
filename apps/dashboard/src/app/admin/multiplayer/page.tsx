"use client";

import { useRouter } from "next/navigation";
import { useEffect, useState } from "react";
import { useAuthStore } from "@/store/auth-store";
import {
  useMultiplayerConfigQuery,
  useUpdateMultiplayerConfigMutation,
  type MultiplayerConfig,
} from "@/hooks/use-admin-multiplayer";
import { BackButton } from "@/components/back-button";

// Batas nilai sama dengan validasi server (curriculumsvc.MultiplayerConfig).
const FIELDS: { key: keyof MultiplayerConfig; label: string; hint: string; min: number; max: number }[] = [
  {
    key: "classicCooldownSeconds",
    label: "Cooldown Klasik",
    hint: "Jeda antar soal mode Klasik: angka hitung mundur gede + kunci jawaban & poin soal barusan.",
    min: 1,
    max: 15,
  },
  {
    key: "raceCooldownSeconds",
    label: "Cooldown Adu Cepat",
    hint: "Jeda antar soal mode Adu Cepat.",
    min: 1,
    max: 15,
  },
  {
    key: "raceWinnerRevealSeconds",
    label: "Pamer yang paling cepat",
    hint: "Adu Cepat (kalau room-nya nampilin yang paling cepat): nama yang jawab benar duluan muncul segini lama sebelum cooldown.",
    min: 1,
    max: 5,
  },
  {
    key: "resultsCountdownSeconds",
    label: "Countdown hasil akhir",
    hint: "Adu Cepat yang hasilnya dirahasiain: hitung mundur setelah host klik \"Tampilkan hasil\".",
    min: 1,
    max: 15,
  },
];

export default function AdminMultiplayerPage() {
  const router = useRouter();
  const session = useAuthStore((s) => s.session);
  const hasHydrated = useAuthStore((s) => s.hasHydrated);
  const configQuery = useMultiplayerConfigQuery();

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
      <main className="mx-auto flex w-full max-w-2xl flex-1 flex-col gap-6 px-6 py-8">
        <div className="flex flex-col gap-1">
          <div className="-ml-2 flex items-center gap-1">
            <BackButton href="/admin" label="Dashboard Admin" />
            <h1 className="text-2xl font-semibold tracking-tight text-black dark:text-zinc-50">Pengaturan Multiplayer</h1>
          </div>
          <p className="text-sm text-zinc-500 dark:text-zinc-500">
            Berlaku buat room yang dibikin setelah disimpan. Room yang lagi jalan tetap pakai pengaturan lama.
          </p>
        </div>

        {configQuery.isLoading && <p className="text-sm text-zinc-500 dark:text-zinc-500">Memuat pengaturan…</p>}
        {configQuery.isError && <p className="text-sm text-red-500">Gagal memuat pengaturan multiplayer.</p>}
        {configQuery.data && <ConfigForm key={JSON.stringify(configQuery.data)} initial={configQuery.data} />}
      </main>
    </div>
  );
}

function ConfigForm({ initial }: { initial: MultiplayerConfig }) {
  const [config, setConfig] = useState(initial);
  const [saved, setSaved] = useState(false);
  const mutation = useUpdateMultiplayerConfigMutation();
  const dirty = FIELDS.some((f) => config[f.key] !== initial[f.key]);

  const set = (key: keyof MultiplayerConfig, value: number) => {
    setSaved(false);
    setConfig((c) => ({ ...c, [key]: value }));
  };

  return (
    <form
      onSubmit={(e) => {
        e.preventDefault();
        mutation.mutate(config, { onSuccess: () => setSaved(true) });
      }}
      className="flex flex-col gap-4"
    >
      {FIELDS.map((f) => (
        <div
          key={f.key}
          className="flex items-center justify-between gap-4 rounded-2xl border border-black/[.08] bg-white p-5 dark:border-white/[.145] dark:bg-zinc-900"
        >
          <div className="flex flex-col gap-1">
            <span className="text-sm font-semibold text-black dark:text-zinc-50">{f.label}</span>
            <span className="text-xs text-zinc-500 dark:text-zinc-500">{f.hint}</span>
          </div>
          <div className="flex shrink-0 items-center gap-2">
            <Stepper disabled={config[f.key] <= f.min} onClick={() => set(f.key, config[f.key] - 1)} label="−" />
            <span className="w-16 text-center text-lg font-semibold text-black tabular-nums dark:text-zinc-50">
              {config[f.key]} dtk
            </span>
            <Stepper disabled={config[f.key] >= f.max} onClick={() => set(f.key, config[f.key] + 1)} label="+" />
          </div>
        </div>
      ))}

      {mutation.isError && <p className="text-sm text-red-500">Gagal menyimpan, coba lagi.</p>}
      <button
        type="submit"
        disabled={!dirty || mutation.isPending}
        className="w-fit rounded-full bg-foreground px-6 py-2.5 text-sm font-medium text-background disabled:opacity-40"
      >
        {mutation.isPending ? "Menyimpan…" : saved && !dirty ? "Tersimpan ✓" : "Simpan"}
      </button>
    </form>
  );
}

function Stepper({ label, disabled, onClick }: { label: string; disabled: boolean; onClick: () => void }) {
  return (
    <button
      type="button"
      onClick={onClick}
      disabled={disabled}
      className="flex h-9 w-9 items-center justify-center rounded-full border border-black/[.08] text-lg font-medium text-zinc-600 transition-colors hover:bg-black/[.04] disabled:opacity-30 dark:border-white/[.145] dark:text-zinc-400 dark:hover:bg-[#1a1a1a]"
    >
      {label}
    </button>
  );
}
