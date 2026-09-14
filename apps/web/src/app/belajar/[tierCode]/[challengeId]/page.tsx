"use client";

import Link from "next/link";
import { useParams, useRouter } from "next/navigation";
import { useCallback, useEffect, useRef, useState } from "react";
import { useAuthStore } from "@/store/auth-store";
import {
  useMeQuery,
  useStartChallengeMutation,
  useSubmitAttemptMutation,
  type StartResult,
  type SubmitResult,
} from "@/hooks/use-curriculum";

type Phase = "loading" | "playing" | "result" | "blocked";

export default function ChallengePage() {
  const router = useRouter();
  const params = useParams<{ tierCode: string; challengeId: string }>();
  const session = useAuthStore((s) => s.session);
  const hasHydrated = useAuthStore((s) => s.hasHydrated);

  const startMutation = useStartChallengeMutation();
  const submitMutation = useSubmitAttemptMutation();
  const { mutate: startAttempt } = startMutation;
  const meQuery = useMeQuery();

  const [game, setGame] = useState<StartResult | null>(null);
  const [currentIndex, setCurrentIndex] = useState(0);
  const [answers, setAnswers] = useState<Record<string, number>>({});
  const [essayDrafts, setEssayDrafts] = useState<Record<string, string>>({});
  const [selected, setSelected] = useState<number | null>(null);
  const [timeLeft, setTimeLeft] = useState(0);
  const [phase, setPhase] = useState<Phase>("loading");
  const [result, setResult] = useState<SubmitResult | null>(null);
  const [blockedReason, setBlockedReason] = useState<string | null>(null);

  const startedForRef = useRef<string | null>(null);
  const submittingRef = useRef(false);

  useEffect(() => {
    if (!hasHydrated) return;
    if (session === null) router.replace("/login");
  }, [hasHydrated, session, router]);

  const beginAttempt = useCallback(() => {
    submittingRef.current = false;
    setPhase("loading");
    setCurrentIndex(0);
    setAnswers({});
    setEssayDrafts({});
    setSelected(null);
    setResult(null);
    startAttempt(params.challengeId, {
      onSuccess: (data) => {
        setGame(data);
        setTimeLeft(data.timeLimitSeconds);
        setPhase("playing");
      },
      onError: (err: unknown) => {
        const code =
          (err as { response?: { data?: { error?: string } } })?.response?.data?.error ??
          "internal_error";
        setBlockedReason(
          code === "no_lives"
            ? "Nyawa kamu habis. Tunggu reset besok, ya."
            : "Gagal memulai challenge, coba lagi.",
        );
        setPhase("blocked");
      },
    });
  }, [params.challengeId, startAttempt]);

  useEffect(() => {
    if (!session || startedForRef.current === params.challengeId) return;
    startedForRef.current = params.challengeId;
    beginAttempt();
  }, [session, params.challengeId, beginAttempt]);

  const submitWithAnswers = useCallback(
    (finalAnswers: Record<string, number>) => {
      if (!game || submittingRef.current) return;
      submittingRef.current = true;
      const payload = game.questions.map((q) => ({
        questionId: q.id,
        selectedValue: finalAnswers[q.id] ?? null,
      }));
      submitMutation.mutate(
        { attemptId: game.attemptId, answers: payload },
        {
          onSuccess: (data) => {
            setResult(data);
            setPhase("result");
          },
        },
      );
    },
    [game, submitMutation],
  );

  // Mode "challenge biasa": waktu per soal, auto-lanjut begitu waktu habis / jawab.
  const advanceRegular = useCallback(
    (selectedValue: number | null) => {
      if (!game) return;
      const q = game.questions[currentIndex];
      const finalAnswers = selectedValue === null ? answers : { ...answers, [q.id]: selectedValue };
      setAnswers(finalAnswers);
      setSelected(null);

      const nextIndex = currentIndex + 1;
      // currentIndex selalu digeser, termasuk pas soal terakhir -- biar currentQuestion
      // jadi undefined dan RegularView berhenti nampilin soal terakhir yang bisa
      // kepencet dobel selagi nunggu submitWithAnswers() kelar (race klik ganda).
      setCurrentIndex(nextIndex);
      if (nextIndex >= game.questions.length) {
        submitWithAnswers(finalAnswers);
      } else {
        setTimeLeft(game.timeLimitSeconds);
      }
    },
    [game, currentIndex, answers, submitWithAnswers],
  );

  // Timer per soal (challenge biasa)
  useEffect(() => {
    if (phase !== "playing" || !game || game.isExam) return;
    if (timeLeft <= 0) {
      const id = setTimeout(() => advanceRegular(null), 0);
      return () => clearTimeout(id);
    }
    const timer = setTimeout(() => setTimeLeft((s) => s - 1), 1000);
    return () => clearTimeout(timer);
  }, [timeLeft, phase, game, advanceRegular]);

  // Timer total (ujian) — jalan terus gak peduli lagi di soal mana
  useEffect(() => {
    if (phase !== "playing" || !game || !game.isExam) return;
    if (timeLeft <= 0) {
      const id = setTimeout(() => submitWithAnswers(answers), 0);
      return () => clearTimeout(id);
    }
    const timer = setTimeout(() => setTimeLeft((s) => s - 1), 1000);
    return () => clearTimeout(timer);
  }, [timeLeft, phase, game, answers, submitWithAnswers]);

  const handleAnswerRegular = (value: number) => {
    if (selected !== null) return;
    setSelected(value);
    window.setTimeout(() => advanceRegular(value), 450);
  };

  const handleAnswerExam = (questionId: string, value: number) => {
    setAnswers((prev) => ({ ...prev, [questionId]: value }));
  };

  // Mode ujian bebas bolak-balik antar soal, jadi draft ketikan essay per soal
  // disimpen di sini (bukan state lokal komponen) biar gak ilang pas pindah soal.
  const handleEssayDraftChange = (questionId: string, draft: string) => {
    setEssayDrafts((prev) => ({ ...prev, [questionId]: draft }));
    const parsed = Number(draft);
    if (draft !== "" && draft !== "-" && Number.isFinite(parsed)) {
      setAnswers((prev) => ({ ...prev, [questionId]: parsed }));
    } else {
      setAnswers((prev) => {
        const next = { ...prev };
        delete next[questionId];
        return next;
      });
    }
  };

  const retry = () => {
    startedForRef.current = null;
    beginAttempt();
  };

  if (!session) {
    return (
      <div className="flex min-h-screen items-center justify-center bg-zinc-50 dark:bg-black">
        <p className="text-zinc-500 dark:text-zinc-500">Memuat…</p>
      </div>
    );
  }

  if (phase === "blocked") {
    return (
      <div className="flex min-h-screen flex-col items-center justify-center gap-4 bg-zinc-50 px-6 text-center dark:bg-black">
        <span className="text-4xl">💔</span>
        <LivesBadge lives={meQuery.data?.livesRemaining ?? 0} />
        <p className="text-zinc-600 dark:text-zinc-400">{blockedReason}</p>
        <Link href={`/belajar/${params.tierCode}`} className="text-sm font-medium text-blue-600 dark:text-blue-400">
          Kembali ke daftar
        </Link>
      </div>
    );
  }

  if (phase === "loading" || !game) {
    return (
      <div className="flex min-h-screen items-center justify-center bg-zinc-50 dark:bg-black">
        <p className="text-zinc-500 dark:text-zinc-500">Menyiapkan soal…</p>
      </div>
    );
  }

  const currentQuestion = game.questions[currentIndex];
  const correctCount = result?.correctCount ?? 0;
  const passed = result?.passed ?? false;

  return (
    <div className="flex min-h-screen flex-col bg-zinc-50 font-sans dark:bg-black">
      <header className="flex items-center justify-between px-6 py-5 sm:px-10">
        <Link
          href={`/belajar/${params.tierCode}`}
          className="text-sm font-medium text-zinc-500 hover:text-zinc-700 dark:text-zinc-500 dark:hover:text-zinc-300"
        >
          ← Keluar
        </Link>
        <span className="text-sm font-medium text-zinc-600 dark:text-zinc-400">
          {game.challengeName}
        </span>
        <LivesBadge lives={meQuery.data?.livesRemaining ?? 0} />
      </header>

      <main className="mx-auto flex w-full max-w-xl flex-1 flex-col gap-8 px-6 py-8">
        {phase === "playing" && game.isExam && (
          <ExamView
            game={game}
            currentIndex={currentIndex}
            setCurrentIndex={setCurrentIndex}
            answers={answers}
            essayDrafts={essayDrafts}
            onEssayChange={handleEssayDraftChange}
            timeLeft={timeLeft}
            onAnswer={handleAnswerExam}
            onFinish={() => submitWithAnswers(answers)}
          />
        )}

        {phase === "playing" && !game.isExam && currentQuestion && (
          <RegularView
            key={currentQuestion.id}
            game={game}
            currentIndex={currentIndex}
            currentQuestion={currentQuestion}
            timeLeft={timeLeft}
            selected={selected}
            onAnswer={handleAnswerRegular}
          />
        )}

        {phase === "playing" && !game.isExam && !currentQuestion && (
          <div className="flex flex-1 items-center justify-center py-16">
            <p className="text-zinc-500 dark:text-zinc-500">Menghitung skor…</p>
          </div>
        )}

        {phase === "result" && result && (
          <ResultView
            challengeName={game.challengeName}
            result={result}
            passed={passed}
            correctCount={correctCount}
            totalQuestions={result.totalQuestions}
            passThresholdPercent={game.passThresholdPercent}
            tierCode={params.tierCode}
            onRetry={retry}
            questions={game.questions}
            answers={answers}
          />
        )}
      </main>
    </div>
  );
}

