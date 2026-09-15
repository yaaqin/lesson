"use client";

// ConfirmDialog: popup konfirmasi generik buat aksi yang gak boleh kepencet gak
// sengaja (mis. reset nyawa user). Belum ada pola modal lain di dashboard ini,
// jadi dibikin reusable buat aksi destruktif berikutnya juga.
export function ConfirmDialog({
  title,
  description,
  confirmLabel = "Konfirmasi",
  cancelLabel = "Batal",
  isPending,
  onConfirm,
  onCancel,
}: {
  title: string;
  description: string;
  confirmLabel?: string;
  cancelLabel?: string;
  isPending?: boolean;
  onConfirm: () => void;
  onCancel: () => void;
}) {
  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50 p-4" onClick={onCancel}>
      <div
        onClick={(e) => e.stopPropagation()}
        className="w-full max-w-sm rounded-2xl bg-white p-5 dark:bg-zinc-900"
      >
        <h3 className="text-base font-semibold text-black dark:text-zinc-50">{title}</h3>
        <p className="mt-2 text-sm leading-relaxed text-zinc-600 dark:text-zinc-400">{description}</p>
        <div className="mt-5 flex justify-end gap-2">
          <button
            type="button"
            onClick={onCancel}
            disabled={isPending}
            className="rounded-full border border-black/[.08] px-4 py-1.5 text-sm font-medium text-zinc-600 transition-colors hover:bg-black/[.04] disabled:opacity-40 dark:border-white/[.145] dark:text-zinc-400 dark:hover:bg-[#1a1a1a]"
          >
            {cancelLabel}
          </button>
          <button
            type="button"
            onClick={onConfirm}
            disabled={isPending}
            className="rounded-full bg-red-600 px-4 py-1.5 text-sm font-medium text-white transition-colors hover:bg-red-700 disabled:opacity-40"
          >
            {isPending ? "Memproses…" : confirmLabel}
          </button>
        </div>
      </div>
    </div>
  );
}
