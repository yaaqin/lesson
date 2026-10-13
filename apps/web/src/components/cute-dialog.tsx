"use client";

import { useEffect, useRef, type ReactNode } from "react";

// CuteDialog: popup konfirmasi pengganti window.confirm -- kartu yang muncul
// membal + maskot emoji goyang (animasi di globals.css, mati otomatis kalau
// user nyalain reduce motion). Esc / klik di luar kartu = onClose.
export function CuteDialog({
  open,
  emoji,
  title,
  children,
  actions,
  onClose,
}: {
  open: boolean;
  emoji: string;
  title: string;
  children?: ReactNode;
  actions: ReactNode;
  onClose: () => void;
}) {
  const cardRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    if (!open) return;
    const onKey = (e: KeyboardEvent) => {
      if (e.key === "Escape") onClose();
    };
    window.addEventListener("keydown", onKey);
    // Fokus ke tombol pertama biar Enter / Space langsung kepake.
    cardRef.current?.querySelector<HTMLButtonElement>("[data-autofocus]")?.focus();
    return () => window.removeEventListener("keydown", onKey);
  }, [open, onClose]);

  if (!open) return null;

  return (
    <div
      className="cute-backdrop fixed inset-0 z-50 flex items-center justify-center bg-black/40 px-4 backdrop-blur-sm"
      onMouseDown={(e) => {
        if (e.target === e.currentTarget) onClose();
      }}
    >
      <div
        ref={cardRef}
        role="dialog"
        aria-modal="true"
        aria-labelledby="cute-dialog-title"
        className="cute-pop relative flex w-full max-w-sm flex-col items-center gap-3 rounded-3xl border-2 border-black/[.06] bg-white px-6 pt-10 pb-6 text-center shadow-2xl dark:border-white/[.1] dark:bg-zinc-900"
      >
        <span
          aria-hidden
          className="absolute -top-9 flex h-[72px] w-[72px] items-center justify-center rounded-full border-4 border-white bg-amber-100 text-4xl shadow-lg dark:border-zinc-900 dark:bg-amber-500/20"
        >
          <span className="cute-wiggle">{emoji}</span>
        </span>
        <h2 id="cute-dialog-title" className="text-lg font-bold text-black dark:text-zinc-50">
          {title}
        </h2>
        {children && <div className="text-sm text-zinc-600 dark:text-zinc-400">{children}</div>}
        <div className="mt-2 flex w-full flex-col gap-2">{actions}</div>
      </div>
    </div>
  );
}
