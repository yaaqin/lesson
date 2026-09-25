"use client";

import Link from "next/link";
import { useRouter } from "next/navigation";
import { useCallback, useEffect, useRef, useState } from "react";
import { useQueryClient } from "@tanstack/react-query";
import { useAuthStore } from "@/store/auth-store";
import { useRequireNickname } from "@/hooks/use-profile";
import {
  useAdventureQuery,
  useAnswerAdventureMutation,
  useRollbackAdventureMutation,
  useStartAdventureMutation,
  type AdventureAnswerResult,
  type AdventureHistoryItem,
  type AdventureQuestion,
  type AdventureStartResult,
  type AdventureState,
  type AdventureTierCode,
} from "@/hooks/use-adventure";
import { QuestionGuard } from "@/components/question-guard";
import { NumericKeypad } from "@/components/numeric-keypad";

type Phase = "lobby" | "loading" | "playing" | "passed" | "failed";

// Jeda setelah feedback benar/salah sebelum pindah soal (sama kayak challenge biasa).
const FEEDBACK_DELAY_MS = 450;

const TIER_LABEL: Record<AdventureTierCode, string> = {
  sd: "SD",
  smp: "SMP",
  smk: "SMK / SMA",
  kampus: "Kampus",
};

const TIER_BADGE: Record<AdventureTierCode, string> = {
  sd: "bg-sky-100 text-sky-700 dark:bg-sky-500/10 dark:text-sky-400",
  smp: "bg-emerald-100 text-emerald-700 dark:bg-emerald-500/10 dark:text-emerald-400",
  smk: "bg-amber-100 text-amber-700 dark:bg-amber-500/10 dark:text-amber-400",
  kampus: "bg-purple-100 text-purple-700 dark:bg-purple-500/10 dark:text-purple-400",
};

function formatDuration(totalSeconds: number) {
  const h = Math.floor(totalSeconds / 3600);
  const m = Math.floor((totalSeconds % 3600) / 60);
  const s = totalSeconds % 60;
  const mmss = `${m.toString().padStart(h > 0 ? 2 : 1, "0")}:${s.toString().padStart(2, "0")}`;
  return h > 0 ? `${h}:${mmss}` : mmss;
}

function errorCode(err: unknown) {
  return (err as { response?: { data?: { error?: string } } })?.response?.data?.error;
}

