"use client";

import Link from "next/link";
import { useParams, useRouter } from "next/navigation";
import { useEffect, useMemo, useState } from "react";
import { useAuthStore } from "@/store/auth-store";
import { useLogoutMutation } from "@/hooks/use-auth";
import { useAdminCategoriesQuery } from "@/hooks/use-admin-curriculum";
import {
  useAdminPuzzleQuestionsQuery,
  useCreatePuzzleQuestionMutation,
  useUpdatePuzzleQuestionMutation,
  useDeletePuzzleQuestionMutation,
  type AdminPuzzleQuestion,
  type PuzzleKind,
  type PuzzleUpsertInput,
} from "@/hooks/use-admin-puzzle-questions";
import { ConfirmDialog } from "@/components/confirm-dialog";
import { PLATFORM_ROLE_LABEL } from "@/lib/dummy-accounts";

const STATUS_LABEL: Record<string, string> = {
  draft: "Draft",
  published: "Terbit",
  archived: "Arsip",
};

const STATUS_BADGE: Record<string, string> = {
  published: "bg-green-100 text-green-700 dark:bg-green-500/10 dark:text-green-400",
  draft: "bg-zinc-100 text-zinc-600 dark:bg-zinc-800 dark:text-zinc-400",
  archived: "bg-red-100 text-red-700 dark:bg-red-500/10 dark:text-red-400",
};

export default function AdminPuzzleQuestionsPage() {
  const router = useRouter();
  const params = useParams<{ tierCode: string; challengeId: string }>();
  const session = useAuthStore((s) => s.session);
  const hasHydrated = useAuthStore((s) => s.hasHydrated);
  const logoutMutation = useLogoutMutation();

  const categoriesQuery = useAdminCategoriesQuery(params.tierCode);
  const questionsQuery = useAdminPuzzleQuestionsQuery(params.challengeId);

  const createMutation = useCreatePuzzleQuestionMutation(params.challengeId);
  const updateMutation = useUpdatePuzzleQuestionMutation(params.challengeId);
  const deleteMutation = useDeletePuzzleQuestionMutation(params.challengeId);

  const [showAddForm, setShowAddForm] = useState(false);
  const [newKind, setNewKind] = useState<PuzzleKind>("addition_grid");
  const [editingId, setEditingId] = useState<string | null>(null);
  const [deletingId, setDeletingId] = useState<string | null>(null);
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
    const flattened = (categoriesQuery.data ?? []).flatMap((category) =>
      category.challenges.map((c) => ({ ...c, categoryName: category.name })),
    );
    return flattened.find((c) => c.id === params.challengeId) ?? null;
  }, [categoriesQuery.data, params.challengeId]);

  if (!hasHydrated || !session || session.area !== "admin") {
    return (
      <div className="flex min-h-screen items-center justify-center bg-zinc-50 dark:bg-black">
        <p className="text-zinc-500 dark:text-zinc-500">Memuat…</p>
      </div>
    );
  }

  const existingKind = questionsQuery.data?.[0]?.kind;
  const activeKind = existingKind ?? newKind;

  const handleCreate = async (input: PuzzleUpsertInput) => {
    setError(null);
    try {
      await createMutation.mutateAsync(input);
      setShowAddForm(false);
    } catch (err) {
      setError(extractErrorMessage(err));
    }
  };

  const handleUpdate = async (questionId: string, input: PuzzleUpsertInput) => {
    setError(null);
    try {
      await updateMutation.mutateAsync({ questionId, input });
      setEditingId(null);
    } catch (err) {
      setError(extractErrorMessage(err));
    }
  };

  const confirmDelete = async () => {
    if (!deletingId) return;
    await deleteMutation.mutateAsync(deletingId);
    setDeletingId(null);
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
          <Link
            href={`/admin/kurikulum/kategori/${params.tierCode}`}
            className="text-sm font-medium text-blue-600 dark:text-blue-400"
          >
            ← Kembali ke Kategori
          </Link>
          <h1 className="text-2xl font-semibold tracking-tight text-black dark:text-zinc-50">
            {challenge ? challenge.name : "Soal Puzzle"}
          </h1>
          {challenge && (
            <p className="text-sm text-zinc-500 dark:text-zinc-500">
              {challenge.categoryName} ·{" "}
              {existingKind === "cryptarithm" ? "Cryptarithm" : "Addition Grid"}
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
              <PuzzleQuestionForm
                key={question.id}
                kind={question.kind}
                initial={question}
                submitting={updateMutation.isPending}
                onCancel={() => setEditingId(null)}
                onSubmit={(input) => handleUpdate(question.id, input)}
              />
            ) : (
              <PuzzleQuestionCard
                key={question.id}
                question={question}
                onEdit={() => setEditingId(question.id)}
                onDelete={() => setDeletingId(question.id)}
              />
            ),
          )}
        </div>

        {showAddForm ? (
          <div className="flex flex-col gap-3">
            {!existingKind && (
              <label className="flex items-center gap-2 text-sm">
                <span className="font-medium text-zinc-700 dark:text-zinc-300">Tipe puzzle</span>
                <select
                  value={newKind}
                  onChange={(e) => setNewKind(e.target.value as PuzzleKind)}
                  className="rounded-lg border border-black/[.08] bg-transparent px-2 py-1 text-black outline-none focus:border-blue-500 dark:border-white/[.145] dark:text-zinc-50"
                >
                  <option value="addition_grid">Addition Grid</option>
                  <option value="cryptarithm">Cryptarithm</option>
                </select>
              </label>
            )}
            <PuzzleQuestionForm
              kind={activeKind}
              submitting={createMutation.isPending}
              onCancel={() => setShowAddForm(false)}
              onSubmit={handleCreate}
            />
          </div>
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

      {deletingId && (
        <ConfirmDialog
          title="Hapus soal?"
          description="Soal ini bakal ilang permanen dari bank soal. Yakin lanjut?"
          confirmLabel="Ya, hapus"
          isPending={deleteMutation.isPending}
          onConfirm={confirmDelete}
          onCancel={() => setDeletingId(null)}
        />
      )}
    </div>
  );
}

