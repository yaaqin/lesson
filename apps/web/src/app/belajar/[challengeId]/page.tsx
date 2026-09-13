"use client";

import Link from "next/link";
import { useParams } from "next/navigation";
import { useCallback, useEffect, useState } from "react";
import { findChallenge, type Question } from "@/lib/dummy-data";
import { usePlayerState } from "@/lib/player-state";

const QUESTION_TIME_SECONDS = 15;

function shuffle<T>(items: T[]): T[] {
  const copy = [...items];
  for (let i = copy.length - 1; i > 0; i--) {
    const j = Math.floor(Math.random() * (i + 1));
    [copy[i], copy[j]] = [copy[j], copy[i]];
  }
  return copy;
}

function buildSession(bank: Question[], count: number): Question[] {
  return shuffle(bank)
    .slice(0, count)
    .map((q) => ({ ...q, options: shuffle(q.options) }));
}

type AnswerRecord = {
  question: Question;
  selectedValue: number | null;
  isCorrect: boolean;
};

export default function ChallengePage() {
  const params = useParams<{ challengeId: string }>();
  const challenge = findChallenge(params.challengeId);
  const { lives, streak, recordAttempt } = usePlayerState();

  const [session, setSession] = useState<Question[] | null>(null);
  const [currentIndex, setCurrentIndex] = useState(0);
  const [answers, setAnswers] = useState<AnswerRecord[]>([]);
  const [selected, setSelected] = useState<number | null>(null);
  const [timeLeft, setTimeLeft] = useState(QUESTION_TIME_SECONDS);
  const [phase, setPhase] = useState<"playing" | "result">("playing");
  const [recorded, setRecorded] = useState(false);

  // "gacha fetch": ambil & acak soal sekali di awal sesi (FSD.md 3.4/3.5)
  // setState di-defer ke microtask biar gak sinkron di badan effect.
  useEffect(() => {
    if (!challenge) return;
    const id = setTimeout(
      () => setSession(buildSession(challenge.bank, challenge.questionCountRequired)),
      0,
    );
    return () => clearTimeout(id);
  }, [challenge]);

  const currentQuestion = session?.[currentIndex] ?? null;

  const goNext = useCallback(
    (question: Question, selectedValue: number | null, isCorrect: boolean) => {
      setAnswers((prev) => [...prev, { question, selectedValue, isCorrect }]);
      setSelected(null);
      setTimeLeft(QUESTION_TIME_SECONDS);
      setCurrentIndex((i) => i + 1);
    },
    [],
  );

  // countdown waktu jawab per soal
  useEffect(() => {
    if (phase !== "playing" || !currentQuestion) return;
    if (timeLeft <= 0) {
      const id = setTimeout(() => goNext(currentQuestion, null, false), 0);
      return () => clearTimeout(id);
    }
    const timer = setTimeout(() => setTimeLeft((s) => s - 1), 1000);
    return () => clearTimeout(timer);
  }, [timeLeft, phase, currentQuestion, goNext]);

  useEffect(() => {
    if (!session || phase !== "playing" || currentIndex < session.length) return;
    const id = setTimeout(() => setPhase("result"), 0);
    return () => clearTimeout(id);
  }, [currentIndex, session, phase]);

  const correctCount = answers.filter((a) => a.isCorrect).length;
  const scorePercent = session ? Math.round((correctCount / session.length) * 100) : 0;
  const passed = challenge ? scorePercent >= challenge.passThresholdPercent : false;

  // catat attempt sekali begitu hasil tampil: skor langsung kelihatan, streak langsung nyala,
  // nyawa berkurang kalau gagal (per kegagalan sesi, sesuai FSD.md 3.6)
  useEffect(() => {
    if (phase !== "result" || !challenge || recorded) return;
    const id = setTimeout(() => {
      recordAttempt(challenge.id, passed);
      setRecorded(true);
    }, 0);
    return () => clearTimeout(id);
  }, [phase, challenge, passed, recorded, recordAttempt]);

  const handleAnswer = (value: number, isCorrect: boolean) => {
    if (selected !== null || !currentQuestion) return;
    setSelected(value);
    const question = currentQuestion;
    window.setTimeout(() => goNext(question, value, isCorrect), 450);
  };

  const retry = () => {
    if (!challenge) return;
    setSession(buildSession(challenge.bank, challenge.questionCountRequired));
    setCurrentIndex(0);
    setAnswers([]);
    setSelected(null);
    setTimeLeft(QUESTION_TIME_SECONDS);
    setPhase("playing");
    setRecorded(false);
  };

  if (!challenge) {
    return (
      <div className="flex min-h-screen flex-col items-center justify-center gap-4 bg-zinc-50 px-6 text-center dark:bg-black">
        <p className="text-zinc-600 dark:text-zinc-400">
          Challenge tidak ditemukan.
        </p>
        <Link
          href="/belajar"
          className="text-sm font-medium text-blue-600 dark:text-blue-400"
        >
          Kembali ke daftar
        </Link>
      </div>
    );
  }

  if (!session) {
    return (
      <div className="flex min-h-screen items-center justify-center bg-zinc-50 dark:bg-black">
        <p className="text-zinc-500 dark:text-zinc-500">Menyiapkan soal…</p>
      </div>
    );
  }

  return (
    <div className="flex min-h-screen flex-col bg-zinc-50 font-sans dark:bg-black">
      <header className="flex items-center justify-between px-6 py-5 sm:px-10">
        <Link
          href="/belajar"
          className="text-sm font-medium text-zinc-500 hover:text-zinc-700 dark:text-zinc-500 dark:hover:text-zinc-300"
        >
          ← Keluar
        </Link>
        <div className="flex items-center gap-4 text-sm">
          <span className="flex items-center gap-1">
            {Array.from({ length: 3 }).map((_, i) => (
              <span
                key={i}
                className={i < lives ? "text-red-500" : "text-zinc-300 dark:text-zinc-700"}
              >
                ❤️
              </span>
            ))}
          </span>
          <span className="flex items-center gap-1 font-medium text-orange-500">
            🔥 {streak}
          </span>
        </div>
      </header>

      <main className="mx-auto flex w-full max-w-xl flex-1 flex-col gap-8 px-6 py-8">
        {phase === "playing" && currentQuestion ? (
          <>
            <div className="flex flex-col gap-2">
              <div className="flex items-center justify-between text-sm text-zinc-500 dark:text-zinc-500">
                <span>{challenge.name}</span>
                <span>
                  Soal {currentIndex + 1} / {session.length}
                </span>
              </div>
              <div className="h-2 w-full overflow-hidden rounded-full bg-black/[.06] dark:bg-white/[.08]">
                <div
                  className="h-full rounded-full bg-blue-600 transition-all dark:bg-blue-500"
                  style={{ width: `${((currentIndex + 1) / session.length) * 100}%` }}
                />
              </div>
            </div>

            <div className="flex flex-col items-center gap-6 py-8">
              <span
                className={`text-sm font-semibold ${
                  timeLeft <= 5 ? "text-red-500" : "text-zinc-400 dark:text-zinc-600"
                }`}
              >
                ⏱️ {timeLeft}s
              </span>
              <h2 className="text-4xl font-semibold tracking-tight text-black dark:text-zinc-50">
                {currentQuestion.prompt}
              </h2>
            </div>

            <div className="grid grid-cols-1 gap-3 sm:grid-cols-3">
              {currentQuestion.options.map((opt) => {
                const isSelected = selected === opt.value;
                const showCorrectness = selected !== null;
                return (
                  <button
                    key={opt.value}
                    type="button"
                    disabled={selected !== null}
                    onClick={() => handleAnswer(opt.value, opt.isCorrect)}
                    className={`rounded-2xl border-2 px-4 py-5 text-xl font-semibold transition-colors ${
                      showCorrectness && opt.isCorrect
                        ? "border-green-500 bg-green-50 text-green-700 dark:bg-green-500/10 dark:text-green-400"
                        : showCorrectness && isSelected
                          ? "border-red-500 bg-red-50 text-red-700 dark:bg-red-500/10 dark:text-red-400"
                          : "border-black/[.08] bg-white text-black hover:border-blue-400 dark:border-white/[.145] dark:bg-zinc-900 dark:text-zinc-50"
                    }`}
                  >
                    {opt.value}
                  </button>
                );
              })}
            </div>
          </>
        ) : (
          <div className="flex flex-col items-center gap-6 py-8 text-center">
            <span className="animate-bounce text-5xl">{passed ? "🎉" : "💪"}</span>

            <div className="flex flex-col gap-1">
              <span className="text-sm font-medium text-zinc-500 dark:text-zinc-500">
                {challenge.name}
              </span>
              <span className="text-5xl font-bold tracking-tight text-black dark:text-zinc-50">
                {scorePercent}%
              </span>
              <span
                className={`text-sm font-medium ${
                  passed ? "text-green-600 dark:text-green-400" : "text-red-500"
                }`}
              >
                {passed
                  ? "Lulus! Nilai minimal terpenuhi."
                  : `Belum lulus — minimal ${challenge.passThresholdPercent}%.`}
              </span>
            </div>

            <div className="flex items-center gap-6 rounded-2xl border border-black/[.08] bg-white px-6 py-4 dark:border-white/[.145] dark:bg-zinc-900">
              <div className="flex flex-col items-center gap-1">
                <span className="text-2xl">🔥</span>
                <span className="text-sm font-semibold text-orange-500">
                  Streak {streak}
                </span>
              </div>
              <div className="flex flex-col items-center gap-1">
                <span className="text-2xl">
                  {correctCount}/{session.length}
                </span>
                <span className="text-xs text-zinc-500 dark:text-zinc-500">
                  Jawaban benar
                </span>
              </div>
              <div className="flex flex-col items-center gap-1">
                <span className="flex gap-0.5 text-lg">
                  {Array.from({ length: 3 }).map((_, i) => (
                    <span
                      key={i}
                      className={i < lives ? "text-red-500" : "text-zinc-300 dark:text-zinc-700"}
                    >
                      ❤️
                    </span>
                  ))}
                </span>
                <span className="text-xs text-zinc-500 dark:text-zinc-500">Nyawa</span>
              </div>
            </div>

            <div className="flex w-full max-w-xs flex-col gap-3">
              {!passed && (
                <button
                  type="button"
                  onClick={retry}
                  className="rounded-full bg-foreground px-6 py-3 text-sm font-medium text-background transition-colors hover:bg-[#383838] dark:hover:bg-[#ccc]"
                >
                  Ulangi Challenge
                </button>
              )}
              <Link
                href="/belajar"
                className={`rounded-full px-6 py-3 text-center text-sm font-medium transition-colors ${
                  passed
                    ? "bg-foreground text-background hover:bg-[#383838] dark:hover:bg-[#ccc]"
                    : "border border-black/[.08] text-zinc-600 hover:bg-black/[.04] dark:border-white/[.145] dark:text-zinc-400 dark:hover:bg-[#1a1a1a]"
                }`}
              >
                Kembali ke Daftar
              </Link>
            </div>

            <ol className="flex w-full flex-col gap-2 text-left text-sm">
              {answers.map((a, idx) => (
                <li
                  key={a.question.id}
                  className="flex items-center justify-between rounded-xl border border-black/[.08] bg-white px-4 py-2 dark:border-white/[.145] dark:bg-zinc-900"
                >
                  <span className="text-zinc-600 dark:text-zinc-400">
                    {idx + 1}. {a.question.prompt}
                  </span>
                  <span
                    className={
                      a.isCorrect ? "text-green-600 dark:text-green-400" : "text-red-500"
                    }
                  >
                    {a.selectedValue ?? "—"} {a.isCorrect ? "✓" : "✗"}
                  </span>
                </li>
              ))}
            </ol>
          </div>
        )}
      </main>
    </div>
  );
}