export default function AdventurePage() {
  const router = useRouter();
  const queryClient = useQueryClient();
  const session = useAuthStore((s) => s.session);
  const hasHydrated = useAuthStore((s) => s.hasHydrated);
  useRequireNickname();

  const adventureQuery = useAdventureQuery();
  const startMutation = useStartAdventureMutation();
  const answerMutation = useAnswerAdventureMutation();
  const rollbackMutation = useRollbackAdventureMutation();
  const { mutate: startRun } = startMutation;
  const { mutate: sendAnswer } = answerMutation;

  const [phase, setPhase] = useState<Phase>("lobby");
  const [run, setRun] = useState<AdventureStartResult | null>(null);
  const [index, setIndex] = useState(0);
  const [failsUsed, setFailsUsed] = useState(0);
  const [timeLeft, setTimeLeft] = useState(0);
  // pending = jawaban soal sekarang yang udah dikirim (null = waktu habis).
  const [pending, setPending] = useState<{ value: number | null } | null>(null);
  const [feedback, setFeedback] = useState<"correct" | "incorrect" | null>(null);
  const [sendFailed, setSendFailed] = useState(false);
  const [lastResult, setLastResult] = useState<AdventureAnswerResult | null>(null);
  const [notice, setNotice] = useState<string | null>(null);
  const [showExitModal, setShowExitModal] = useState(false);
  const [rollbackTarget, setRollbackTarget] = useState<number | null>(null);

  const advanceTimerRef = useRef<number | null>(null);

  useEffect(() => {
    if (!hasHydrated) return;
    if (session === null) router.replace("/login");
  }, [hasHydrated, session, router]);

  useEffect(() => {
    return () => {
      if (advanceTimerRef.current !== null) window.clearTimeout(advanceTimerRef.current);
    };
  }, []);

  // Selama main: tombol back browser dimatiin (history di-push ulang tiap
  // popstate) & refresh/tutup tab minta konfirmasi. Keluar cuma lewat tombol
  // Keluar + modal konfirmasi. Kalau user tetap maksa keluar (tutup tab dll),
  // attempt-nya ke-abandon pas start berikutnya dan lanjut dari awal CP.
  useEffect(() => {
    if (phase !== "playing") return;
    window.history.pushState(null, "", window.location.href);
    const onPopState = () => window.history.pushState(null, "", window.location.href);
    const onBeforeUnload = (e: BeforeUnloadEvent) => e.preventDefault();
    window.addEventListener("popstate", onPopState);
    window.addEventListener("beforeunload", onBeforeUnload);
    return () => {
      window.removeEventListener("popstate", onPopState);
      window.removeEventListener("beforeunload", onBeforeUnload);
    };
  }, [phase]);

  const backToLobby = useCallback(
    (message: string | null = null) => {
      if (advanceTimerRef.current !== null) window.clearTimeout(advanceTimerRef.current);
      setPhase("lobby");
      setRun(null);
      setPending(null);
      setFeedback(null);
      setSendFailed(false);
      setShowExitModal(false);
      setNotice(message);
      queryClient.invalidateQueries({ queryKey: ["adventure"] });
    },
    [queryClient],
  );

  const startCheckpoint = useCallback(() => {
    setNotice(null);
    setPhase("loading");
    startRun(undefined, {
      onSuccess: (data) => {
        setRun(data);
        setIndex(0);
        setFailsUsed(0);
        setPending(null);
        setFeedback(null);
        setSendFailed(false);
        setLastResult(null);
        setTimeLeft(data.questions[0]?.timeLimitSeconds ?? 0);
        setPhase("playing");
      },
      onError: (err) => {
        const code = errorCode(err);
        backToLobby(
          code === "adventure_completed"
            ? "Adventure udah kamu tamatin."
            : code === "not_enough_questions"
              ? "Bank soal belum cukup buat checkpoint ini."
              : "Gagal nyiapin checkpoint, coba lagi.",
        );
      },
    });
  }, [startRun, backToLobby]);

  const submitAnswer = useCallback(
    (value: number | null) => {
      if (!run) return;
      setPending({ value });
      setSendFailed(false);
      sendAnswer(
        { attemptId: run.attemptId, questionIndex: index, selectedValue: value },
        {
          onSuccess: (res) => {
            setFeedback(res.correct ? "correct" : "incorrect");
            setFailsUsed(res.failsUsed);
            advanceTimerRef.current = window.setTimeout(() => {
              advanceTimerRef.current = null;
              setFeedback(null);
              setPending(null);
              if (res.status === "in_progress") {
                setIndex(res.nextIndex);
                setTimeLeft(run.questions[res.nextIndex]?.timeLimitSeconds ?? 0);
                return;
              }
              setLastResult(res);
              setPhase(res.status === "passed" ? "passed" : "failed");
              queryClient.invalidateQueries({ queryKey: ["adventure"] });
            }, FEEDBACK_DELAY_MS);
          },
          onError: (err) => {
            const code = errorCode(err);
            if (code === "attempt_finished" || code === "attempt_not_found" || code === "question_index_mismatch") {
              backToLobby("Sesi checkpoint ini udah gak aktif (mungkin dibuka di tab lain). Mulai lagi dari awal checkpoint, ya.");
              return;
            }
            setSendFailed(true);
          },
        },
      );
    },
    [run, index, sendAnswer, backToLobby, queryClient],
  );

  // Timer per soal -- berhenti selagi jawaban dikirim / feedback tampil.
  useEffect(() => {
    if (phase !== "playing" || !run || pending !== null) return;
    if (timeLeft <= 0) {
      const id = window.setTimeout(() => submitAnswer(null), 0);
      return () => window.clearTimeout(id);
    }
    const id = window.setTimeout(() => setTimeLeft((s) => s - 1), 1000);
    return () => window.clearTimeout(id);
  }, [phase, run, pending, timeLeft, submitAnswer]);

  if (!hasHydrated || !session) {
    return <CenteredMessage text="Memuat…" />;
  }

  if (phase === "loading") {
    return <CenteredMessage text="Menyiapkan checkpoint…" />;
  }

  if (phase === "playing" && run) {
    const question = run.questions[index];
    return (
      <div className="flex min-h-screen flex-col bg-zinc-50 font-sans dark:bg-black">
        <header className="flex items-center justify-between px-4 py-4 sm:px-10">
          <button
            type="button"
            onClick={() => setShowExitModal(true)}
            className="rounded-full border border-black/[.08] px-4 py-1.5 text-sm font-medium text-zinc-600 transition-colors hover:bg-black/[.04] dark:border-white/[.145] dark:text-zinc-400 dark:hover:bg-[#1a1a1a]"
          >
            Keluar
          </button>
          <span className="text-sm font-semibold text-black dark:text-zinc-50">
            Checkpoint {run.checkpointNo}
            <span className="font-normal text-zinc-500"> / {run.totalCheckpoints}</span>
          </span>
          <FailsMeter used={failsUsed} max={run.maxFails} />
        </header>

        <main className="mx-auto flex w-full max-w-xl flex-1 flex-col gap-6 px-4 py-4 sm:px-6">
          {question && (
            <QuestionGuard>
              <AdventureQuestionView
                key={question.id + index}
                question={question}
                index={index}
                total={run.questions.length}
                globalNumber={run.questionOffset + index + 1}
                timeLeft={timeLeft}
                pending={pending}
                feedback={feedback}
                onAnswer={submitAnswer}
              />
            </QuestionGuard>
          )}

          {sendFailed && (
            <div className="flex items-center justify-between gap-3 rounded-2xl border border-red-300 bg-red-50 px-4 py-3 text-sm text-red-700 dark:border-red-500/30 dark:bg-red-500/10 dark:text-red-400">
              <span>Jawaban gagal terkirim. Cek koneksi kamu.</span>
              <button
                type="button"
                onClick={() => pending && submitAnswer(pending.value)}
                className="shrink-0 rounded-full bg-red-600 px-3 py-1 text-xs font-semibold text-white"
              >
                Kirim lagi
              </button>
            </div>
          )}
        </main>

        {showExitModal && (
          <ConfirmModal
            title="Keluar dari checkpoint?"
            body={`Progres di Checkpoint ${run.checkpointNo} bakal hilang. Nanti kamu lanjut lagi dari soal pertama Checkpoint ${run.checkpointNo}.`}
            confirmLabel="Ya, keluar"
            danger
            onCancel={() => setShowExitModal(false)}
            onConfirm={() => backToLobby()}
          />
        )}
      </div>
    );
  }

  if ((phase === "passed" || phase === "failed") && run && lastResult) {
    return (
      <CheckpointResultView
        run={run}
        result={lastResult}
        onNext={startCheckpoint}
        onLobby={() => backToLobby()}
      />
    );
  }

  const state = adventureQuery.data;

  return (
    <div className="flex min-h-screen flex-col bg-zinc-50 font-sans dark:bg-black">
      <header className="flex items-center justify-between px-4 py-5 sm:px-10">
        <Link
          href="/belajar"
          className="text-sm font-medium text-zinc-500 hover:text-zinc-700 dark:text-zinc-500 dark:hover:text-zinc-300"
        >
          ← Jenjang
        </Link>
        <span className="text-lg font-semibold tracking-tight text-black dark:text-zinc-50">Adventure</span>
        <span className="w-16" />
      </header>

      <main className="mx-auto flex w-full max-w-2xl flex-1 flex-col gap-6 px-4 pb-10 sm:px-6">
        {notice && (
          <div className="rounded-2xl border border-amber-300 bg-amber-50 px-4 py-3 text-sm text-amber-800 dark:border-amber-500/30 dark:bg-amber-500/10 dark:text-amber-300">
            {notice}
          </div>
        )}

        {adventureQuery.isLoading && <p className="text-sm text-zinc-500">Memuat progres…</p>}
        {adventureQuery.isError && <p className="text-sm text-red-500">Gagal memuat progres adventure.</p>}

        {state && (
          <>
            <ProgressCard state={state} onStart={startCheckpoint} />
            <CheckpointMap state={state} onPickRollback={setRollbackTarget} />
            <HistoryList state={state} onPickRollback={setRollbackTarget} />
          </>
        )}
      </main>

      {rollbackTarget !== null && state && (
        <ConfirmModal
          title={`Balik ke Checkpoint ${rollbackTarget}?`}
          body={`Kemajuan level kamu bakal berkurang: progres Checkpoint ${rollbackTarget}${
            state.currentCheckpoint - 1 > rollbackTarget ? `–${Math.min(state.currentCheckpoint - 1, state.totalCheckpoints)}` : ""
          } dibatalkan dan kamu harus ngerjain ulang mulai dari Checkpoint ${rollbackTarget}. History-nya tetap kesimpan.`}
          confirmLabel={rollbackMutation.isPending ? "Memproses…" : "Ya, balik"}
          danger
          disabled={rollbackMutation.isPending}
          onCancel={() => setRollbackTarget(null)}
          onConfirm={() =>
            rollbackMutation.mutate(rollbackTarget, {
              onSuccess: () => setRollbackTarget(null),
              onError: () => {
                setRollbackTarget(null);
                setNotice("Gagal balik ke checkpoint itu, coba lagi.");
              },
            })
          }
        />
      )}
    </div>
  );
}

