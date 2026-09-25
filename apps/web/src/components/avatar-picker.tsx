"use client";

import { AVATAR_CHARACTERS } from "@/lib/avatars";

export type AvatarChoice = { type: "character"; key: string } | { type: "google" };

// Grid pilihan avatar: foto Google (kalau akunnya punya) + karakter kartun.
export function AvatarPicker({
  value,
  googleUrl,
  onChange,
}: {
  value: AvatarChoice;
  googleUrl: string | null;
  onChange: (choice: AvatarChoice) => void;
}) {
  const ring = (active: boolean) =>
    active ? "ring-2 ring-blue-600 ring-offset-2 ring-offset-white dark:ring-blue-500 dark:ring-offset-zinc-900" : "";

  return (
    <div className="grid grid-cols-5 gap-3 sm:grid-cols-6">
      {googleUrl && (
        <button
          type="button"
          onClick={() => onChange({ type: "google" })}
          title="Foto Google"
          className={`flex aspect-square items-center justify-center overflow-hidden rounded-full ${ring(value.type === "google")}`}
        >
          {/* eslint-disable-next-line @next/next/no-img-element -- foto Google eksternal */}
          <img src={googleUrl} alt="Foto Google" referrerPolicy="no-referrer" className="h-full w-full object-cover" />
        </button>
      )}
      {AVATAR_CHARACTERS.map((c) => (
        <button
          key={c.key}
          type="button"
          onClick={() => onChange({ type: "character", key: c.key })}
          title={c.label}
          className={`flex aspect-square items-center justify-center rounded-full text-2xl transition-transform hover:scale-105 sm:text-3xl ${c.bg} ${ring(
            value.type === "character" && value.key === c.key,
          )}`}
        >
          {c.emoji}
        </button>
      ))}
    </div>
  );
}