function extractErrorMessage(err: unknown): string {
  const message = (err as { response?: { data?: { message?: string } } })?.response?.data
    ?.message;
  return message ?? "Gagal menyimpan soal, coba lagi.";
}

function PuzzleQuestionForm({
  kind,
  initial,
  submitting,
  onSubmit,
  onCancel,
}: {
  kind: PuzzleKind;
  initial?: AdminPuzzleQuestion;
  submitting: boolean;
  onSubmit: (input: PuzzleUpsertInput) => void;
  onCancel: () => void;
}) {
  return kind === "cryptarithm" ? (
    <CryptarithmForm initial={initial} submitting={submitting} onSubmit={onSubmit} onCancel={onCancel} />
  ) : (
    <AdditionGridForm initial={initial} submitting={submitting} onSubmit={onSubmit} onCancel={onCancel} />
  );
}

function PuzzleQuestionCard({
  question,
  onEdit,
  onDelete,
}: {
  question: AdminPuzzleQuestion;
  onEdit: () => void;
  onDelete: () => void;
}) {
  return (
    <div className="flex flex-col gap-2 rounded-2xl border border-black/[.08] bg-white p-4 dark:border-white/[.145] dark:bg-zinc-900">
      <div className="flex items-start justify-between gap-3">
        <span className="text-sm text-black dark:text-zinc-50">{question.prompt}</span>
        <span
          className={`shrink-0 rounded-full px-2 py-0.5 text-[10px] font-semibold uppercase ${STATUS_BADGE[question.status] ?? ""}`}
        >
          {STATUS_LABEL[question.status] ?? question.status}
        </span>
      </div>

      {question.kind === "addition_grid" && question.solutionGrid && question.givenMask ? (
        <div
          className="grid w-fit gap-1"
          style={{ gridTemplateColumns: `repeat(${question.size ?? 3}, 28px)` }}
        >
          {question.solutionGrid.map((row, r) =>
            row.map((v, c) => (
              <div
                key={`${r}-${c}`}
                className={`flex h-7 w-7 items-center justify-center rounded text-[11px] font-bold ${
                  question.givenMask?.[r]?.[c]
                    ? "bg-zinc-100 text-black dark:bg-zinc-800 dark:text-zinc-50"
                    : "bg-blue-50 text-blue-700 dark:bg-blue-500/10 dark:text-blue-400"
                }`}
              >
                {v}
              </div>
            )),
          )}
        </div>
      ) : (
        <div className="flex flex-col gap-1.5">
          <span className="font-mono text-sm text-zinc-700 dark:text-zinc-300">
            {question.words?.slice(0, -1).join(" + ")} = {question.words?.[(question.words?.length ?? 1) - 1]}
          </span>
          <div className="flex flex-wrap gap-1.5">
            {Object.entries(question.solution ?? {}).map(([letter, digit]) => (
              <span
                key={letter}
                className="rounded-full bg-black/[.06] px-2 py-0.5 text-xs font-medium text-zinc-600 dark:bg-white/[.08] dark:text-zinc-400"
              >
                {letter}={digit}
              </span>
            ))}
          </div>
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

type GridCell = { value: string; given: boolean };

function randomGrid(size: number): GridCell[][] {
  const values = Array.from({ length: size * size }, (_, i) => i + 1);
  for (let i = values.length - 1; i > 0; i--) {
    const j = Math.floor(Math.random() * (i + 1));
    [values[i], values[j]] = [values[j], values[i]];
  }
  const givenCount = Math.max(2, size);
  const givenIdx = new Set<number>();
  while (givenIdx.size < givenCount) {
    givenIdx.add(Math.floor(Math.random() * size * size));
  }
  const grid: GridCell[][] = [];
  for (let r = 0; r < size; r++) {
    const row: GridCell[] = [];
    for (let c = 0; c < size; c++) {
      const idx = r * size + c;
      row.push({ value: String(values[idx]), given: givenIdx.has(idx) });
    }
    grid.push(row);
  }
  return grid;
}

function gridFromInitial(initial?: AdminPuzzleQuestion): { size: number; grid: GridCell[][] } {
  if (initial?.solutionGrid && initial.givenMask) {
    const size = initial.size ?? initial.solutionGrid.length;
    return {
      size,
      grid: initial.solutionGrid.map((row, r) =>
        row.map((v, c) => ({ value: String(v), given: initial.givenMask?.[r]?.[c] ?? false })),
      ),
    };
  }
  return { size: 3, grid: randomGrid(3) };
}

function AdditionGridForm({
  initial,
  submitting,
  onSubmit,
  onCancel,
}: {
  initial?: AdminPuzzleQuestion;
  submitting: boolean;
  onSubmit: (input: PuzzleUpsertInput) => void;
  onCancel: () => void;
}) {
  const initialState = gridFromInitial(initial);
  const [size, setSize] = useState(initialState.size);
  const [grid, setGrid] = useState<GridCell[][]>(initialState.grid);
  const [prompt, setPrompt] = useState(initial?.prompt ?? "");
  const [status, setStatus] = useState<string>(initial?.status ?? "published");
  const [localError, setLocalError] = useState<string | null>(null);

  const resize = (n: number) => {
    setSize(n);
    setGrid(randomGrid(n));
  };

  const setCellValue = (r: number, c: number, value: string) => {
    setGrid((prev) => prev.map((row, ri) => row.map((cell, ci) => (ri === r && ci === c ? { ...cell, value } : cell))));
  };
  const toggleGiven = (r: number, c: number) => {
    setGrid((prev) =>
      prev.map((row, ri) => row.map((cell, ci) => (ri === r && ci === c ? { ...cell, given: !cell.given } : cell))),
    );
  };

  const numbers = grid.flat().map((cell) => Number(cell.value));
  const usedCount = new Map<number, number>();
  numbers.forEach((n) => usedCount.set(n, (usedCount.get(n) ?? 0) + 1));
  const isInvalidCell = (n: number) => !Number.isInteger(n) || n < 1 || n > size * size || (usedCount.get(n) ?? 0) > 1;

  const rowSums = grid.map((row) => row.reduce((sum, cell) => sum + (Number(cell.value) || 0), 0));
  const colSums = Array.from({ length: size }, (_, c) =>
    grid.reduce((sum, row) => sum + (Number(row[c].value) || 0), 0),
  );
  const blankCount = grid.flat().filter((cell) => !cell.given).length;
  const cellPx = size >= 5 ? 40 : 52;

  const handleSubmit = () => {
    setLocalError(null);
    const flatNums = grid.flat().map((c) => Number(c.value));
    if (flatNums.some((n) => !Number.isInteger(n) || n < 1 || n > size * size)) {
      setLocalError(`Semua kotak harus diisi angka 1..${size * size}.`);
      return;
    }
    if (new Set(flatNums).size !== flatNums.length) {
      setLocalError("Ada angka yang dobel -- tiap angka cuma boleh dipakai sekali.");
      return;
    }
    if (blankCount === 0) {
      setLocalError('Minimal 1 kotak harus dikosongin (uncheck "given") biar ada yang diisi user.');
      return;
    }
    onSubmit({
      kind: "addition_grid",
      status,
      prompt: prompt.trim() || undefined,
      size,
      solutionGrid: grid.map((row) => row.map((cell) => Number(cell.value))),
      givenMask: grid.map((row) => row.map((cell) => cell.given)),
    });
  };

  return (
    <div className="flex flex-col gap-3 rounded-2xl border border-blue-300 bg-white p-4 dark:border-blue-500/40 dark:bg-zinc-900">
      <div className="flex flex-wrap items-center gap-3">
        <label className="flex items-center gap-2 text-sm">
          <span className="font-medium text-zinc-700 dark:text-zinc-300">Ukuran</span>
          <select
            value={size}
            onChange={(e) => resize(Number(e.target.value))}
            className="rounded-lg border border-black/[.08] bg-transparent px-2 py-1 text-black outline-none focus:border-blue-500 dark:border-white/[.145] dark:text-zinc-50"
          >
            {[3, 4, 5, 6].map((n) => (
              <option key={n} value={n}>
                {n}x{n}
              </option>
            ))}
          </select>
        </label>
        <button
          type="button"
          onClick={() => setGrid(randomGrid(size))}
          className="rounded-full border border-black/[.08] px-3 py-1 text-xs font-medium text-zinc-600 dark:border-white/[.145] dark:text-zinc-400"
        >
          🎲 Acak
        </button>
        <span className="text-xs text-zinc-500 dark:text-zinc-500">
          Centang &ldquo;given&rdquo; buat kotak yang jadi petunjuk (gak diisi user).
        </span>
      </div>

      <div className="overflow-x-auto">
        <div className="grid w-fit gap-1.5" style={{ gridTemplateColumns: `repeat(${size + 1}, ${cellPx}px)` }}>
          {grid.map((row, r) => (
            <div key={r} className="contents">
              {row.map((cell, c) => {
                const invalid = isInvalidCell(Number(cell.value));
                return (
                  <div key={c} className="flex flex-col items-center gap-0.5">
                    <input
                      type="number"
                      value={cell.value}
                      onChange={(e) => setCellValue(r, c, e.target.value)}
                      className={`rounded-lg border-2 text-center text-sm font-bold outline-none ${
                        invalid
                          ? "border-red-500 text-red-600"
                          : cell.given
                            ? "border-black/[.08] bg-zinc-100 text-black dark:border-white/[.145] dark:bg-zinc-800 dark:text-zinc-50"
                            : "border-blue-400 text-blue-700 dark:text-blue-400"
                      }`}
                      style={{ width: cellPx, height: cellPx }}
                    />
                    <label className="flex items-center gap-1 text-[10px] text-zinc-500 dark:text-zinc-500">
                      <input type="checkbox" checked={cell.given} onChange={() => toggleGiven(r, c)} />
                      given
                    </label>
                  </div>
                );
              })}
              <div className="flex h-full items-center justify-center text-xs font-semibold text-amber-700 dark:text-amber-400">
                {rowSums[r]}
              </div>
            </div>
          ))}
          <div className="contents">
            {colSums.map((sum, c) => (
              <div key={c} className="text-center text-xs font-semibold text-amber-700 dark:text-amber-400">
                {sum}
              </div>
            ))}
            <div />
          </div>
        </div>
      </div>

      <label className="flex flex-col gap-1 text-sm">
        <span className="font-medium text-zinc-700 dark:text-zinc-300">Prompt (kosongin buat default)</span>
        <input
          value={prompt}
          onChange={(e) => setPrompt(e.target.value)}
          placeholder="Isi kotak kosong dengan angka 1–9…"
          className="rounded-lg border border-black/[.08] bg-transparent px-3 py-2 text-black outline-none focus:border-blue-500 dark:border-white/[.145] dark:text-zinc-50"
        />
      </label>

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
        <button type="button" onClick={onCancel} className="text-sm font-medium text-zinc-500 dark:text-zinc-500">
          Batal
        </button>
      </div>
    </div>
  );
}

function CryptarithmForm({
  initial,
  submitting,
  onSubmit,
  onCancel,
}: {
  initial?: AdminPuzzleQuestion;
  submitting: boolean;
  onSubmit: (input: PuzzleUpsertInput) => void;
  onCancel: () => void;
}) {
  const [addends, setAddends] = useState<string[]>(() => initial?.words?.slice(0, -1) ?? ["", ""]);
  const [result, setResult] = useState(() => initial?.words?.[(initial?.words?.length ?? 1) - 1] ?? "");
  const [prompt, setPrompt] = useState(initial?.prompt ?? "");
  const [status, setStatus] = useState<string>(initial?.status ?? "published");
  const [localError, setLocalError] = useState<string | null>(null);

  const handleSubmit = () => {
    setLocalError(null);
    const cleanedAddends = addends.map((w) => w.trim().toUpperCase()).filter(Boolean);
    const cleanedResult = result.trim().toUpperCase();
    if (cleanedAddends.length < 1 || !cleanedResult) {
      setLocalError("Minimal 1 kata addend + 1 kata hasil.");
      return;
    }
    if (!cleanedAddends.every((w) => /^[A-Z]+$/.test(w)) || !/^[A-Z]+$/.test(cleanedResult)) {
      setLocalError("Kata cuma boleh huruf A-Z, gak boleh kosong.");
      return;
    }
    onSubmit({
      kind: "cryptarithm",
      status,
      prompt: prompt.trim() || undefined,
      words: cleanedAddends,
      result: cleanedResult,
    });
  };

  return (
    <div className="flex flex-col gap-3 rounded-2xl border border-blue-300 bg-white p-4 dark:border-blue-500/40 dark:bg-zinc-900">
      <div className="flex flex-col gap-2 text-sm">
        <span className="font-medium text-zinc-700 dark:text-zinc-300">Kata (addend)</span>
        {addends.map((word, i) => (
          <div key={i} className="flex items-center gap-2">
            <input
              value={word}
              onChange={(e) =>
                setAddends((prev) => prev.map((w, idx) => (idx === i ? e.target.value.toUpperCase() : w)))
              }
              placeholder="mis. SEND"
              className="flex-1 rounded-lg border border-black/[.08] bg-transparent px-3 py-1.5 text-black outline-none focus:border-blue-500 dark:border-white/[.145] dark:text-zinc-50"
            />
            {addends.length > 1 && (
              <button
                type="button"
                onClick={() => setAddends((prev) => prev.filter((_, idx) => idx !== i))}
                className="text-xs font-medium text-red-500"
              >
                Hapus
              </button>
            )}
          </div>
        ))}
        <button
          type="button"
          onClick={() => setAddends((prev) => [...prev, ""])}
          className="w-fit text-xs font-medium text-blue-600 dark:text-blue-400"
        >
          + Tambah kata
        </button>
      </div>

      <label className="flex flex-col gap-1 text-sm">
        <span className="font-medium text-zinc-700 dark:text-zinc-300">Kata hasil</span>
        <input
          value={result}
          onChange={(e) => setResult(e.target.value.toUpperCase())}
          placeholder="mis. MONEY"
          className="rounded-lg border border-black/[.08] bg-transparent px-3 py-2 text-black outline-none focus:border-blue-500 dark:border-white/[.145] dark:text-zinc-50"
        />
      </label>

      <p className="text-xs text-zinc-500 dark:text-zinc-500">
        Solusi (mapping huruf → angka) dihitung otomatis pas disimpan — kalau kombinasi katanya
        gak solvable, penyimpanan bakal ditolak & pesan errornya ditampilin di atas.
      </p>

      <label className="flex flex-col gap-1 text-sm">
        <span className="font-medium text-zinc-700 dark:text-zinc-300">Prompt (kosongin buat default)</span>
        <input
          value={prompt}
          onChange={(e) => setPrompt(e.target.value)}
          className="rounded-lg border border-black/[.08] bg-transparent px-3 py-2 text-black outline-none focus:border-blue-500 dark:border-white/[.145] dark:text-zinc-50"
        />
      </label>

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
        <button type="button" onClick={onCancel} className="text-sm font-medium text-zinc-500 dark:text-zinc-500">
          Batal
        </button>
      </div>
    </div>
  );
}
