"use client";

import { HugeiconsIcon } from "@hugeicons/react";
import { ComputerIcon, Moon02Icon, Sun03Icon } from "@hugeicons/core-free-icons";
import { useUpdateThemeMutation } from "@/hooks/use-profile";
import type { ThemePreference } from "@/lib/theme";

const OPTIONS = [
  { value: "light", label: "Terang", icon: Sun03Icon },
  { value: "dark", label: "Gelap", icon: Moon02Icon },
  { value: "system", label: "Ikut sistem", icon: ComputerIcon },
] as const;

// Pilihan tema tampilan (segmented control) di halaman Profil.
export function ThemePicker({ value }: { value: ThemePreference }) {
  const mutation = useUpdateThemeMutation();

  return (
    <div className="flex flex-col gap-2 rounded-2xl border border-black/[.08] bg-white p-4 dark:border-white/[.145] dark:bg-zinc-900">
      <span className="text-sm font-semibold text-black dark:text-zinc-50">Tampilan</span>
      <div role="radiogroup" aria-label="Tema tampilan" className="grid grid-cols-3 gap-1 rounded-full bg-black/[.05] p-1 dark:bg-white/[.06]">
        {OPTIONS.map((o) => {
          const active = value === o.value;
          return (
            <button
              key={o.value}
              type="button"
              role="radio"
              aria-checked={active}
              onClick={() => !active && mutation.mutate(o.value)}
              className={`flex items-center justify-center gap-1.5 rounded-full py-2 text-xs font-medium transition-colors ${
                active
                  ? "bg-white text-black shadow-sm dark:bg-zinc-700 dark:text-zinc-50"
                  : "text-zinc-500 hover:text-zinc-800 dark:text-zinc-400 dark:hover:text-zinc-200"
              }`}
            >
              <HugeiconsIcon icon={o.icon} size={16} strokeWidth={1.8} />
              {o.label}
            </button>
          );
        })}
      </div>
      {mutation.isError && (
        <p className="text-xs text-red-500">Tema kepake di device ini, tapi gagal disimpan ke akun. Coba lagi nanti.</p>
      )}
    </div>
  );
}