function RegularView({
  game,
  currentIndex,
  currentQuestion,
  timeLeft,
  selected,
  onAnswer,
}: {
  game: StartResult;
  currentIndex: number;
  currentQuestion: StartResult["questions"][number];
  timeLeft: number;
  selected: number | null;
  onAnswer: (value: number) => void;
}) {
  const [essayValue, setEssayValue] = useState("");
  const isEssay = currentQuestion.type === "essay_numeric";

  const submitEssay = () => {
    if (selected !== null || essayValue === "" || essayValue === "-") return;
    const num = Number(essayValue);
    if (!Number.isFinite(num)) return;
    onAnswer(num);
  };

  const essayFeedback: "correct" | "incorrect" | null =
    selected === null
      ? null
      : currentQuestion.correctAnswerValue !== undefined &&
          Math.abs(selected - currentQuestion.correctAnswerValue) < 1e-9
        ? "correct"
        : "incorrect";

  return (
    <>
      <div className="flex flex-col gap-2">
        <div className="flex items-center justify-between text-sm text-zinc-500 dark:text-zinc-500">
          <span>{game.challengeName}</span>
          <span>
            Soal {currentIndex + 1} / {game.questions.length}
          </span>
        </div>
        <div className="h-2 w-full overflow-hidden rounded-full bg-black/[.06] dark:bg-white/[.08]">
          <div
            className="h-full rounded-full bg-blue-600 transition-all dark:bg-blue-500"
            style={{ width: `${((currentIndex + 1) / game.questions.length) * 100}%` }}
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
        {isEssay && (
          <span className="rounded-full bg-purple-100 px-3 py-1 text-xs font-semibold text-purple-700 dark:bg-purple-500/10 dark:text-purple-400">
            ✏️ Isian — ketik jawabannya
          </span>
        )}
        <h2 className="text-4xl font-semibold tracking-tight text-black dark:text-zinc-50">
          {currentQuestion.prompt}
        </h2>
      </div>

      {isEssay ? (
        <div className="flex justify-center">
          <NumericKeypad
            value={selected !== null ? String(selected) : essayValue}
            onChange={setEssayValue}
            onSubmit={submitEssay}
            disabled={selected !== null}
            feedback={essayFeedback}
          />
        </div>
      ) : (
        <div className="grid grid-cols-1 gap-3 sm:grid-cols-3">
          {(currentQuestion.options ?? []).map((opt) => {
            const isSelected = selected === opt.value;
            const showCorrectness = selected !== null;
            return (
              <button
                key={opt.value}
                type="button"
                disabled={selected !== null}
                onClick={() => onAnswer(opt.value)}
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
      )}
    </>
  );
}

function ExamView({
  game,
  currentIndex,
  setCurrentIndex,
  answers,
  essayDrafts,
  onEssayChange,
  timeLeft,
  onAnswer,
  onFinish,
}: {
  game: StartResult;
  currentIndex: number;
  setCurrentIndex: (fn: (i: number) => number) => void;
  answers: Record<string, number>;
  essayDrafts: Record<string, string>;
  onEssayChange: (questionId: string, draft: string) => void;
  timeLeft: number;
  onAnswer: (questionId: string, value: number) => void;
  onFinish: () => void;
}) {
  const q = game.questions[currentIndex];
  const isEssay = q.type === "essay_numeric";
  const answeredCount = Object.keys(answers).length;
  const minutes = Math.floor(timeLeft / 60);
  const seconds = timeLeft % 60;

  return (
    <>
      <div className="flex flex-col gap-2 rounded-2xl border border-amber-300 bg-amber-50 p-4 dark:border-amber-500/30 dark:bg-amber-500/10">
        <div className="flex items-center justify-between">
          <span className="text-sm font-semibold text-amber-700 dark:text-amber-400">
            🏁 {game.challengeName} · mode ujian
          </span>
          <span
            className={`font-mono text-lg font-bold ${
              timeLeft <= 30 ? "text-red-500" : "text-amber-700 dark:text-amber-400"
            }`}
          >
            {minutes}:{seconds.toString().padStart(2, "0")}
          </span>
        </div>
        <span className="text-xs text-amber-700/80 dark:text-amber-400/80">
          Waktu ini buat semua {game.questions.length} soal — bebas pindah-pindah soal, submit
          kapan aja sebelum waktu habis.
        </span>
      </div>

      <div className="flex flex-wrap gap-2">
        {game.questions.map((question, i) => (
          <button
            key={question.id}
            type="button"
            onClick={() => setCurrentIndex(() => i)}
            className={`flex h-8 w-8 items-center justify-center rounded-full text-xs font-semibold transition-colors ${
              i === currentIndex
                ? "bg-blue-600 text-white dark:bg-blue-500"
                : answers[question.id] !== undefined
                  ? "bg-green-100 text-green-700 dark:bg-green-500/10 dark:text-green-400"
                  : "bg-black/[.06] text-zinc-500 dark:bg-white/[.08] dark:text-zinc-400"
            }`}
          >
            {i + 1}
          </button>
        ))}
      </div>

      <div className="flex flex-col items-center gap-6 py-6">
        <span className="text-sm text-zinc-500 dark:text-zinc-500">
          Soal {currentIndex + 1} / {game.questions.length}
        </span>
        {isEssay && (
          <span className="rounded-full bg-purple-100 px-3 py-1 text-xs font-semibold text-purple-700 dark:bg-purple-500/10 dark:text-purple-400">
            ✏️ Isian — ketik jawabannya
          </span>
        )}
        <h2 className="text-4xl font-semibold tracking-tight text-black dark:text-zinc-50">
          {q.prompt}
        </h2>
      </div>

      {isEssay ? (
        <div className="flex justify-center">
          <NumericKeypad value={essayDrafts[q.id] ?? ""} onChange={(v) => onEssayChange(q.id, v)} />
        </div>
      ) : (
        <div className="grid grid-cols-1 gap-3 sm:grid-cols-3">
          {(q.options ?? []).map((opt) => {
            const isSelected = answers[q.id] === opt.value;
            return (
              <button
                key={opt.value}
                type="button"
                onClick={() => onAnswer(q.id, opt.value)}
                className={`rounded-2xl border-2 px-4 py-5 text-xl font-semibold transition-colors ${
                  isSelected
                    ? "border-blue-600 bg-blue-50 text-blue-700 dark:bg-blue-500/10 dark:text-blue-400"
                    : "border-black/[.08] bg-white text-black hover:border-blue-400 dark:border-white/[.145] dark:bg-zinc-900 dark:text-zinc-50"
                }`}
              >
                {opt.value}
              </button>
            );
          })}
        </div>
      )}

      <div className="flex items-center justify-between">
        <button
          type="button"
          disabled={currentIndex === 0}
          onClick={() => setCurrentIndex((i) => Math.max(0, i - 1))}
          className="rounded-full border border-black/[.08] px-4 py-2 text-sm font-medium text-zinc-600 disabled:opacity-40 dark:border-white/[.145] dark:text-zinc-400"
        >
          ← Sebelumnya
        </button>
        <span className="text-xs text-zinc-500 dark:text-zinc-500">
          {answeredCount}/{game.questions.length} terjawab
        </span>
        {currentIndex === game.questions.length - 1 ? (
          <button
            type="button"
            onClick={onFinish}
            className="rounded-full bg-foreground px-5 py-2 text-sm font-medium text-background"
          >
            Selesai
          </button>
        ) : (
          <button
            type="button"
            onClick={() => setCurrentIndex((i) => Math.min(game.questions.length - 1, i + 1))}
            className="rounded-full border border-black/[.08] px-4 py-2 text-sm font-medium text-zinc-600 dark:border-white/[.145] dark:text-zinc-400"
          >
            Berikutnya →
          </button>
        )}
      </div>
    </>
  );
}

function ResultView({
  result,
  passed,
  correctCount,
  totalQuestions,
  passThresholdPercent,
  tierCode,
  onRetry,
  questions,
  answers,
}: {
  challengeName: string;
  result: SubmitResult;
  passed: boolean;
  correctCount: number;
  totalQuestions: number;
  passThresholdPercent: number;
  tierCode: string;
  onRetry: () => void;
  questions: StartResult["questions"];
  answers: Record<string, number>;
}) {
  return (
    <div className="flex flex-col items-center gap-6 py-8 text-center">
      <span className="animate-bounce text-5xl">{passed ? "🎉" : "💪"}</span>

      <div className="flex flex-col gap-1">
        <span className="text-5xl font-bold tracking-tight text-black dark:text-zinc-50">
          {result.scorePercent}%
        </span>
        <span
          className={`text-sm font-medium ${passed ? "text-green-600 dark:text-green-400" : "text-red-500"}`}
        >
          {passed ? "Lulus! Nilai minimal terpenuhi." : `Belum lulus — minimal ${passThresholdPercent}%.`}
        </span>
      </div>

      <div className="flex items-center gap-6 rounded-2xl border border-black/[.08] bg-white px-6 py-4 dark:border-white/[.145] dark:bg-zinc-900">
        <div className="flex flex-col items-center gap-1">
          <span className="text-2xl">🔥</span>
          <span className="text-sm font-semibold text-orange-500">Streak {result.currentStreak}</span>
        </div>
        <div className="flex flex-col items-center gap-1">
          <span className="text-2xl">
            {correctCount}/{totalQuestions}
          </span>
          <span className="text-xs text-zinc-500 dark:text-zinc-500">Jawaban benar</span>
        </div>
        <div className="flex flex-col items-center gap-1">
          <span className="flex gap-0.5 text-lg">
            {Array.from({ length: 3 }).map((_, i) => (
              <span key={i} className={i < result.livesRemaining ? "text-red-500" : "text-zinc-300 dark:text-zinc-700"}>
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
            onClick={onRetry}
            className="rounded-full bg-foreground px-6 py-3 text-sm font-medium text-background transition-colors hover:bg-[#383838] dark:hover:bg-[#ccc]"
          >
            Ulangi Challenge
          </button>
        )}
        <Link
          href={`/belajar/${tierCode}`}
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
        {questions.map((q, idx) => {
          const selectedValue = answers[q.id];
          const correctValue =
            q.type === "essay_numeric"
              ? q.correctAnswerValue
              : q.options?.find((o) => o.isCorrect)?.value;
          const isCorrect =
            selectedValue !== undefined &&
            correctValue !== undefined &&
            Math.abs(selectedValue - correctValue) < 1e-9;
          return (
            <li
              key={q.id}
              className="flex items-center justify-between rounded-xl border border-black/[.08] bg-white px-4 py-2 dark:border-white/[.145] dark:bg-zinc-900"
            >
              <span className="text-zinc-600 dark:text-zinc-400">
                {idx + 1}. {q.prompt}
              </span>
              <span className={isCorrect ? "text-green-600 dark:text-green-400" : "text-red-500"}>
                {selectedValue ?? "—"} {isCorrect ? "✓" : "✗"}
              </span>
            </li>
          );
        })}
      </ol>
    </div>
  );
}

function LivesBadge({ lives }: { lives: number }) {
  return (
    <span className="flex items-center gap-1">
      {Array.from({ length: 3 }).map((_, i) => (
        <span key={i} className={i < lives ? "text-red-500" : "text-zinc-300 dark:text-zinc-700"}>
          ❤️
        </span>
      ))}
    </span>
  );
}

const KEYPAD_KEYS = ["7", "8", "9", "4", "5", "6", "1", "2", "3", "-", "0", "."];

// Keypad angka on-screen buat soal essay_numeric -- sengaja bukan <input type="number">
// biar konsisten di semua device (gak gantung keyboard native OS) & gampang dikunci
// (disabled) begitu jawaban udah disubmit, kayak tombol opsi pilihan ganda.
function NumericKeypad({
  value,
  onChange,
  onSubmit,
  disabled,
  feedback,
}: {
  value: string;
  onChange: (value: string) => void;
  onSubmit?: () => void;
  disabled?: boolean;
  feedback?: "correct" | "incorrect" | null;
}) {
  const press = (key: string) => {
    if (disabled) return;
    if (key === "-") {
      onChange(value.startsWith("-") ? value.slice(1) : "-" + value);
      return;
    }
    if (key === "." && value.includes(".")) return;
    onChange(value + key);
  };

  return (
    <div className="flex w-full max-w-xs flex-col gap-3">
      <div
        className={`flex h-16 items-center justify-center rounded-2xl border-2 text-3xl font-semibold transition-colors ${
          feedback === "correct"
            ? "border-green-500 bg-green-50 text-green-700 dark:bg-green-500/10 dark:text-green-400"
            : feedback === "incorrect"
              ? "border-red-500 bg-red-50 text-red-700 dark:bg-red-500/10 dark:text-red-400"
              : "border-black/[.08] bg-white text-black dark:border-white/[.145] dark:bg-zinc-900 dark:text-zinc-50"
        }`}
      >
        {value || <span className="text-zinc-300 dark:text-zinc-700">0</span>}
      </div>

      <div className="grid grid-cols-3 gap-2">
        {KEYPAD_KEYS.map((key) => (
          <button
            key={key}
            type="button"
            disabled={disabled}
            onClick={() => press(key)}
            className="rounded-xl border border-black/[.08] bg-white py-4 text-xl font-semibold text-black transition-colors hover:border-blue-400 disabled:opacity-40 dark:border-white/[.145] dark:bg-zinc-900 dark:text-zinc-50"
          >
            {key}
          </button>
        ))}
      </div>

      <div className={onSubmit ? "grid grid-cols-2 gap-2" : "grid grid-cols-1 gap-2"}>
        <button
          type="button"
          disabled={disabled}
          onClick={() => onChange(value.slice(0, -1))}
          className="rounded-xl border border-black/[.08] py-3 text-sm font-medium text-zinc-600 disabled:opacity-40 dark:border-white/[.145] dark:text-zinc-400"
        >
          ⌫ Hapus
        </button>
        {onSubmit && (
          <button
            type="button"
            disabled={disabled || value === "" || value === "-"}
            onClick={onSubmit}
            className="rounded-xl bg-foreground py-3 text-sm font-semibold text-background disabled:opacity-40"
          >
            Jawab
          </button>
        )}
      </div>
    </div>
  );
}
