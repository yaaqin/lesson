"use client";

import { useCallback, useEffect, useRef, useState, type ReactNode } from "react";
import { CourseCode, RichText } from "@/components/course-material";
import { CuteDialog } from "@/components/cute-dialog";
import {
  formatDuration,
  type CourseAnswer,
  type CourseQuestion,
  type CourseStartResult,
  type CourseSubmitResult,
} from "@/hooks/use-course";

type Draft = { choice?: number; text?: string };

function isAnswered(q: CourseQuestion, d: Draft | undefined): boolean {
  if (!d) return false;
  if (q.kind === "multiple_choice") return d.choice !== undefined;
  return !!d.text?.trim();
}

// Opsi yang bentuknya kode / output (ada kurung, kutip, dst) dirender monospace.
function looksLikeCode(s: string): boolean {
  return /[{}[\]"'`;=()]|^[\w.$-]+$/.test(s);
}

// CourseQuiz: pengerjaan latihan level & final test. Jawaban baru dikirim ke
// server sekali di akhir (onSubmit); final test punya timer total yang
// otomatis submit pas habis.
export function CourseQuiz({
  start,
  submitting,
  onSubmit,
}: {
  start: CourseStartResult;
  submitting: boolean;
  onSubmit: (answers: CourseAnswer[]) => void;
}) {
  const [index, setIndex] = useState(0);
  const [drafts, setDrafts] = useState<Record<string, Draft>>({});
  const [remaining, setRemaining] = useState(start.timeLimitSeconds ?? 0);
  const [confirmOpen, setConfirmOpen] = useState(false);
  const submittedRef = useRef(false);
  const questions = start.questions;
  const q = questions[index];
  const answeredCount = questions.filter((x) => isAnswered(x, drafts[x.id])).length;
  const unanswered = questions.map((x, i) => (isAnswered(x, drafts[x.id]) ? -1 : i)).filter((i) => i >= 0);

  const submit = useCallback(() => {
    if (submittedRef.current) return;
    submittedRef.current = true;
    onSubmit(
      questions.map((x) => ({
        questionId: x.id,
        choice: drafts[x.id]?.choice,
        text: drafts[x.id]?.text,
      })),
    );
  }, [drafts, onSubmit, questions]);

  // Timer final test: dihitung dari waktu mulai biar gak melenceng kalau tab
  // sempet di-background.
  const submitRef = useRef(submit);
  useEffect(() => {
    submitRef.current = submit;
  }, [submit]);
  useEffect(() => {
    if (!start.isExam || !start.timeLimitSeconds) return;
    const endsAt = Date.now() + start.timeLimitSeconds * 1000;
    const id = setInterval(() => {
      const left = Math.max(0, Math.ceil((endsAt - Date.now()) / 1000));
      setRemaining(left);
      if (left === 0) {
        clearInterval(id);
        submitRef.current();
      }
    }, 250);
    return () => clearInterval(id);
  }, [start.isExam, start.timeLimitSeconds]);

  const setDraft = (patch: Draft) => setDrafts((d) => ({ ...d, [q.id]: { ...d[q.id], ...patch } }));
  const draft = drafts[q.id];
  const isLast = index === questions.length - 1;

  return (
    <div className="flex flex-col gap-6">
      <div className="flex items-center justify-between gap-3">
        <div className="flex flex-wrap gap-1.5">
          {questions.map((x, i) => (
            <button
              key={x.id}
              type="button"
              onClick={() => setIndex(i)}
              aria-label={`Soal ${i + 1}`}
              className={`h-2.5 w-6 rounded-full transition-colors ${
                i === index
                  ? "bg-emerald-500"
                  : isAnswered(x, drafts[x.id])
                    ? "bg-emerald-500/40"
                    : "bg-black/[.1] dark:bg-white/[.15]"
              }`}
            />
          ))}
        </div>
        {start.isExam && (
          <span
            className={`shrink-0 rounded-full px-3 py-1 font-mono text-sm font-semibold ${
              remaining <= 30
                ? "bg-red-100 text-red-700 dark:bg-red-500/15 dark:text-red-400"
                : "bg-black/[.05] text-zinc-700 dark:bg-white/[.08] dark:text-zinc-300"
            }`}
          >
            ⏱ {formatDuration(remaining * 1000)}
          </span>
        )}
      </div>

      <div className="flex flex-col gap-3">
        <span className="text-xs font-medium text-zinc-500 dark:text-zinc-500">
          Soal {index + 1} dari {questions.length}
          {q.kind === "type_answer" && " · ketik nama tipenya"}
          {q.kind === "write_value" && " · tulis jawabanmu sendiri"}
        </span>
        <h2 className="text-lg leading-snug font-semibold text-black dark:text-zinc-50">
          <RichText text={q.prompt} />
        </h2>
        {q.code && <CourseCode code={q.code} />}
      </div>

      {q.kind === "multiple_choice" && (
        <div className="grid grid-cols-1 gap-3 sm:grid-cols-2">
          {(q.options ?? []).map((opt, i) => (
            <button
              key={i}
              type="button"
              onClick={() => setDraft({ choice: i })}
              className={`rounded-2xl border-2 px-4 py-4 text-left text-sm whitespace-pre-wrap transition-colors ${
                looksLikeCode(opt) ? "font-mono" : "font-medium"
              } ${
                draft?.choice === i
                  ? "border-emerald-500 bg-emerald-50 text-emerald-800 dark:bg-emerald-500/10 dark:text-emerald-300"
                  : "border-black/[.08] bg-white text-black hover:border-emerald-400 dark:border-white/[.145] dark:bg-zinc-900 dark:text-zinc-50"
              }`}
            >
              {opt}
            </button>
          ))}
        </div>
      )}

      {q.kind === "type_answer" && (
        <input
          key={q.id}
          autoFocus
          value={draft?.text ?? ""}
          onChange={(e) => setDraft({ text: e.target.value })}
          placeholder="mis. string"
          autoCapitalize="off"
          autoCorrect="off"
          spellCheck={false}
          className="rounded-2xl border-2 border-black/[.08] bg-white px-4 py-3 font-mono text-black outline-none focus:border-emerald-500 dark:border-white/[.145] dark:bg-zinc-900 dark:text-zinc-50"
        />
      )}

      {q.kind === "write_value" && (
        <div className="flex flex-col gap-2">
          {q.strictJson && (
            <span className="w-fit rounded-full bg-sky-100 px-2.5 py-0.5 text-xs font-semibold text-sky-700 dark:bg-sky-500/15 dark:text-sky-300">
              Format JSON — key & string wajib kutip dua
            </span>
          )}
          <div className="overflow-hidden rounded-2xl border-2 border-black/[.08] bg-zinc-900 focus-within:border-emerald-500 dark:border-white/[.145] dark:bg-zinc-950">
            {q.placeholder && (
              <div className="px-4 pt-3 font-mono text-sm text-zinc-500">{q.placeholder}</div>
            )}
            <textarea
              key={q.id}
              autoFocus
              value={draft?.text ?? ""}
              onChange={(e) => setDraft({ text: e.target.value })}
              onKeyDown={(e) => {
                if (e.key !== "Tab") return;
                e.preventDefault();
                const el = e.currentTarget;
                const { selectionStart: s, selectionEnd: end, value } = el;
                const next = value.slice(0, s) + "  " + value.slice(end);
                setDraft({ text: next });
                requestAnimationFrame(() => el.setSelectionRange(s + 2, s + 2));
              }}
              rows={Math.min(14, Math.max(4, (draft?.text ?? "").split("\n").length + 1))}
              placeholder={q.strictJson ? '{ "key": "value" }' : "tulis nilainya di sini…"}
              autoCapitalize="off"
              autoCorrect="off"
              spellCheck={false}
              className="w-full resize-none bg-transparent px-4 py-3 font-mono text-sm text-zinc-100 outline-none placeholder:text-zinc-600"
            />
          </div>
        </div>
      )}

      <div className="flex items-center justify-between gap-3">
        <button
          type="button"
          disabled={index === 0}
          onClick={() => setIndex((i) => i - 1)}
          className="rounded-full border border-black/[.08] px-4 py-2 text-sm font-medium text-zinc-600 disabled:opacity-40 dark:border-white/[.145] dark:text-zinc-400"
        >
          ← Sebelumnya
        </button>
        {isLast ? (
          <button
            type="button"
            disabled={submitting}
            onClick={() => (unanswered.length > 0 ? setConfirmOpen(true) : submit())}
            className="rounded-full bg-emerald-600 px-5 py-2 text-sm font-semibold text-white transition-colors hover:bg-emerald-700 disabled:opacity-60"
          >
            {submitting ? "Mengecek…" : `Kumpulkan (${answeredCount}/${questions.length})`}
          </button>
        ) : (
          <button
            type="button"
            onClick={() => setIndex((i) => i + 1)}
            className="rounded-full bg-foreground px-5 py-2 text-sm font-semibold text-background"
          >
            Lanjut →
          </button>
        )}
      </div>

      <CuteDialog
        open={confirmOpen}
        emoji="🙈"
        title="Eits, belum semua dijawab!"
        onClose={() => setConfirmOpen(false)}
        actions={
          <>
            <button
              type="button"
              data-autofocus
              onClick={() => {
                setConfirmOpen(false);
                setIndex(unanswered[0] ?? index);
              }}
              className="rounded-full bg-emerald-600 px-5 py-2.5 text-sm font-semibold text-white transition-transform hover:bg-emerald-700 active:scale-95"
            >
              Isi dulu deh 👀
            </button>
            <button
              type="button"
              onClick={() => {
                setConfirmOpen(false);
                submit();
              }}
              className="rounded-full px-5 py-2.5 text-sm font-semibold text-zinc-500 transition-colors hover:bg-black/[.04] hover:text-zinc-700 dark:text-zinc-400 dark:hover:bg-white/[.06]"
            >
              Kumpulin aja, aku yakin 🚀
            </button>
          </>
        }
      >
        <p>
          Masih ada <b className="text-black dark:text-zinc-50">{unanswered.length} soal</b> yang kosong. Soal kosong
          otomatis dihitung salah lho.
        </p>
        <div className="mt-3 flex flex-wrap justify-center gap-1.5">
          {unanswered.map((i) => (
            <button
              key={i}
              type="button"
              onClick={() => {
                setConfirmOpen(false);
                setIndex(i);
              }}
              className="rounded-full bg-amber-100 px-3 py-1 text-xs font-bold text-amber-700 transition-transform hover:scale-110 dark:bg-amber-500/15 dark:text-amber-300"
            >
              Soal {i + 1}
            </button>
          ))}
        </div>
      </CuteDialog>
    </div>
  );
}

// useReviewVisible: mode timed -> pembahasan lengkap tampil N detik, abis itu
// disembunyiin. Ngembaliin [masih tampil?, sisa detik].
function useReviewVisible(review: CourseSubmitResult["review"]): [boolean, number] {
  const timed = review.mode === "timed";
  const [left, setLeft] = useState(timed ? review.seconds : 0);
  useEffect(() => {
    if (!timed) return;
    const endsAt = Date.now() + review.seconds * 1000;
    const id = setInterval(() => {
      const next = Math.max(0, Math.ceil((endsAt - Date.now()) / 1000));
      setLeft(next);
      if (next === 0) clearInterval(id);
    }, 200);
    return () => clearInterval(id);
  }, [timed, review.seconds]);
  if (review.mode === "full") return [true, 0];
  if (review.mode === "result_only") return [false, 0];
  return [left > 0, left];
}

// CourseResult: skor + pembahasan per soal, tampilannya ngikutin setelan admin
// (result.review): full = jawaban + kunci + penjelasan, timed = sama tapi
// hilang setelah N detik, result_only = cuma benar/salah per soal.
// `actions` = tombol navigasi (ulang, level berikutnya, dst) dari pemanggil.
export function CourseResult({
  start,
  result,
  answers,
  actions,
}: {
  start: CourseStartResult;
  result: CourseSubmitResult;
  answers: CourseAnswer[];
  actions: ReactNode;
}) {
  const [showDetail, secondsLeft] = useReviewVisible(result.review);
  const answerOf = (q: CourseQuestion) => {
    const a = answers.find((x) => x.questionId === q.id);
    if (q.kind === "multiple_choice") return a?.choice !== undefined ? q.options?.[a.choice] : undefined;
    return a?.text?.trim() || undefined;
  };

  return (
    <div className="flex flex-col gap-6">
      <div
        className={`flex flex-col items-center gap-2 rounded-2xl border-2 px-6 py-8 text-center ${
          result.passed
            ? "border-emerald-500 bg-emerald-50 dark:bg-emerald-500/10"
            : "border-amber-400 bg-amber-50 dark:bg-amber-500/10"
        }`}
      >
        <span className="text-4xl">{result.passed ? "🎉" : "💪"}</span>
        <span className="text-2xl font-semibold text-black dark:text-zinc-50">
          {result.passed ? "Lulus!" : result.timedOut ? "Waktu habis" : "Belum lulus"}
        </span>
        <span className="text-sm text-zinc-600 dark:text-zinc-400">
          Benar {result.correctCount} dari {result.total} ({result.scorePercent}%) · butuh {result.passPercent}%
          {start.isExam && ` · waktu ${formatDuration(result.durationMs)}`}
        </span>
        {!result.passed && (
          <span className="text-sm text-zinc-600 dark:text-zinc-400">
            {start.isExam
              ? "Coba lagi — soalnya bakal diacak ulang."
              : "Level ini harus 100% bener. Ulang materinya kalau perlu, terus coba lagi."}
          </span>
        )}
        <div className="mt-2 flex flex-wrap justify-center gap-2">{actions}</div>
      </div>

      <div className="flex flex-wrap items-center justify-between gap-2">
        <h3 className="text-lg font-semibold text-black dark:text-zinc-50">Pembahasan</h3>
        {result.review.mode === "timed" && showDetail && (
          <span className="rounded-full bg-amber-100 px-3 py-1 text-xs font-semibold text-amber-700 tabular-nums dark:bg-amber-500/15 dark:text-amber-300">
            🙈 Pembahasan hilang dalam {secondsLeft} dtk
          </span>
        )}
      </div>
      {!showDetail && (
        <p className="-mt-3 text-sm text-zinc-500 dark:text-zinc-400">
          {result.review.mode === "timed"
            ? "Pembahasan lengkap udah disembunyiin — yang tersisa cuma benar / salah tiap soal."
            : "Cuma ditampilin benar / salah tiap soal. Bingung? Baca lagi materinya ya."}
        </p>
      )}
      <ol className="flex flex-col gap-3">
        {start.questions.map((q, i) => {
          const r = result.results.find((x) => x.questionId === q.id);
          const yours = answerOf(q);
          return (
            <li
              key={q.id}
              className={`flex flex-col gap-2 rounded-2xl border bg-white p-4 dark:bg-zinc-900 ${
                r?.correct ? "border-emerald-500/40" : "border-red-400/50"
              }`}
            >
              <div className="flex items-start justify-between gap-3">
                <span className="text-sm font-medium text-black dark:text-zinc-50">
                  {i + 1}. <RichText text={q.prompt} />
                </span>
                <span className={`shrink-0 text-sm font-semibold ${r?.correct ? "text-emerald-600" : "text-red-500"}`}>
                  {r?.correct ? "Benar ✓" : "Salah ✗"}
                </span>
              </div>
              {q.code && <CourseCode code={q.code} />}
              {showDetail && (
                <>
                  <div className="flex flex-col gap-1 text-sm">
                    <span className="text-zinc-500 dark:text-zinc-400">Jawabanmu:</span>
                    <pre className="overflow-x-auto rounded-lg bg-black/[.04] px-3 py-2 font-mono text-xs whitespace-pre-wrap text-black dark:bg-white/[.06] dark:text-zinc-100">
                      {yours ?? "— (kosong)"}
                    </pre>
                    {r?.feedback && <span className="text-red-600 dark:text-red-400">↳ {r.feedback}</span>}
                  </div>
                  {r && !r.correct && r.expected && (
                    <div className="flex flex-col gap-1 text-sm">
                      <span className="text-zinc-500 dark:text-zinc-400">
                        {q.kind === "write_value" ? "Contoh jawaban yang bener:" : "Jawaban yang bener:"}
                      </span>
                      <pre className="overflow-x-auto rounded-lg bg-emerald-500/[.08] px-3 py-2 font-mono text-xs whitespace-pre-wrap text-emerald-800 dark:text-emerald-300">
                        {r.expected}
                      </pre>
                    </div>
                  )}
                  {r?.explanation && (
                    <p className="text-sm text-zinc-600 dark:text-zinc-400">
                      💡 <RichText text={r.explanation} />
                    </p>
                  )}
                </>
              )}
            </li>
          );
        })}
      </ol>
    </div>
  );
}