function CenteredMessage({ text }: { text: string }) {
  return (
    <div className="flex min-h-screen items-center justify-center bg-zinc-50 dark:bg-black">
      <p className="text-zinc-500 dark:text-zinc-500">{text}</p>
    </div>
  );
}

function FailsMeter({ used, max }: { used: number; max: number }) {
  const remaining = Math.max(max - used, 0);
  return (
    <span className="flex items-center gap-1" title={`Sisa jatah salah: ${remaining} dari ${max}`}>
      {Array.from({ length: max }).map((_, i) => (
        <span
          key={i}
          className={`h-2.5 w-2.5 rounded-full ${i < remaining ? "bg-red-500" : "bg-zinc-300 dark:bg-zinc-700"}`}
        />
      ))}
    </span>
  );
}

function AdventureQuestionView({
  question,
  index,
  total,
  globalNumber,
  timeLeft,
  pending,
  feedback,
  onAnswer,
}: {
  question: AdventureQuestion;
  index: number;
  total: number;
  globalNumber: number;
  timeLeft: number;
  pending: { value: number | null } | null;
  feedback: "correct" | "incorrect" | null;
  onAnswer: (value: number | null) => void;
}) {
  const [essayValue, setEssayValue] = useState("");
  const isEssay = question.type === "essay_numeric";
  const locked = pending !== null;

  const submitEssay = () => {
    if (locked || essayValue === "" || essayValue === "-") return;
    const num = Number(essayValue);
    if (!Number.isFinite(num)) return;
    onAnswer(num);
  };

  return (
    <>
      <div className="flex flex-col gap-2">
        <div className="flex items-center justify-between text-sm text-zinc-500 dark:text-zinc-500">
          <span>
            Soal {index + 1} / {total}
          </span>
          <span>#{globalNumber.toLocaleString("id-ID")}</span>
        </div>
        <div className="h-2 w-full overflow-hidden rounded-full bg-black/[.06] dark:bg-white/[.08]">
          <div
            className="h-full rounded-full bg-blue-600 transition-all dark:bg-blue-500"
            style={{ width: `${((index + 1) / total) * 100}%` }}
          />
        </div>
      </div>

      <div className="flex flex-col items-center gap-4 py-4">
        <div className="flex items-center gap-2">
          <span className={`rounded-full px-3 py-1 text-xs font-semibold ${TIER_BADGE[question.tierCode]}`}>
            {TIER_LABEL[question.tierCode]}
          </span>
          <span
            className={`text-sm font-semibold ${timeLeft <= 5 ? "text-red-500" : "text-zinc-400 dark:text-zinc-600"}`}
          >
            ⏱️ {timeLeft}s
          </span>
        </div>
        {isEssay && (
          <span className="rounded-full bg-purple-100 px-3 py-1 text-xs font-semibold text-purple-700 dark:bg-purple-500/10 dark:text-purple-400">
            ✏️ Isian — ketik jawabannya
          </span>
        )}
        <h2 className="text-center text-2xl font-semibold tracking-tight break-words text-black sm:text-4xl dark:text-zinc-50">
          {question.prompt}
        </h2>
      </div>

      {isEssay ? (
        <div className="flex justify-center">
          <NumericKeypad
            value={pending?.value != null ? String(pending.value) : essayValue}
            onChange={setEssayValue}
            onSubmit={submitEssay}
            disabled={locked}
            feedback={feedback}
          />
        </div>
      ) : (
        <div className="grid grid-cols-2 gap-3">
          {(question.options ?? []).map((value) => {
            const isSelected = pending?.value === value;
            return (
              <button
                key={value}
                type="button"
                disabled={locked}
                onClick={() => onAnswer(value)}
                className={`rounded-2xl border-2 px-4 py-5 text-xl font-semibold break-all transition-colors ${
                  isSelected && feedback === "correct"
                    ? "border-green-500 bg-green-50 text-green-700 dark:bg-green-500/10 dark:text-green-400"
                    : isSelected && feedback === "incorrect"
                      ? "border-red-500 bg-red-50 text-red-700 dark:bg-red-500/10 dark:text-red-400"
                      : isSelected
                        ? "border-blue-500 bg-blue-50 text-blue-700 dark:bg-blue-500/10 dark:text-blue-400"
                        : "border-black/[.08] bg-white text-black hover:border-blue-400 dark:border-white/[.145] dark:bg-zinc-900 dark:text-zinc-50"
                }`}
              >
                {value}
              </button>
            );
          })}
        </div>
      )}

      {pending?.value === null && feedback === "incorrect" && (
        <p className="text-center text-sm font-medium text-red-500">Waktu habis!</p>
      )}
    </>
  );
}

