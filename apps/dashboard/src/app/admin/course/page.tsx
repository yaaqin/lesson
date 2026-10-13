"use client";

import { useRouter } from "next/navigation";
import { useEffect, useState } from "react";
import { useAuthStore } from "@/store/auth-store";
import {
  useCourseConfigQuery,
  useUpdateCourseConfigMutation,
  type CourseConfig,
  type ReviewMode,
  type ReviewSetting,
} from "@/hooks/use-admin-course";
import { BackButton } from "@/components/back-button";

// Batas detik sama dengan validasi server (coursesvc.ReviewSetting).
const MIN_SECONDS = 1;
const MAX_SECONDS = 120;

const MODES: { value: ReviewMode; label: string; hint: string }[] = [
  {
    value: "full",
    label: "Tampil penuh",
    hint: "Jawaban user, kunci jawaban, dan penjelasan tampil terus.",
  },
  {
    value: "timed",
    label: "Tampil sebentar",
    hint: "Pembahasan lengkap muncul beberapa detik, abis itu hilang — tinggal benar/salah per soal.",
  },
  {
    value: "result_only",
    label: "Cuma benar / salah",
    hint: "Gak ada kunci & penjelasan sama sekali (gak dikirim server juga, jadi gak bisa diintip).",
  },
];

const SECTIONS: { key: keyof CourseConfig; title: string; hint: string }[] = [
  { key: "level", title: "Latihan level", hint: "Pembahasan setelah latihan tiap level (lulus = 100% bener)." },
  { key: "exam", title: "Final test", hint: "Pembahasan setelah speedrun final test tiap module." },
];

export default function AdminCoursePage() {
  const router = useRouter();
  const session = useAuthStore((s) => s.session);
  const hasHydrated = useAuthStore((s) => s.hasHydrated);
  const configQuery = useCourseConfigQuery();

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
            <h1 className="text-2xl font-semibold tracking-tight text-black dark:text-zinc-50">Pengaturan Course</h1>
          </div>
          <p className="text-sm text-zinc-500 dark:text-zinc-500">
            Atur cara pembahasan ditampilin ke murid setelah ngumpulin jawaban di Course Coding Dasar. Berlaku buat
            submit setelah disimpan.
          </p>
        </div>

        {configQuery.isLoading && <p className="text-sm text-zinc-500 dark:text-zinc-500">Memuat pengaturan…</p>}
        {configQuery.isError && <p className="text-sm text-red-500">Gagal memuat pengaturan course.</p>}
        {configQuery.data && <ConfigForm key={JSON.stringify(configQuery.data)} initial={configQuery.data} />}
      </main>
    </div>
  );
}

function ConfigForm({ initial }: { initial: CourseConfig }) {
  const [config, setConfig] = useState(initial);
  const [saved, setSaved] = useState(false);
  const mutation = useUpdateCourseConfigMutation();
  const dirty = JSON.stringify(config) !== JSON.stringify(initial);

  const set = (key: keyof CourseConfig, patch: Partial<ReviewSetting>) => {
    setSaved(false);
    setConfig((c) => ({ ...c, [key]: { ...c[key], ...patch } }));
  };

  return (
    <form
      onSubmit={(e) => {
        e.preventDefault();
        mutation.mutate(config, { onSuccess: () => setSaved(true) });
      }}
      className="flex flex-col gap-6"
    >
      {SECTIONS.map((section) => {
        const value = config[section.key];
        return (
          <section
            key={section.key}
            className="flex flex-col gap-4 rounded-2xl border border-black/[.08] bg-white p-5 dark:border-white/[.145] dark:bg-zinc-900"
          >
            <div className="flex flex-col gap-1">
              <span className="font-semibold text-black dark:text-zinc-50">{section.title}</span>
              <span className="text-xs text-zinc-500 dark:text-zinc-500">{section.hint}</span>
            </div>

            <div className="flex flex-col gap-2" role="radiogroup" aria-label={`Mode pembahasan ${section.title}`}>
              {MODES.map((m) => {
                const active = value.mode === m.value;
                return (
                  <button
                    key={m.value}
                    type="button"
                    role="radio"
                    aria-checked={active}
                    onClick={() => set(section.key, { mode: m.value })}
                    className={`flex items-start gap-3 rounded-xl border-2 px-4 py-3 text-left transition-colors ${
                      active
                        ? "border-blue-600 bg-blue-50 dark:border-blue-500 dark:bg-blue-500/10"
                        : "border-black/[.06] hover:border-blue-300 dark:border-white/[.1]"
                    }`}
                  >
                    <span
                      className={`mt-0.5 flex h-4 w-4 shrink-0 items-center justify-center rounded-full border-2 ${
                        active ? "border-blue-600 dark:border-blue-400" : "border-zinc-300 dark:border-zinc-600"
                      }`}
                    >
                      {active && <span className="h-2 w-2 rounded-full bg-blue-600 dark:bg-blue-400" />}
                    </span>
                    <span className="flex flex-col gap-0.5">
                      <span className="text-sm font-semibold text-black dark:text-zinc-50">{m.label}</span>
                      <span className="text-xs text-zinc-500 dark:text-zinc-500">{m.hint}</span>
                    </span>
                  </button>
                );
              })}
            </div>

            {value.mode === "timed" && (
              <div className="flex items-center justify-between gap-4 rounded-xl bg-black/[.03] px-4 py-3 dark:bg-white/[.04]">
                <span className="text-sm font-medium text-black dark:text-zinc-50">Lama pembahasan tampil</span>
                <div className="flex shrink-0 items-center gap-2">
                  <Stepper
                    disabled={value.seconds <= MIN_SECONDS}
                    onClick={() => set(section.key, { seconds: value.seconds - 1 })}
                    label="−"
                  />
                  <span className="w-16 text-center text-lg font-semibold text-black tabular-nums dark:text-zinc-50">
                    {value.seconds} dtk
                  </span>
                  <Stepper
                    disabled={value.seconds >= MAX_SECONDS}
                    onClick={() => set(section.key, { seconds: value.seconds + 1 })}
                    label="+"
                  />
                </div>
              </div>
            )}
          </section>
        );
      })}

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
