"use client";

import { useParams, useRouter } from "next/navigation";
import { useEffect, useMemo, useState } from "react";
import { useAuthStore } from "@/store/auth-store";
import { useLogoutMutation } from "@/hooks/use-auth";
import { useAdminCurriculumQuery } from "@/hooks/use-admin-curriculum";
import {
  useAdminQuestionsQuery,
  useCreateQuestionMutation,
  useUpdateQuestionMutation,
  useDeleteQuestionMutation,
  type AdminQuestion,
  type AdminQuestionOption,
  type QuestionInput,
} from "@/hooks/use-admin-questions";
import { PLATFORM_ROLE_LABEL } from "@/lib/dummy-accounts";
import { BackButton } from "@/components/back-button";

const STATUS_LABEL: Record<string, string> = {
  draft: "Draft",
  published: "Terbit",
  archived: "Arsip",
};

export default function AdminQuestionsPage() {
  const router = useRouter();
  const params = useParams<{ challengeId: string }>();
  const session = useAuthStore((s) => s.session);
  const hasHydrated = useAuthStore((s) => s.hasHydrated);
  const logoutMutation = useLogoutMutation();

  const curriculumQuery = useAdminCurriculumQuery();
  const questionsQuery = useAdminQuestionsQuery(params.challengeId);

  const createMutation = useCreateQuestionMutation(params.challengeId);
  const updateMutation = useUpdateQuestionMutation(params.challengeId);
  const deleteMutation = useDeleteQuestionMutation(params.challengeId);

  const [showAddForm, setShowAddForm] = useState(false);
  const [editingId, setEditingId] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);

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

  const challenge = useMemo(() => {
    const flattened = (curriculumQuery.data ?? []).flatMap((tier) =>
      tier.batches.flatMap((batch) =>
        batch.challenges.map((c) => ({ ...c, tierName: tier.name, batchName: batch.name })),
      ),
    );
    return flattened.find((c) => c.id === params.challengeId) ?? null;
  }, [curriculumQuery.data, params.challengeId]);

  if (!hasHydrated || !session || session.area !== "admin") {
    return (
      <div className="flex min-h-screen items-center justify-center bg-zinc-50 dark:bg-black">
        <p className="text-zinc-500 dark:text-zinc-500">Memuat…</p>
      </div>
    );
  }

  const optionCount = challenge?.optionCount ?? 3;

  const handleCreate = async (input: QuestionInput) => {
    setError(null);
    try {
      await createMutation.mutateAsync(input);
      setShowAddForm(false);
    } catch (err) {
      setError(extractErrorMessage(err));
    }
  };

  const handleUpdate = async (questionId: string, input: QuestionInput) => {
    setError(null);
    try {
      await updateMutation.mutateAsync({ questionId, input });
      setEditingId(null);
    } catch (err) {
      setError(extractErrorMessage(err));
    }
  };

  const handleDelete = async (questionId: string) => {
    if (!window.confirm("Hapus soal ini?")) return;
    await deleteMutation.mutateAsync(questionId);
  };

  return (
    <div className="flex min-h-screen flex-col bg-zinc-50 font-sans dark:bg-black">
      <header className="flex items-center justify-between px-6 py-5 sm:px-10">
        <div className="flex flex-col">
          <span className="text-lg font-semibold tracking-tight text-black dark:text-zinc-50">
            MathQuest Admin
          </span>
          <span className="text-xs text-zinc-500 dark:text-zinc-500">
            {session.displayName} · {PLATFORM_ROLE_LABEL[session.platformRole]}
          </span>
        </div>
        <button
          type="button"
          onClick={async () => {
            await logoutMutation.mutateAsync();
            router.push("/login");
          }}
          className="rounded-full border border-black/[.08] px-4 py-1.5 text-sm font-medium text-zinc-600 transition-colors hover:bg-black/[.04] dark:border-white/[.145] dark:text-zinc-400 dark:hover:bg-[#1a1a1a]"
        >
          Keluar
        </button>
      </header>

      <main className="mx-auto flex w-full max-w-3xl flex-1 flex-col gap-6 px-6 py-8">
        <div className="flex flex-col gap-1">
          <div className="-ml-2 flex items-center gap-1">
            <BackButton href="/admin/kurikulum" label="Kembali ke Kurikulum" />
            <h1 className="text-2xl font-semibold tracking-tight text-black dark:text-zinc-50">{challenge ? challenge.name : "Soal"}</h1>
          </div>
          {challenge && (
            <p className="text-sm text-zinc-500 dark:text-zinc-500">
              {challenge.tierName} · {challenge.batchName} · {optionCount} opsi jawaban ·{" "}
              {challenge.isExam ? "ujian" : "challenge biasa"}
            </p>
          )}
        </div>

        {error && (
          <p className="rounded-lg bg-red-50 px-3 py-2 text-sm text-red-600 dark:bg-red-500/10 dark:text-red-400">
            {error}
          </p>
        )}

        {questionsQuery.isLoading && (
          <p className="text-sm text-zinc-500 dark:text-zinc-500">Memuat soal…</p>
        )}

        <div className="flex flex-col gap-3">
          {questionsQuery.data?.map((question) =>
            editingId === question.id ? (
              <QuestionForm
                key={question.id}
                optionCount={optionCount}
                initial={question}
                submitting={updateMutation.isPending}
                onCancel={() => setEditingId(null)}
                onSubmit={(input) => handleUpdate(question.id, input)}
              />
            ) : (
              <QuestionRow
                key={question.id}
                question={question}
                onEdit={() => setEditingId(question.id)}
                onDelete={() => handleDelete(question.id)}
              />
            ),
          )}
        </div>

        {showAddForm ? (
          <QuestionForm
            optionCount={optionCount}
            submitting={createMutation.isPending}
            onCancel={() => setShowAddForm(false)}
            onSubmit={handleCreate}
          />
        ) : (
          <button
            type="button"
            onClick={() => setShowAddForm(true)}
            className="self-start rounded-full bg-foreground px-5 py-2 text-sm font-medium text-background transition-colors hover:bg-[#383838] dark:hover:bg-[#ccc]"
          >
            + Tambah Soal
          </button>
        )}
      </main>
    </div>
  );
}