function ProgressCard({ state, onStart }: { state: AdventureState; onStart: () => void }) {
  const totalQuestions = state.totalCheckpoints * state.questionsPerCheckpoint;
  const doneQuestions = Math.min(state.currentCheckpoint - 1, state.totalCheckpoints) * state.questionsPerCheckpoint;
  const percent = (doneQuestions / totalQuestions) * 100;

  return (
    <div className="flex flex-col gap-4 rounded-2xl border border-black/[.08] bg-white p-5 dark:border-white/[.145] dark:bg-zinc-900">
      <div className="flex items-baseline justify-between gap-3">
        <div className="flex flex-col">
          <span className="text-xs font-medium tracking-wide text-zinc-500 uppercase">Posisi kamu</span>
          <span className="text-2xl font-semibold tracking-tight text-black dark:text-zinc-50">
            {state.completed ? "Tamat 🏆" : `Checkpoint ${state.currentCheckpoint}`}
            {!state.completed && <span className="text-base font-normal text-zinc-500"> / {state.totalCheckpoints}</span>}
          </span>
        </div>
        <span className="text-sm text-zinc-500">
          {doneQuestions.toLocaleString("id-ID")} / {totalQuestions.toLocaleString("id-ID")} soal
        </span>
      </div>

      <div className="h-2 w-full overflow-hidden rounded-full bg-black/[.06] dark:bg-white/[.08]">
        <div className="h-full rounded-full bg-blue-600 dark:bg-blue-500" style={{ width: `${percent}%` }} />
      </div>

      <ul className="flex flex-col gap-1 text-sm leading-relaxed text-zinc-600 dark:text-zinc-400">
        <li>📦 Tiap checkpoint isinya {state.questionsPerCheckpoint} soal, campuran SD sampai Kampus.</li>
        <li>❤️ Jatah salah {state.maxFails}x per checkpoint (waktu habis juga dihitung salah). Lewat dari itu, checkpoint diulang dari awal.</li>
        <li>⏱️ Waktu tiap soal beda-beda, ngikutin jenjang asal soalnya.</li>
        <li>🚪 Keluar di tengah jalan = lanjut lagi dari soal pertama checkpoint terakhir.</li>
      </ul>

      {!state.completed && (
        <button
          type="button"
          onClick={onStart}
          className="rounded-full bg-foreground px-6 py-3 text-sm font-medium text-background transition-colors hover:bg-[#383838] dark:hover:bg-[#ccc]"
        >
          Mulai Checkpoint {state.currentCheckpoint}
        </button>
      )}
    </div>
  );
}

