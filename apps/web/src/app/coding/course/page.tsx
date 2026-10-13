"use client";

import Link from "next/link";
import { CourseShell } from "@/components/course-shell";
import { formatDuration, useCourseModulesQuery, type CourseModule } from "@/hooks/use-course";

const MODULE_ICON: Record<string, string> = {
  variabel: "📦",
  looping: "🔁",
  async: "⏳",
};

export default function CourseModulesPage() {
  const modulesQuery = useCourseModulesQuery();

  return (
    <CourseShell
      backHref="/coding"
      backLabel="Kembali ke Coding"
      title="Course JavaScript Dasar"
      subtitle="Selesaikan module berurutan. Lulus final test buat buka module berikutnya — gak ada nyawa, bebas ngulang."
    >
      {modulesQuery.isLoading && <p className="text-sm text-zinc-500">Memuat module…</p>}
      {modulesQuery.isError && <p className="text-sm text-red-500">Gagal memuat course. Coba refresh.</p>}
      <div className="flex flex-col gap-5">
        {modulesQuery.data?.map((m, i) => (
          <ModuleCard key={m.code} module={m} number={i + 1} />
        ))}
      </div>
    </CourseShell>
  );
}

function ModuleCard({ module: m, number }: { module: CourseModule; number: number }) {
  const passedLevels = m.levels.filter((l) => l.passed).length;
  const status = m.completed
    ? { label: "Selesai ✓", cls: "bg-emerald-100 text-emerald-700 dark:bg-emerald-500/15 dark:text-emerald-400" }
    : m.comingSoon
      ? { label: "Segera hadir", cls: "bg-zinc-200 text-zinc-600 dark:bg-zinc-700 dark:text-zinc-300" }
      : m.locked
        ? { label: "🔒 Terkunci", cls: "bg-zinc-200 text-zinc-600 dark:bg-zinc-700 dark:text-zinc-300" }
        : { label: `${passedLevels}/${m.levels.length} level`, cls: "bg-sky-100 text-sky-700 dark:bg-sky-500/15 dark:text-sky-300" };

  return (
    <section
      className={`flex flex-col gap-4 rounded-2xl border bg-white p-5 dark:bg-zinc-900 ${
        m.locked || m.comingSoon ? "border-black/[.06] opacity-70 dark:border-white/[.1]" : "border-black/[.08] dark:border-white/[.145]"
      }`}
    >
      <div className="flex items-start gap-4">
        <span className="text-3xl">{MODULE_ICON[m.code] ?? "📘"}</span>
        <div className="flex min-w-0 flex-1 flex-col gap-1">
          <span className="text-xs font-semibold tracking-wide text-zinc-500 uppercase">Module {number}</span>
          <span className="text-lg font-semibold text-black dark:text-zinc-50">{m.title}</span>
          <span className="text-sm text-zinc-500 dark:text-zinc-400">{m.description}</span>
        </div>
        <span className={`shrink-0 rounded-full px-2.5 py-1 text-xs font-semibold ${status.cls}`}>{status.label}</span>
      </div>

      {m.locked && !m.comingSoon && (
        <p className="text-sm text-zinc-500">Lulus final test module sebelumnya buat buka module ini.</p>
      )}

      {!m.comingSoon && (
        <ol className="flex flex-col gap-2">
          {m.levels.map((l, i) => {
            const content = (
              <>
                <span
                  className={`flex h-8 w-8 shrink-0 items-center justify-center rounded-full text-sm font-semibold ${
                    l.passed
                      ? "bg-emerald-500 text-white"
                      : l.locked
                        ? "bg-black/[.06] text-zinc-400 dark:bg-white/[.08]"
                        : "bg-sky-600 text-white"
                  }`}
                >
                  {l.passed ? "✓" : i + 1}
                </span>
                <span className="flex min-w-0 flex-1 flex-col">
                  <span className="font-medium text-black dark:text-zinc-50">{l.title}</span>
                  <span className="text-xs text-zinc-500 dark:text-zinc-400">{l.summary}</span>
                </span>
                {l.locked && <span aria-label="Terkunci">🔒</span>}
              </>
            );
            const cls = "flex items-center gap-3 rounded-xl border border-black/[.06] px-3 py-3 dark:border-white/[.08]";
            return (
              <li key={l.id}>
                {l.locked ? (
                  <div className={`${cls} cursor-not-allowed opacity-60`}>{content}</div>
                ) : (
                  <Link href={`/coding/course/level/${l.id}`} className={`${cls} transition-colors hover:border-sky-400`}>
                    {content}
                  </Link>
                )}
              </li>
            );
          })}
          <li>
            <ExamRow module={m} />
          </li>
        </ol>
      )}
    </section>
  );
}

function ExamRow({ module: m }: { module: CourseModule }) {
  const e = m.exam;
  const cls = `flex items-center gap-3 rounded-xl border-2 px-3 py-3 ${
    e.locked ? "border-dashed border-black/[.08] opacity-60 dark:border-white/[.12]" : "border-amber-400"
  }`;
  const content = (
    <>
      <span className="flex h-8 w-8 shrink-0 items-center justify-center rounded-full bg-amber-500 text-sm text-white">
        {e.passed ? "🏆" : "🏁"}
      </span>
      <span className="flex min-w-0 flex-1 flex-col">
        <span className="font-semibold text-black dark:text-zinc-50">Final Test — Speedrun</span>
        <span className="text-xs text-zinc-500 dark:text-zinc-400">
          {e.questionCount} soal acak dari {e.bankSize} · {formatDuration(e.timeLimitSeconds * 1000)} · lulus {e.passPercent}%
          {e.bestTimeMs !== undefined && ` · rekor ${formatDuration(e.bestTimeMs)}`}
        </span>
        {e.locked && !m.locked && (
          <span className="text-xs font-medium text-amber-600 dark:text-amber-400">Lulusin semua level dulu</span>
        )}
      </span>
      {e.locked && <span aria-label="Terkunci">🔒</span>}
    </>
  );
  return e.locked ? (
    <div className={`${cls} cursor-not-allowed`}>{content}</div>
  ) : (
    <Link href={`/coding/course/ujian/${m.code}`} className={`${cls} transition-colors hover:bg-amber-50 dark:hover:bg-amber-500/10`}>
      {content}
    </Link>
  );
}
