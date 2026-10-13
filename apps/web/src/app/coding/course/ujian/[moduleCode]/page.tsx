"use client";

import Link from "next/link";
import { useParams } from "next/navigation";
import { useState } from "react";
import { CourseShell } from "@/components/course-shell";
import { CourseQuiz, CourseResult } from "@/components/course-quiz";
import {
  formatDuration,
  useCourseModulesQuery,
  useStartCourseExamMutation,
  useSubmitCourseAttemptMutation,
  type CourseAnswer,
  type CourseStartResult,
  type CourseSubmitResult,
} from "@/hooks/use-course";

// Final test module: speedrun -- soal acak dari bank seluruh module (lebih
// banyak dari yang dikerjain), satu timer total, auto-submit pas waktu habis.
type Phase = "intro" | "quiz" | "result";

export default function CourseExamPage() {
  const { moduleCode } = useParams<{ moduleCode: string }>();
  const modulesQuery = useCourseModulesQuery();
  const startMutation = useStartCourseExamMutation();
  const submitMutation = useSubmitCourseAttemptMutation();

  const [phase, setPhase] = useState<Phase>("intro");
  const [attempt, setAttempt] = useState<CourseStartResult | null>(null);
  const [answers, setAnswers] = useState<CourseAnswer[]>([]);
  const [result, setResult] = useState<CourseSubmitResult | null>(null);

  const moduleIndex = modulesQuery.data?.findIndex((m) => m.code === moduleCode) ?? -1;
  const mod = moduleIndex >= 0 ? modulesQuery.data?.[moduleIndex] : undefined;
  const nextModule = moduleIndex >= 0 ? modulesQuery.data?.[moduleIndex + 1] : undefined;

  const start = () =>
    startMutation.mutate(moduleCode, {
      onSuccess: (data) => {
        setAttempt(data);
        setResult(null);
        setPhase("quiz");
        window.scrollTo({ top: 0 });
      },
    });

  const submit = (a: CourseAnswer[]) => {
    if (!attempt) return;
    setAnswers(a);
    submitMutation.mutate(
      { attemptId: attempt.attemptId, answers: a },
      {
        onSuccess: (data) => {
          setResult(data);
          setPhase("result");
          window.scrollTo({ top: 0 });
        },
      },
    );
  };

  const btn = "rounded-full px-5 py-2 text-sm font-semibold transition-colors";
  const e = mod?.exam;

  return (
    <CourseShell
      backHref="/coding/course"
      backLabel="Kembali ke daftar module"
      title={mod ? `Final Test — ${mod.title}` : "Final Test"}
      subtitle="Speedrun: jawab secepat & setepat mungkin."
    >
      {phase === "intro" && e && (
        <div className="flex flex-col gap-4 rounded-2xl border-2 border-amber-400 bg-white p-6 dark:bg-zinc-900">
          <span className="text-4xl">🏁</span>
          <ul className="flex flex-col gap-2 text-sm text-zinc-700 dark:text-zinc-300">
            <li>
              📝 <b>{e.questionCount} soal</b> diacak dari bank {e.bankSize} soal — tiap percobaan soalnya beda.
            </li>
            <li>
              ⏱ Waktu total <b>{formatDuration(e.timeLimitSeconds * 1000)}</b>, otomatis dikumpulin pas habis.
            </li>
            <li>
              🎯 Lulus minimal <b>{e.passPercent}%</b> buat buka module berikutnya.
            </li>
            <li>♾️ Gak ada nyawa — gagal? Langsung coba lagi.</li>
            {e.bestTimeMs !== undefined && (
              <li>
                🏆 Rekor waktumu: <b>{formatDuration(e.bestTimeMs)}</b> — coba pecahin!
              </li>
            )}
          </ul>
          {e.locked ? (
            <p className="text-sm font-medium text-amber-600 dark:text-amber-400">
              Final test masih kekunci — lulusin semua level di module ini dulu.
            </p>
          ) : (
            <button
              type="button"
              onClick={start}
              disabled={startMutation.isPending}
              className={`${btn} w-fit bg-amber-500 text-white hover:bg-amber-600 disabled:opacity-60`}
            >
              {startMutation.isPending ? "Menyiapkan…" : "Mulai Speedrun"}
            </button>
          )}
          {startMutation.isError && <span className="text-sm text-red-500">Gagal mulai final test, coba lagi.</span>}
        </div>
      )}
      {phase === "intro" && modulesQuery.isLoading && <p className="text-sm text-zinc-500">Memuat…</p>}

      {phase === "quiz" && attempt && (
        <>
          <CourseQuiz key={attempt.attemptId} start={attempt} submitting={submitMutation.isPending} onSubmit={submit} />
          {submitMutation.isError && <p className="text-sm text-red-500">Gagal ngirim jawaban, coba lagi.</p>}
        </>
      )}

      {phase === "result" && attempt && result && (
        <CourseResult
          key={attempt.attemptId}
          start={attempt}
          result={result}
          answers={answers}
          actions={
            <>
              {result.passed && result.nextModuleUnlocked && nextModule && (
                <Link href="/coding/course" className={`${btn} bg-emerald-600 text-white hover:bg-emerald-700`}>
                  🔓 Module {nextModule.title} kebuka!
                </Link>
              )}
              <button
                type="button"
                onClick={start}
                disabled={startMutation.isPending}
                className={`${btn} bg-amber-500 text-white hover:bg-amber-600 disabled:opacity-60`}
              >
                {result.passed ? "Speedrun lagi" : "Coba lagi"}
              </button>
              <Link href="/coding/course" className={`${btn} border border-black/[.1] dark:border-white/[.2]`}>
                Daftar module
              </Link>
            </>
          }
        />
      )}
    </CourseShell>
  );
}