function extractErrorMessage(err: unknown): string {
  const message = (err as { response?: { data?: { message?: string } } })?.response?.data
    ?.message;
  return message ?? "Gagal menyimpan soal, coba lagi.";
}

function QuestionRow({
  question,
  onEdit,
  onDelete,
}: {
  question: AdminQuestion;
  onEdit: () => void;
  onDelete: () => void;
}) {
  return (
    <div className="flex flex-col gap-2 rounded-2xl border border-black/[.08] bg-white p-4 dark:border-white/[.145] dark:bg-zinc-900">
      <div className="flex items-start justify-between gap-3">
        <span className="font-medium text-black dark:text-zinc-50">{question.prompt}</span>
        <span
          className={`shrink-0 rounded-full px-2 py-0.5 text-[10px] font-semibold uppercase ${
            question.status === "published"
              ? "bg-green-100 text-green-700 dark:bg-green-500/10 dark:text-green-400"
              : question.status === "draft"
                ? "bg-zinc-100 text-zinc-600 dark:bg-zinc-800 dark:text-zinc-400"
                : "bg-red-100 text-red-700 dark:bg-red-500/10 dark:text-red-400"
          }`}
        >
          {STATUS_LABEL[question.status] ?? question.status}
        </span>
      </div>
      {question.type === "essay_numeric" ? (
        <div className="flex flex-wrap items-center gap-2">
          <span className="rounded-full bg-purple-100 px-2 py-0.5 text-[10px] font-semibold text-purple-700 dark:bg-purple-500/10 dark:text-purple-400">
            ✏️ Essay
          </span>
          <span className="rounded-full bg-green-100 px-3 py-1 text-xs font-medium text-green-700 dark:bg-green-500/10 dark:text-green-400">
            Jawaban: {question.correctAnswerValue} ✓
          </span>
        </div>
      ) : (
        <div className="flex flex-wrap gap-2">
          {question.options?.map((opt) => (
            <span
              key={opt.value}
              className={`rounded-full px-3 py-1 text-xs font-medium ${
                opt.isCorrect
                  ? "bg-green-100 text-green-700 dark:bg-green-500/10 dark:text-green-400"
                  : "bg-black/[.06] text-zinc-600 dark:bg-white/[.08] dark:text-zinc-400"
              }`}
            >
              {opt.value} {opt.isCorrect && "✓"}
            </span>
          ))}
        </div>
      )}
      <div className="flex gap-3 text-sm">
        <button type="button" onClick={onEdit} className="font-medium text-blue-600 dark:text-blue-400">
          Edit
        </button>
        <button type="button" onClick={onDelete} className="font-medium text-red-500">
          Hapus
        </button>
      </div>
    </div>
  );
}