function CheckpointMap({
  state,
  onPickRollback,
}: {
  state: AdventureState;
  onPickRollback: (checkpoint: number) => void;
}) {
  return (
    <div className="flex flex-col gap-3">
      <h2 className="text-sm font-semibold text-black dark:text-zinc-50">Peta checkpoint</h2>
      <div className="grid grid-cols-8 gap-2 sm:grid-cols-10">
        {Array.from({ length: state.totalCheckpoints }).map((_, i) => {
          const no = i + 1;
          const passed = no < state.currentCheckpoint;
          const current = no === state.currentCheckpoint && !state.completed;
          return (
            <button
              key={no}
              type="button"
              disabled={!passed}
              onClick={() => onPickRollback(no)}
              title={passed ? `Balik ke Checkpoint ${no}` : current ? "Checkpoint sekarang" : "Belum kebuka"}
              className={`flex aspect-square items-center justify-center rounded-lg text-xs font-semibold transition-colors ${
                passed
                  ? "bg-green-100 text-green-700 hover:ring-2 hover:ring-green-400 dark:bg-green-500/10 dark:text-green-400"
                  : current
                    ? "bg-blue-600 text-white dark:bg-blue-500"
                    : "bg-black/[.06] text-zinc-400 dark:bg-white/[.08] dark:text-zinc-600"
              }`}
            >
              {no}
            </button>
          );
        })}
      </div>
      <p className="text-xs text-zinc-500">Klik checkpoint yang udah lulus buat balik ke sana.</p>
    </div>
  );
}

