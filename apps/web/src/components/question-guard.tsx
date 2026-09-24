"use client";

import type { ReactNode, SyntheticEvent } from "react";

// QuestionGuard: bungkus area soal (pilihan ganda, isian, puzzle) -- matiin
// copy/cut, klik kanan, drag, dan seleksi teks. Murni deterrence level rendah
// (nyusahin copy-paste cepet ke chat AI); user tetap bisa screenshot / ngetik
// ulang manual. Percobaan copy dilaporin lewat onCopyBlocked.
export function QuestionGuard({
  children,
  onCopyBlocked,
}: {
  children: ReactNode;
  onCopyBlocked?: () => void;
}) {
  const blockCopy = (e: SyntheticEvent) => {
    e.preventDefault();
    onCopyBlocked?.();
  };
  const block = (e: SyntheticEvent) => e.preventDefault();

  return (
    <div
      className="flex flex-col gap-8 select-none [-webkit-touch-callout:none]"
      onCopy={blockCopy}
      onCut={blockCopy}
      onContextMenu={block}
      onDragStart={block}
    >
      {children}
    </div>
  );
}
