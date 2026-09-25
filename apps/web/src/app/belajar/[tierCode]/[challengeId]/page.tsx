"use client";

import Link from "next/link";
import { useParams, useRouter } from "next/navigation";
import { useCallback, useEffect, useRef, useState } from "react";
import { useAuthStore } from "@/store/auth-store";
import { usePuzzlePrefsStore } from "@/store/puzzle-prefs-store";
import {
  useMeQuery,
  useStartChallengeMutation,
  useSubmitAttemptMutation,
  type PuzzlePayload,
  type StartResult,
  type SubmitResult,
} from "@/hooks/use-curriculum";
import { useSecurityEventReporter, type SecurityEvent } from "@/hooks/use-security-event-reporter";
import { useVisibilityTracker } from "@/hooks/use-visibility-tracker";
import { useFullscreenGuard } from "@/hooks/use-fullscreen-guard";
import { QuestionGuard } from "@/components/question-guard";
import { NumericKeypad } from "@/components/numeric-keypad";

// "intro" = layar aturan sebelum mulai: attempt baru dibikin (dan fullscreen
// diminta) pas user klik "Mulai Challenge", karena requestFullscreen wajib
// dipanggil dari klik user.
type Phase = "intro" | "loading" | "playing" | "result" | "blocked";

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
  const [gridAnswers, setGridAnswers] = useState<Record<string, Record<string, number>>>({});
  const [essayDrafts, setEssayDrafts] = useState<Record<string, string>>({});
  const [selected, setSelected] = useState<number | null>(null);
  const [timeLeft, setTimeLeft] = useState(0);
  const [phase, setPhase] = useState<Phase>("intro");
  const [result, setResult] = useState<SubmitResult | null>(null);
  const [blockedReason, setBlockedReason] = useState<string | null>(null);
  const [awayCount, setAwayCount] = useState(0);
  const [showAwayWarning, setShowAwayWarning] = useState(false);

  const submittingRef = useRef(false);
  const currentIndexRef = useRef(0);

  useEffect(() => {
    if (!hasHydrated) return;
    if (session === null) router.replace("/login");
  }, [hasHydrated, session, router]);

  // ---------- anti-cheating ----------
  const security = useSecurityEventReporter(game?.attemptId ?? null);
  const { record: recordSecurityEvent, flush: flushSecurityEvents, drainForSubmit } = security;
  const getQuestionIndex = useCallback(() => currentIndexRef.current, []);

  const { closeOpenEpisode } = useVisibilityTracker({
    enabled: phase === "playing",
    getQuestionIndex,
    onAwayEnd: useCallback(
      (event: SecurityEvent) => {
        recordSecurityEvent(event);
        setAwayCount((n) => n + 1);
        setShowAwayWarning(true);
      },
      [recordSecurityEvent],
    ),
  });

  const fullscreen = useFullscreenGuard({
    active: phase === "playing",
    onExit: useCallback(() => {
      recordSecurityEvent({
        eventType: "fullscreen_exit",
        questionIndex: currentIndexRef.current,
        startedAt: new Date().toISOString(),
        durationMs: null,
      });
    }, [recordSecurityEvent]),
  });
  const { exit: exitFullscreen } = fullscreen;

  // Keluar fullscreen selagi ngerjain -> timer di-pause & soal ditutup overlay
  // sampai user balik ke fullscreen.
  const pausedForFullscreen = fullscreen.supported && phase === "playing" && !fullscreen.isFullscreen;

  const handleCopyBlocked = useCallback(() => {
    recordSecurityEvent({
      eventType: "copy_blocked",
      questionIndex: currentIndexRef.current,
      startedAt: new Date().toISOString(),
      durationMs: null,
    });
  }, [recordSecurityEvent]);

  // Batch event dikirim tiap ganti soal (selain tiap 5 event / 15 detik).
  useEffect(() => {
    currentIndexRef.current = currentIndex;
    flushSecurityEvents();
  }, [currentIndex, flushSecurityEvents]);

  useEffect(() => {
    if (phase === "result" || phase === "blocked") exitFullscreen();
  }, [phase, exitFullscreen]);

  const beginAttempt = useCallback(() => {
    submittingRef.current = false;
    setAwayCount(0);
    setShowAwayWarning(false);
    setPhase("loading");
    setCurrentIndex(0);
    setAnswers({});
    setGridAnswers({});
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

  // Dipanggil dari klik tombol (Mulai / Ulangi) -- requestFullscreen harus
  // jalan di dalam user gesture, jadi dipanggil sebelum request start.
  const startWithGuard = () => {
    fullscreen.enter();
    beginAttempt();
  };

  const submitWithAnswers = useCallback(
    async (finalAnswers: Record<string, number>, finalGridAnswers: Record<string, Record<string, number>>) => {
      if (!game || submittingRef.current) return;
      submittingRef.current = true;
      const payload = game.questions.map((q) => ({
        questionId: q.id,
        selectedValue: finalAnswers[q.id] ?? null,
        selectedGrid: finalGridAnswers[q.id],
      }));
      // Sisa event anti-cheating (termasuk episode "pergi" yang masih kebuka
      // kalau waktu habis selagi user di luar halaman) ikut body submit, biar
      // pasti kehitung di risk score server.
      closeOpenEpisode();
      const securityEvents = await drainForSubmit();
      submitMutation.mutate(
        { attemptId: game.attemptId, answers: payload, securityEvents },
        {
          onSuccess: (data) => {
            setResult(data);
            setPhase("result");
          },
        },
      );
    },
    [game, submitMutation, closeOpenEpisode, drainForSubmit],
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
        submitWithAnswers(finalAnswers, gridAnswers);
      } else {
        setTimeLeft(game.timeLimitSeconds);
      }
    },
    [game, currentIndex, answers, gridAnswers, submitWithAnswers],
  );

  // grid_puzzle (kotak addition / cryptarithm): submit sekali per puzzle (gak
  // ada auto-advance kayak MC/essay karena challenge puzzle selalu cuma 1 soal),
  // lanjutannya sama kayak advanceRegular -- geser currentIndex, submit kalau abis.
  const handleAnswerPuzzle = useCallback(
    (grid: Record<string, number>) => {
      if (!game) return;
      const q = game.questions[currentIndex];
      const finalGridAnswers = { ...gridAnswers, [q.id]: grid };
      setGridAnswers(finalGridAnswers);

      const nextIndex = currentIndex + 1;
      setCurrentIndex(nextIndex);
      if (nextIndex >= game.questions.length) {
        submitWithAnswers(answers, finalGridAnswers);
      } else {
        setTimeLeft(game.timeLimitSeconds);
      }
    },
    [game, currentIndex, answers, gridAnswers, submitWithAnswers],
  );

  // Timer per soal (challenge biasa)
  useEffect(() => {
    if (phase !== "playing" || !game || game.isExam || pausedForFullscreen) return;
    if (timeLeft <= 0) {
      const id = setTimeout(() => advanceRegular(null), 0);
      return () => clearTimeout(id);
    }
    const timer = setTimeout(() => setTimeLeft((s) => s - 1), 1000);
    return () => clearTimeout(timer);
  }, [timeLeft, phase, game, advanceRegular, pausedForFullscreen]);

  // Timer total (ujian) — jalan terus gak peduli lagi di soal mana
  useEffect(() => {
    if (phase !== "playing" || !game || !game.isExam || pausedForFullscreen) return;
    if (timeLeft <= 0) {
      const id = setTimeout(() => submitWithAnswers(answers, gridAnswers), 0);
      return () => clearTimeout(id);
    }
    const timer = setTimeout(() => setTimeLeft((s) => s - 1), 1000);
    return () => clearTimeout(timer);
  }, [timeLeft, phase, game, answers, gridAnswers, submitWithAnswers, pausedForFullscreen]);

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

  const retry = startWithGuard;

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

  if (phase === "intro") {
    return (
      <IntroView
        tierCode={params.tierCode}
        lives={meQuery.data?.livesRemaining ?? 0}
        fullscreenSupported={fullscreen.supported}
        onStart={startWithGuard}
      />
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
        {phase === "playing" && showAwayWarning && (
          <AwayWarning count={awayCount} onDismiss={() => setShowAwayWarning(false)} />
        )}

        {phase === "playing" && (
          <QuestionGuard onCopyBlocked={handleCopyBlocked}>
            {game.isExam && (
              <ExamView
                game={game}
                currentIndex={currentIndex}
                setCurrentIndex={setCurrentIndex}
                answers={answers}
                essayDrafts={essayDrafts}
                onEssayChange={handleEssayDraftChange}
                timeLeft={timeLeft}
                onAnswer={handleAnswerExam}
                onFinish={() => submitWithAnswers(answers, gridAnswers)}
              />
            )}

            {!game.isExam && currentQuestion && (
              <RegularView
                key={currentQuestion.id}
                game={game}
                currentIndex={currentIndex}
                currentQuestion={currentQuestion}
                timeLeft={timeLeft}
                selected={selected}
                onAnswer={handleAnswerRegular}
                onAnswerPuzzle={handleAnswerPuzzle}
              />
            )}

            {!game.isExam && !currentQuestion && (
              <div className="flex flex-1 items-center justify-center py-16">
                <p className="text-zinc-500 dark:text-zinc-500">Menghitung skor…</p>
              </div>
            )}
          </QuestionGuard>
        )}

        {pausedForFullscreen && <FullscreenPausedOverlay onResume={fullscreen.enter} />}

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
            gridAnswers={gridAnswers}
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
  onAnswerPuzzle,
}: {
  game: StartResult;
  currentIndex: number;
  currentQuestion: StartResult["questions"][number];
  timeLeft: number;
  selected: number | null;
  onAnswer: (value: number) => void;
  onAnswerPuzzle: (grid: Record<string, number>) => void;
}) {
  const [essayValue, setEssayValue] = useState("");
  const [showPuzzleInfo, setShowPuzzleInfo] = useState(false);
  const isEssay = currentQuestion.type === "essay_numeric";
  const isPuzzle = currentQuestion.type === "grid_puzzle";

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
        {isPuzzle && (
          <div className="flex items-center gap-2">
            <span className="rounded-full bg-purple-100 px-3 py-1 text-xs font-semibold text-purple-700 dark:bg-purple-500/10 dark:text-purple-400">
              🧩 Puzzle
            </span>
            <button
              type="button"
              onClick={() => setShowPuzzleInfo(true)}
              aria-label="Cara main"
              className="flex h-6 w-6 items-center justify-center rounded-full border border-purple-300 text-xs font-bold text-purple-700 hover:bg-purple-50 dark:border-purple-500/40 dark:text-purple-400 dark:hover:bg-purple-500/10"
            >
              ?
            </button>
          </div>
        )}
        <h2
          className={
            isPuzzle
              ? "text-center text-base text-zinc-600 dark:text-zinc-400"
              : isEssay
                ? "text-center text-2xl font-semibold tracking-tight text-black sm:text-4xl dark:text-zinc-50"
                : "text-4xl font-semibold tracking-tight text-black dark:text-zinc-50"
          }
        >
          {currentQuestion.prompt}
        </h2>
      </div>

      {isPuzzle && showPuzzleInfo && currentQuestion.puzzle && (
        <PuzzleInfoModal kind={currentQuestion.puzzle.kind} onClose={() => setShowPuzzleInfo(false)} />
      )}

      {isPuzzle && currentQuestion.puzzle ? (
        <PuzzleAnswerView puzzle={currentQuestion.puzzle} onSubmit={onAnswerPuzzle} />
      ) : isEssay ? (
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
            // Cuma opsi yang dipilih yang dikasih warna -- kalau salah, jawaban bener
            // sengaja gak di-highlight biar gak kebocoran kunci sebelum lanjut soal.
            return (
              <button
                key={opt.value}
                type="button"
                disabled={selected !== null}
                onClick={() => onAnswer(opt.value)}
                className={`rounded-2xl border-2 px-4 py-5 text-xl font-semibold transition-colors ${
                  isSelected && opt.isCorrect
                    ? "border-green-500 bg-green-50 text-green-700 dark:bg-green-500/10 dark:text-green-400"
                    : isSelected && !opt.isCorrect
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
        <h2
          className={
            isEssay
              ? "text-center text-2xl font-semibold tracking-tight text-black sm:text-4xl dark:text-zinc-50"
              : "text-4xl font-semibold tracking-tight text-black dark:text-zinc-50"
          }
        >
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
  gridAnswers,
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
  gridAnswers: Record<string, Record<string, number>>;
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

      {result.riskLevel !== "normal" && (
        <p className="max-w-sm rounded-2xl border border-amber-300 bg-amber-50 px-4 py-3 text-sm text-amber-800 dark:border-amber-500/30 dark:bg-amber-500/10 dark:text-amber-300">
          {result.riskLevel === "review"
            ? "⚠️ Hasil ini ditandai buat ditinjau, karena tercatat banyak aktivitas di luar halaman soal selama pengerjaan."
            : "⚠️ Hasil ini ditandai kurang meyakinkan, karena kamu beberapa kali ninggalin halaman soal selama pengerjaan."}
        </p>
      )}

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
          if (q.type === "grid_puzzle") {
            // Solusi puzzle sengaja gak pernah dikirim ke klien (lihat PuzzlePayload),
            // jadi benar/salahnya dibaca dari `passed` -- challenge puzzle selalu cuma
            // 1 soal & lulus minimal 100%, jadi passed = puzzle ini bener persis.
            const attempted = !!gridAnswers[q.id];
            return (
              <li
                key={q.id}
                className="flex items-center justify-between rounded-xl border border-black/[.08] bg-white px-4 py-2 dark:border-white/[.145] dark:bg-zinc-900"
              >
                <span className="text-zinc-600 dark:text-zinc-400">
                  {idx + 1}. {q.prompt}
                </span>
                <span className={passed ? "text-green-600 dark:text-green-400" : "text-red-500"}>
                  {attempted ? (passed ? "Benar ✓" : "Belum tepat ✗") : "— ✗"}
                </span>
              </li>
            );
          }

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

function IntroView({
  tierCode,
  lives,
  fullscreenSupported,
  onStart,
}: {
  tierCode: string;
  lives: number;
  fullscreenSupported: boolean;
  onStart: () => void;
}) {
  return (
    <div className="flex min-h-screen flex-col items-center justify-center gap-6 bg-zinc-50 px-6 font-sans dark:bg-black">
      <LivesBadge lives={lives} />
      <div className="flex w-full max-w-sm flex-col gap-4 rounded-2xl border border-black/[.08] bg-white p-6 dark:border-white/[.145] dark:bg-zinc-900">
        <h1 className="text-lg font-semibold text-black dark:text-zinc-50">Sebelum mulai</h1>
        <ul className="flex flex-col gap-2 text-sm leading-relaxed text-zinc-600 dark:text-zinc-400">
          {fullscreenSupported && (
            <li>🖥️ Soal dikerjain dalam layar penuh. Keluar layar penuh bikin timer berhenti dan soal ketutup sampai kamu balik.</li>
          )}
          <li>👀 Pindah tab, pindah aplikasi, atau buka sidebar/extension selama ngerjain bakal dicatat.</li>
          <li>📋 Soal gak bisa di-copy.</li>
          <li>⚠️ Kalau aktivitasnya kebanyakan, hasilmu bisa ditandai buat ditinjau.</li>
        </ul>
        <button
          type="button"
          onClick={onStart}
          className="rounded-full bg-foreground px-6 py-3 text-sm font-medium text-background transition-colors hover:bg-[#383838] dark:hover:bg-[#ccc]"
        >
          Mulai Challenge
        </button>
      </div>
      <Link href={`/belajar/${tierCode}`} className="text-sm font-medium text-zinc-500 dark:text-zinc-500">
        ← Kembali ke daftar
      </Link>
    </div>
  );
}

function AwayWarning({ count, onDismiss }: { count: number; onDismiss: () => void }) {
  return (
    <div className="flex items-start justify-between gap-3 rounded-2xl border border-red-300 bg-red-50 px-4 py-3 text-sm text-red-700 dark:border-red-500/30 dark:bg-red-500/10 dark:text-red-400">
      <span>
        Kamu barusan ninggalin halaman soal ({count}x). Aktivitas ini dicatat dan bisa bikin hasilmu ditandai.
      </span>
      <button type="button" onClick={onDismiss} aria-label="Tutup" className="shrink-0 font-bold">
        ✕
      </button>
    </div>
  );
}

// Nutup soal sepenuhnya (bukan cuma dim) selagi di luar fullscreen, biar timer
// yang di-pause gak bisa dipake buat mikir sambil liat soal.
function FullscreenPausedOverlay({ onResume }: { onResume: () => void }) {
  return (
    <div className="fixed inset-0 z-50 flex flex-col items-center justify-center gap-4 bg-zinc-50 px-6 text-center dark:bg-black">
      <span className="text-4xl">⏸️</span>
      <h2 className="text-lg font-semibold text-black dark:text-zinc-50">Challenge dijeda</h2>
      <p className="max-w-sm text-sm text-zinc-600 dark:text-zinc-400">
        Kamu keluar dari layar penuh. Balik ke layar penuh buat lanjut ngerjain. Keluar layar penuh lebih dari 2 kali bikin
        hasilmu ditandai.
      </p>
      <button
        type="button"
        onClick={onResume}
        className="rounded-full bg-foreground px-6 py-3 text-sm font-medium text-background"
      >
        Kembali ke layar penuh
      </button>
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

// PuzzleAnswerView: soal grid_puzzle selalu satu-satunya soal di sesi (lihat
// seed-puzzle), jadi gak ada auto-advance kayak MC/essay -- cuma satu tombol
// "Jawab" yang aktif begitu semua sel/huruf keisi, langsung submit ke server
// (solusinya emang gak pernah ada di klien buat dicek sendiri -- PuzzlePayload
// sengaja gak bawa itu).
function PuzzleAnswerView({
  puzzle,
  onSubmit,
}: {
  puzzle: PuzzlePayload;
  onSubmit: (grid: Record<string, number>) => void;
}) {
  const [values, setValues] = useState<Record<string, string>>({});
  const [selectedKey, setSelectedKey] = useState<string | null>(null);

  const setCell = (key: string, raw: string) => {
    setValues((prev) => ({ ...prev, [key]: raw }));
  };

  const allFilled = puzzle.blankKeys.every((key) => values[key] !== undefined && values[key] !== "");

  const submit = () => {
    if (!allFilled) return;
    const grid: Record<string, number> = {};
    for (const key of puzzle.blankKeys) {
      grid[key] = Number(values[key]);
    }
    onSubmit(grid);
  };

  return (
    <div className="flex flex-col items-center gap-6">
      {puzzle.kind === "addition_grid" ? (
        <AdditionGridAnswer
          puzzle={puzzle}
          values={values}
          selectedKey={selectedKey}
          onSelectKey={setSelectedKey}
          onCellChange={setCell}
        />
      ) : (
        <CryptarithmInput puzzle={puzzle} values={values} onCellChange={setCell} />
      )}
      <button
        type="button"
        disabled={!allFilled}
        onClick={submit}
        className="w-full max-w-xs rounded-full bg-foreground py-3 text-sm font-semibold text-background disabled:opacity-40"
      >
        Jawab
      </button>
    </div>
  );
}

// AdditionGridAnswer: gabungin kotak grid + keypad angka on-screen. Angka yang
// udah kepake di kotak lain otomatis di-disable di keypad, jadi user gak bisa
// gak sengaja masukin angka dobel (lihat curriculumsvc/puzzle.go -- solusinya
// emang harus tiap angka 1..N² dipakai tepat sekali, bukan cuma soal sum cocok).
function AdditionGridAnswer({
  puzzle,
  values,
  selectedKey,
  onSelectKey,
  onCellChange,
}: {
  puzzle: PuzzlePayload;
  values: Record<string, string>;
  selectedKey: string | null;
  onSelectKey: (key: string | null) => void;
  onCellChange: (key: string, raw: string) => void;
}) {
  const size = puzzle.size ?? 3;
  const maxNumber = size * size;
  const keyboardPosition = usePuzzlePrefsStore((s) => s.keyboardPosition);
  const setKeyboardPosition = usePuzzlePrefsStore((s) => s.setKeyboardPosition);

  const usedNumbers = new Set<number>();
  for (const v of Object.values(puzzle.given ?? {})) usedNumbers.add(v);
  for (const [key, raw] of Object.entries(values)) {
    if (key === selectedKey || raw === "") continue;
    usedNumbers.add(Number(raw));
  }

  const pick = (n: number) => {
    if (!selectedKey) return;
    onCellChange(selectedKey, String(n));
    const idx = puzzle.blankKeys.indexOf(selectedKey);
    const next = puzzle.blankKeys.slice(idx + 1).find((k) => !values[k]);
    onSelectKey(next ?? null);
  };

  const keypad = (
    <div className="flex flex-col items-center gap-3">
      <div className="hidden gap-1 rounded-full bg-black/[.04] p-1 text-xs md:flex dark:bg-white/[.06]">
        {(["left", "right"] as const).map((pos) => (
          <button
            key={pos}
            type="button"
            onClick={() => setKeyboardPosition(pos)}
            className={`rounded-full px-3 py-1 font-medium transition-colors ${
              keyboardPosition === pos
                ? "bg-white shadow-sm dark:bg-zinc-800 dark:text-zinc-50"
                : "text-zinc-500 dark:text-zinc-500"
            }`}
          >
            {pos === "left" ? "⬅ Kiri" : "Kanan ➡"}
          </button>
        ))}
      </div>
      <div className="grid w-fit grid-cols-5 gap-2">
        {Array.from({ length: maxNumber }, (_, i) => i + 1).map((n) => {
          const disabled = !selectedKey || usedNumbers.has(n);
          return (
            <button
              key={n}
              type="button"
              disabled={disabled}
              onClick={() => pick(n)}
              className="flex h-10 w-10 items-center justify-center rounded-lg border-2 border-black/[.08] bg-white text-sm font-bold text-black transition-colors hover:enabled:border-blue-400 disabled:cursor-not-allowed disabled:opacity-30 dark:border-white/[.145] dark:bg-zinc-900 dark:text-zinc-50"
            >
              {n}
            </button>
          );
        })}
      </div>
      <button
        type="button"
        disabled={!selectedKey}
        onClick={() => selectedKey && onCellChange(selectedKey, "")}
        className="rounded-lg border border-black/[.08] px-4 py-2 text-xs font-medium text-zinc-600 disabled:opacity-40 dark:border-white/[.145] dark:text-zinc-400"
      >
        ⌫ Hapus
      </button>
    </div>
  );

  return (
    <div
      className={`flex flex-col items-center gap-6 md:items-start ${
        keyboardPosition === "left" ? "md:flex-row-reverse" : "md:flex-row"
      }`}
    >
      <AdditionGridInput puzzle={puzzle} values={values} selectedKey={selectedKey} onSelectKey={onSelectKey} />
      {keypad}
    </div>
  );
}

function AdditionGridInput({
  puzzle,
  values,
  selectedKey,
  onSelectKey,
}: {
  puzzle: PuzzlePayload;
  values: Record<string, string>;
  selectedKey: string | null;
  onSelectKey: (key: string) => void;
}) {
  const size = puzzle.size ?? 3;
  const cellPx = size >= 5 ? 44 : size === 4 ? 52 : 64;
  const fontSize = size >= 5 ? 15 : 20;

  return (
    <div className="grid gap-1.5" style={{ gridTemplateColumns: `repeat(${size + 1}, ${cellPx}px)` }}>
      {Array.from({ length: size }).map((_, r) => (
        <div key={`row-${r}`} className="contents">
          {Array.from({ length: size }).map((_, c) => {
            const key = `r${r}c${c}`;
            const given = puzzle.given?.[key];
            if (given !== undefined) {
              return (
                <div
                  key={key}
                  className="flex items-center justify-center rounded-lg border-2 border-black/[.08] bg-zinc-100 font-bold text-black dark:border-white/[.145] dark:bg-zinc-800 dark:text-zinc-50"
                  style={{ width: cellPx, height: cellPx, fontSize }}
                >
                  {given}
                </div>
              );
            }
            const isSelected = selectedKey === key;
            return (
              <button
                key={key}
                type="button"
                onClick={() => onSelectKey(key)}
                className={`flex items-center justify-center rounded-lg border-2 text-center font-bold outline-none transition-colors ${
                  isSelected
                    ? "border-blue-600 bg-blue-100 text-blue-800 dark:bg-blue-500/20 dark:text-blue-300"
                    : "border-blue-400 bg-white text-blue-700 hover:border-blue-500 dark:bg-zinc-900 dark:text-blue-400"
                }`}
                style={{ width: cellPx, height: cellPx, fontSize }}
              >
                {values[key] ?? ""}
              </button>
            );
          })}
          <div
            className="flex items-center justify-center rounded-lg bg-amber-50 text-sm font-semibold text-amber-700 dark:bg-amber-500/10 dark:text-amber-400"
            style={{ width: cellPx, height: cellPx }}
          >
            {puzzle.rowSums?.[r]}
          </div>
        </div>
      ))}
      <div className="contents">
        {Array.from({ length: size }).map((_, c) => (
          <div
            key={`colsum-${c}`}
            className="flex items-center justify-center rounded-lg bg-amber-50 text-sm font-semibold text-amber-700 dark:bg-amber-500/10 dark:text-amber-400"
            style={{ width: cellPx, height: cellPx }}
          >
            {puzzle.colSums?.[c]}
          </div>
        ))}
        <div style={{ width: cellPx, height: cellPx }} />
      </div>
    </div>
  );
}

function PuzzleInfoModal({ kind, onClose }: { kind: PuzzlePayload["kind"]; onClose: () => void }) {
  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50 p-4" onClick={onClose}>
      <div
        onClick={(e) => e.stopPropagation()}
        className="w-full max-w-sm rounded-2xl bg-white p-5 dark:bg-zinc-900"
      >
        <div className="flex items-center justify-between">
          <h3 className="text-sm font-semibold text-black dark:text-zinc-50">Cara main</h3>
          <button
            type="button"
            onClick={onClose}
            aria-label="Tutup"
            className="flex h-7 w-7 items-center justify-center rounded-full text-zinc-500 hover:bg-black/[.06] dark:hover:bg-white/[.08]"
          >
            ✕
          </button>
        </div>
        <p className="mt-3 text-sm leading-relaxed text-zinc-600 dark:text-zinc-400">
          {kind === "addition_grid"
            ? "Isi tiap kotak kosong dengan angka 1 sampai jumlah kotaknya (misal 1-9 buat kotak 3x3) -- setiap angka cuma boleh dipakai TEPAT SEKALI di seluruh kotak, gak boleh ada yang dobel. Klik kotak kosong, lalu pilih angkanya di keypad -- angka yang udah kepake otomatis kekunci biar gak salah pilih. Jumlah angka di tiap baris & kolom harus sama persis dengan angka target di pinggirnya."
            : "Tiap huruf mewakili satu angka (0-9) yang nilainya tetap sama di semua kemunculannya, dan huruf beda harus angka beda. Isi digitnya supaya hasil penjumlahannya benar."}
        </p>
      </div>
    </div>
  );
}

function CryptarithmInput({
  puzzle,
  values,
  onCellChange,
}: {
  puzzle: PuzzlePayload;
  values: Record<string, string>;
  onCellChange: (key: string, raw: string) => void;
}) {
  const words = puzzle.words ?? [];
  const addends = words.slice(0, -1);
  const result = words[words.length - 1];

  return (
    <div className="flex w-full max-w-sm flex-col items-center gap-6">
      <div className="flex flex-col items-end gap-1 font-mono text-2xl font-bold tracking-widest text-black dark:text-zinc-50">
        {addends.map((w, i) => (
          <span key={i}>
            {i > 0 ? "+ " : ""}
            {w}
          </span>
        ))}
        <span className="w-full border-t-2 border-black/20 pt-1 text-right dark:border-white/20">{result}</span>
      </div>

      <div className="flex flex-wrap justify-center gap-2">
        {puzzle.blankKeys.map((letter) => (
          <div key={letter} className="flex flex-col items-center gap-1">
            <span className="text-xs font-semibold text-zinc-500 dark:text-zinc-500">{letter}</span>
            <input
              type="text"
              inputMode="numeric"
              maxLength={1}
              value={values[letter] ?? ""}
              onChange={(e) => onCellChange(letter, e.target.value.replace(/[^0-9]/g, "").slice(0, 1))}
              className="h-12 w-12 rounded-lg border-2 border-blue-400 bg-white text-center text-xl font-bold text-blue-700 outline-none focus:border-blue-600 dark:bg-zinc-900 dark:text-blue-400"
            />
          </div>
        ))}
      </div>
    </div>
  );
}