function HistoryList({
  state,
  onPickRollback,
}: {
  state: AdventureState;
  onPickRollback: (checkpoint: number) => void;
}) {
  return (
    <div className="flex flex-col gap-3">
      <h2 className="text-sm font-semibold text-black dark:text-zinc-50">History checkpoint</h2>
      {state.history.length === 0 ? (
        <p className="text-sm text-zinc-500">Belum ada checkpoint yang lulus.</p>
      ) : (
        <ul className="flex flex-col gap-2">
          {state.history.map((h) => (
            <HistoryRow
              key={h.attemptId}
              item={h}
              maxFails={state.maxFails}
              questionsPerCheckpoint={state.questionsPerCheckpoint}
              canRollback={!h.rolledBack && h.checkpointNo < state.currentCheckpoint}
              onRollback={() => onPickRollback(h.checkpointNo)}
            />
          ))}
        </ul>
      )}
    </div>
  );
}

function HistoryRow({
  item,
  maxFails,
  questionsPerCheckpoint,
  canRollback,
  onRollback,
}: {
  item: AdventureHistoryItem;
  maxFails: number;
  questionsPerCheckpoint: number;
  canRollback: boolean;
  onRollback: () => void;
}) {
  return (
    <li
      className={`flex flex-col gap-2 rounded-2xl border border-black/[.08] bg-white p-4 sm:flex-row sm:items-center sm:justify-between dark:border-white/[.145] dark:bg-zinc-900 ${
        item.rolledBack ? "opacity-60" : ""
      }`}
    >
      <div className="flex flex-col gap-1">
        <div className="flex items-center gap-2">
          <span className="font-semibold text-black dark:text-zinc-50">Checkpoint {item.checkpointNo}</span>
          {item.rolledBack && (
            <span className="rounded-full bg-black/[.06] px-2 py-0.5 text-[11px] font-medium text-zinc-500 dark:bg-white/[.08]">
              dibatalkan
            </span>
          )}
        </div>
        <div className="flex flex-wrap gap-x-3 gap-y-1 text-xs text-zinc-500">
          <span>⏱️ {formatDuration(item.durationSeconds)}</span>
          <span>
            ✅ {item.correctCount}/{questionsPerCheckpoint}
          </span>
          <span>
            ❤️ sisa jatah {item.failsRemaining}/{maxFails}
          </span>
          {item.failedAttempts > 0 && <span>💥 gagal {item.failedAttempts}x sebelum lulus</span>}
          <span>
            {new Date(item.completedAt).toLocaleString("id-ID", { dateStyle: "medium", timeStyle: "short" })}
          </span>
        </div>
      </div>
      {canRollback && (
        <button
          type="button"
          onClick={onRollback}
          className="self-start rounded-full border border-black/[.08] px-3 py-1.5 text-xs font-medium text-zinc-600 hover:bg-black/[.04] sm:self-auto dark:border-white/[.145] dark:text-zinc-400 dark:hover:bg-[#1a1a1a]"
        >
          Balik ke sini
        </button>
      )}
    </li>
  );
}

