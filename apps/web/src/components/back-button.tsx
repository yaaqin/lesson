"use client";

import Link from "next/link";
import { HugeiconsIcon } from "@hugeicons/react";
import { ArrowLeft02Icon } from "@hugeicons/core-free-icons";
import { useGoBack } from "@/lib/nav-history";

// Tombol back icon-only (Hugeicons). Dipasang sejajar judul/header --
// `label` jadi aria-label + tooltip karena gak ada teks yang keliatan.
// Klik = router.back() kalau ada riwayat in-app; `href` cuma fallback
// (dibuka langsung / abis login) & tetap jadi href buat buka di tab baru.
export function BackButton({ href, label }: { href: string; label: string }) {
  const goBack = useGoBack(href);
  return (
    <Link
      href={href}
      onClick={(e) => {
        if (e.metaKey || e.ctrlKey || e.shiftKey || e.button !== 0) return;
        e.preventDefault();
        goBack();
      }}
      aria-label={label}
      title={label}
      className="flex h-9 w-9 shrink-0 items-center justify-center rounded-full text-zinc-600 transition-colors hover:bg-black/[.06] hover:text-black dark:text-zinc-400 dark:hover:bg-white/[.08] dark:hover:text-zinc-50"
    >
      <HugeiconsIcon icon={ArrowLeft02Icon} size={22} strokeWidth={1.8} />
    </Link>
  );
}
