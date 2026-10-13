"use client";

import Link from "next/link";
import { useParams, useRouter } from "next/navigation";
import { useState } from "react";
import { CourseShell } from "@/components/course-shell";
import { CourseMaterial } from "@/components/course-material";
import { CourseQuiz, CourseResult } from "@/components/course-quiz";
import {
  useCourseLevelQuery,
  useStartCourseLevelMutation,
  useSubmitCourseAttemptMutation,
  type CourseAnswer,
  type CourseStartResult,
  type CourseSubmitResult,
} from "@/hooks/use-course";

// Alur level: baca materi -> latihan (soal acak dari bank level, harus 100%
// bener) -> pembahasan. Materi selalu bisa dibuka lagi dari hasil.
type Phase = "material" | "quiz" | "result";

export default function CourseLevelPage() {
  const router = useRouter();
  const { levelId } = useParams<{ levelId: string }>();
  const levelQuery = useCourseLevelQuery(levelId);
  const startMutation = useStartCourseLevelMutation();
  const submitMutation = useSubmitCourseAttemptMutation();

  const [phase, setPhase] = useState<Phase>("material");
  const [attempt, setAttempt] = useState<CourseStartResult | null>(null);
  const [answers, setAnswers] = useState<CourseAnswer[]>([]);
  const [result, setResult] = useState<CourseSubmitResult | null>(null);

  const level = levelQuery.data;

  const startQuiz = () =>
    startMutation.mutate(levelId, {
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

  return (
    <CourseShell
      backHref="/coding/course"
      backLabel="Kembali ke daftar module"
      title={level ? level.title : "Memuat…"}
      subtitle={level && `${level.moduleTitle} · Level ${level.levelNumber}${level.passed ? " · Lulus ✓" : ""}`}
    >
      {levelQuery.isError && (
        <p className="text-sm text-red-500">Level ini belum kebuka atau gak ketemu. Lulusin level sebelumnya dulu.</p>
      )}

      {level && phase === "material" && (
        <>
          <p className="rounded-2xl bg-sky-50 px-4 py-3 text-sm text-sky-800 dark:bg-sky-500/10 dark:text-sky-300">
            📖 {level.summary}
          </p>
          <CourseMaterial blocks={level.material} />
          <div className="mt-4 flex flex-col items-center gap-2 rounded-2xl border border-black/[.08] bg-white px-6 py-6 text-center dark:border-white/[.145] dark:bg-zinc-900">
            <span className="font-semibold text-black dark:text-zinc-50">Udah paham? Saatnya latihan</span>
            <span className="text-sm text-zinc-500 dark:text-zinc-400">
              {level.questionCount} soal acak · harus bener semua (100%) buat lulus level ini · bebas ngulang
            </span>
            <button
              type="button"
              onClick={startQuiz}
              disabled={startMutation.isPending}
              className={`${btn} mt-2 bg-emerald-600 text-white hover:bg-emerald-700 disabled:opacity-60`}
            >
              {startMutation.isPending ? "Menyiapkan soal…" : "Mulai Latihan"}
            </button>
            {startMutation.isError && <span className="text-sm text-red-500">Gagal mulai latihan, coba lagi.</span>}
          </div>
        </>
      )}

      {phase === "quiz" && attempt && (
        <>
          <button
            type="button"
            onClick={() => setPhase("material")}
            className="w-fit text-sm font-medium text-sky-600 dark:text-sky-400"
          >
            📖 Lihat materi lagi
          </button>
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
            result.passed ? (
              <>
                {result.nextLevelId && (
                  <button
                    type="button"
                    onClick={() => {
                      setPhase("material");
                      router.push(`/coding/course/level/${result.nextLevelId}`);
                    }}
                    className={`${btn} bg-emerald-600 text-white hover:bg-emerald-700`}
                  >
                    Level berikutnya →
                  </button>
                )}
                {result.examUnlocked && (
                  <Link
                    href={`/coding/course/ujian/${attempt.moduleCode}`}
                    className={`${btn} bg-amber-500 text-white hover:bg-amber-600`}
                  >
                    🏁 Final test kebuka!
                  </Link>
                )}
                <Link href="/coding/course" className={`${btn} border border-black/[.1] dark:border-white/[.2]`}>
                  Daftar module
                </Link>
              </>
            ) : (
              <>
                <button
                  type="button"
                  onClick={startQuiz}
                  disabled={startMutation.isPending}
                  className={`${btn} bg-emerald-600 text-white hover:bg-emerald-700 disabled:opacity-60`}
                >
                  Coba lagi
                </button>
                <button
                  type="button"
                  onClick={() => {
                    setPhase("material");
                    window.scrollTo({ top: 0 });
                  }}
                  className={`${btn} border border-black/[.1] dark:border-white/[.2]`}
                >
                  Baca materi lagi
                </button>
              </>
            )
          }
        />
      )}
    </CourseShell>
  );
}