function QuestionForm({
  optionCount,
  initial,
  submitting,
  onSubmit,
  onCancel,
}: {
  optionCount: number;
  initial?: AdminQuestion;
  submitting: boolean;
  onSubmit: (input: QuestionInput) => void;
  onCancel: () => void;
}) {
  const [type, setType] = useState<"multiple_choice" | "essay_numeric">(
    initial?.type ?? "multiple_choice",
  );
  const [prompt, setPrompt] = useState(initial?.prompt ?? "");
  const [status, setStatus] = useState<string>(initial?.status ?? "published");
  const [optionValues, setOptionValues] = useState<string[]>(() => {
    const base = initial?.options?.map((o) => String(o.value)) ?? [];
    return Array.from({ length: optionCount }, (_, i) => base[i] ?? "");
  });
  const [correctIndex, setCorrectIndex] = useState<number>(() => {
    const idx = initial?.options?.findIndex((o) => o.isCorrect) ?? -1;
    return idx >= 0 ? idx : 0;
  });
  const [essayAnswer, setEssayAnswer] = useState(
    initial?.correctAnswerValue !== undefined ? String(initial.correctAnswerValue) : "",
  );
  const [localError, setLocalError] = useState<string | null>(null);

  const handleSubmit = () => {
    setLocalError(null);
    if (!prompt.trim()) {
      setLocalError("Prompt soal gak boleh kosong.");
      return;
    }

    if (type === "essay_numeric") {
      const value = Number(essayAnswer);
      if (essayAnswer.trim() === "" || !Number.isFinite(value)) {
        setLocalError("Jawaban benar harus diisi angka.");
        return;
      }
      onSubmit({ type, prompt: prompt.trim(), status, options: [], correctAnswerValue: value });
      return;
    }

    const options: AdminQuestionOption[] = optionValues.map((raw, i) => ({
      value: Number(raw),
      isCorrect: i === correctIndex,
    }));
    if (options.some((o) => !Number.isFinite(o.value))) {
      setLocalError("Semua opsi harus diisi angka.");
      return;
    }
    const distinct = new Set(options.map((o) => o.value));
    if (distinct.size !== options.length) {
      setLocalError("Nilai opsi jangan ada yang sama.");
      return;
    }
    onSubmit({ type, prompt: prompt.trim(), status, options });
  };

  return (
    <div className="flex flex-col gap-3 rounded-2xl border border-blue-300 bg-white p-4 dark:border-blue-500/40 dark:bg-zinc-900">
      <label className="flex flex-col gap-1 text-sm">
        <span className="font-medium text-zinc-700 dark:text-zinc-300">Tipe soal</span>
        <select
          value={type}
          onChange={(e) => setType(e.target.value as "multiple_choice" | "essay_numeric")}
          className="rounded-lg border border-black/[.08] bg-transparent px-3 py-2 text-black outline-none focus:border-blue-500 dark:border-white/[.145] dark:text-zinc-50"
        >
          <option value="multiple_choice">Pilihan Ganda</option>
          <option value="essay_numeric">Essay (isian angka)</option>
        </select>
      </label>

      <label className="flex flex-col gap-1 text-sm">
        <span className="font-medium text-zinc-700 dark:text-zinc-300">Prompt soal</span>
        <input
          autoFocus
          value={prompt}
          onChange={(e) => setPrompt(e.target.value)}
          placeholder={type === "essay_numeric" ? "mis. Sebuah kelas punya 8 baris kursi..." : "mis. 7 + 8"}
          className="rounded-lg border border-black/[.08] bg-transparent px-3 py-2 text-black outline-none focus:border-blue-500 dark:border-white/[.145] dark:text-zinc-50"
        />
      </label>

      {type === "essay_numeric" ? (
        <label className="flex flex-col gap-1 text-sm">
          <span className="font-medium text-zinc-700 dark:text-zinc-300">Jawaban benar</span>
          <input
            type="number"
            value={essayAnswer}
            onChange={(e) => setEssayAnswer(e.target.value)}
            placeholder="mis. 32"
            className="rounded-lg border border-black/[.08] bg-transparent px-3 py-2 text-black outline-none focus:border-blue-500 dark:border-white/[.145] dark:text-zinc-50"
          />
        </label>
      ) : (
        <div className="flex flex-col gap-1.5 text-sm">
          <span className="font-medium text-zinc-700 dark:text-zinc-300">
            Opsi jawaban (pilih yang benar)
          </span>
          <div className="flex flex-col gap-2">
            {optionValues.map((value, i) => (
              <div key={i} className="flex items-center gap-2">
                <input
                  type="radio"
                  name={`correct-${initial?.id ?? "new"}`}
                  checked={correctIndex === i}
                  onChange={() => setCorrectIndex(i)}
                />
                <input
                  type="number"
                  value={value}
                  onChange={(e) =>
                    setOptionValues((prev) => prev.map((v, idx) => (idx === i ? e.target.value : v)))
                  }
                  placeholder={`Opsi ${i + 1}`}
                  className="flex-1 rounded-lg border border-black/[.08] bg-transparent px-3 py-1.5 text-sm text-black outline-none focus:border-blue-500 dark:border-white/[.145] dark:text-zinc-50"
                />
              </div>
            ))}
          </div>
        </div>
      )}

      <label className="flex flex-col gap-1 text-sm">
        <span className="font-medium text-zinc-700 dark:text-zinc-300">Status</span>
        <select
          value={status}
          onChange={(e) => setStatus(e.target.value)}
          className="rounded-lg border border-black/[.08] bg-transparent px-3 py-2 text-black outline-none focus:border-blue-500 dark:border-white/[.145] dark:text-zinc-50"
        >
          <option value="published">Terbit</option>
          <option value="draft">Draft</option>
          <option value="archived">Arsip</option>
        </select>
      </label>

      {localError && <p className="text-sm text-red-500">{localError}</p>}

      <div className="flex gap-2">
        <button
          type="button"
          onClick={handleSubmit}
          disabled={submitting}
          className="rounded-full bg-foreground px-4 py-1.5 text-sm font-medium text-background disabled:opacity-50"
        >
          {submitting ? "Menyimpan…" : "Simpan"}
        </button>
        <button
          type="button"
          onClick={onCancel}
          className="text-sm font-medium text-zinc-500 dark:text-zinc-500"
        >
          Batal
        </button>
      </div>
    </div>
  );
}