function CheckpointResultView({
  run,
  result,
  onNext,
  onLobby,
}: {
  run: AdventureStartResult;
  result: AdventureAnswerResult;
  onNext: () => void;
  onLobby: () => void;
}) {
  const passed = result.status === "passed";
  const nextLabel = passed ? `Lanjut ke Checkpoint ${run.checkpointNo + 1}` : `Ulangi Checkpoint ${run.checkpointNo}`;

  return (
    <div className="flex min-h-screen flex-col items-center justify-center gap-6 bg-zinc-50 px-4 text-center font-sans dark:bg-black">
      <span className="text-5xl">{result.adventureCompleted ? "🏆" : passed ? "✅" : "💥"}</span>
      <div className="flex flex-col gap-1">
        <h1 className="text-2xl font-semibold tracking-tight text-black dark:text-zinc-50">
          {result.adventureCompleted
            ? "Kamu namatin Adventure!"
            : passed
              ? `Checkpoint ${run.checkpointNo} lulus!`
              : "Jatah salah habis"}
        </h1>
        {!passed && (
          <p className="text-sm text-zinc-500">
            Kamu salah {result.failsUsed}x di Checkpoint {run.checkpointNo} (sampai soal ke-{result.nextIndex}). Ulang
            dari soal pertama, ya.
          </p>
        )}
      </div>

      <div className="grid w-full max-w-sm grid-cols-3 gap-3">
        <Stat label="Waktu" value={formatDuration(result.durationSeconds ?? 0)} />
        <Stat label="Benar" value={`${result.correctCount}/${run.questions.length}`} />
        <Stat label="Sisa jatah" value={`${result.failsRemaining}/${run.maxFails}`} />
      </div>

      <div className="flex w-full max-w-sm flex-col gap-2">
        {!result.adventureCompleted && (
          <button
            type="button"
            onClick={onNext}
            className="rounded-full bg-foreground px-6 py-3 text-sm font-medium text-background transition-colors hover:bg-[#383838] dark:hover:bg-[#ccc]"
          >
            {nextLabel}
          </button>
        )}
        <button
          type="button"
          onClick={onLobby}
          className="rounded-full border border-black/[.08] px-6 py-3 text-sm font-medium text-zinc-600 dark:border-white/[.145] dark:text-zinc-400"
        >
          Kembali ke peta
        </button>
      </div>
    </div>
  );
}

function Stat({ label, value }: { label: string; value: string }) {
  return (
    <div className="flex flex-col gap-1 rounded-2xl border border-black/[.08] bg-white px-3 py-4 dark:border-white/[.145] dark:bg-zinc-900">
      <span className="text-lg font-semibold text-black dark:text-zinc-50">{value}</span>
      <span className="text-xs text-zinc-500">{label}</span>
    </div>
  );
}

function ConfirmModal({
  title,
  body,
  confirmLabel,
  danger,
  disabled,
  onCancel,
  onConfirm,
}: {
  title: string;
  body: string;
  confirmLabel: string;
  danger?: boolean;
  disabled?: boolean;
  onCancel: () => void;
  onConfirm: () => void;
}) {
  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50 p-4" onClick={onCancel}>
      <div
        role="dialog"
        aria-modal="true"
        onClick={(e) => e.stopPropagation()}
        className="flex w-full max-w-sm flex-col gap-4 rounded-2xl bg-white p-5 dark:bg-zinc-900"
      >
        <h3 className="text-base font-semibold text-black dark:text-zinc-50">{title}</h3>
        <p className="text-sm leading-relaxed text-zinc-600 dark:text-zinc-400">{body}</p>
        <div className="grid grid-cols-2 gap-2">
          <button
            type="button"
            onClick={onCancel}
            className="rounded-full border border-black/[.08] py-2.5 text-sm font-medium text-zinc-600 dark:border-white/[.145] dark:text-zinc-400"
          >
            Batal
          </button>
          <button
            type="button"
            disabled={disabled}
            onClick={onConfirm}
            className={`rounded-full py-2.5 text-sm font-semibold disabled:opacity-50 ${
              danger ? "bg-red-600 text-white hover:bg-red-700" : "bg-foreground text-background"
            }`}
          >
            {confirmLabel}
          </button>
        </div>
      </div>
    </div>
  );
}
