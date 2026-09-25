import Link from "next/link";
import { HugeiconsIcon } from "@hugeicons/react";
import { ArrowLeft02Icon } from "@hugeicons/core-free-icons";

// Tombol back icon-only (Hugeicons). Dipasang sejajar judul/header --
// `label` jadi aria-label + tooltip karena gak ada teks yang keliatan.
export function BackButton({ href, label }: { href: string; label: string }) {
  return (
    <Link
      href={href}
      aria-label={label}
      title={label}
      className="flex h-9 w-9 shrink-0 items-center justify-center rounded-full text-zinc-600 transition-colors hover:bg-black/[.06] hover:text-black dark:text-zinc-400 dark:hover:bg-white/[.08] dark:hover:text-zinc-50"
    >
      <HugeiconsIcon icon={ArrowLeft02Icon} size={22} strokeWidth={1.8} />
    </Link>
  );
}
