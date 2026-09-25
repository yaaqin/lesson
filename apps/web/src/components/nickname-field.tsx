"use client";

import { useEffect, useState } from "react";
import { useNicknameCheckQuery } from "@/hooks/use-profile";
import { NICKNAME_MAX, nicknameFormatError, normalizeNicknameInput } from "@/lib/nickname";

export type NicknameStatus = "invalid" | "checking" | "taken" | "available" | "unchanged";

// Input nickname: validasi format langsung di klien, cek ketersediaan ke
// server (debounce 400ms). Status akhirnya dikabarin ke parent lewat onStatus
// biar tombol simpan bisa dikunci. Validasi final tetap di server.
export function NicknameField({
  value,
  onChange,
  current,
  onStatus,
}: {
  value: string;
  onChange: (value: string) => void;
  current?: string | null;
  onStatus: (status: NicknameStatus) => void;
}) {
  const [debounced, setDebounced] = useState(value);
  useEffect(() => {
    const id = window.setTimeout(() => setDebounced(value), 400);
    return () => window.clearTimeout(id);
  }, [value]);

  const formatError = nicknameFormatError(value);
  const unchanged = !!current && value === current;
  const check = useNicknameCheckQuery(debounced, !formatError && !unchanged && debounced === value);

  let status: NicknameStatus;
  if (formatError) status = "invalid";
  else if (unchanged) status = "unchanged";
  else if (debounced !== value || check.isFetching || !check.data) status = "checking";
  else status = check.data.available ? "available" : "taken";

  useEffect(() => {
    onStatus(status);
  }, [status, onStatus]);

  const message =
    value.length === 0
      ? null
      : status === "invalid"
        ? { text: formatError, tone: "text-red-500" }
        : status === "taken"
          ? { text: "Nickname ini udah dipakai user lain.", tone: "text-red-500" }
          : status === "available"
            ? { text: "Nickname tersedia ✓", tone: "text-green-600 dark:text-green-400" }
            : status === "checking"
              ? { text: "Mengecek…", tone: "text-zinc-500" }
              : null;

  return (
    <label className="flex flex-col gap-1.5 text-sm">
      <span className="font-medium text-zinc-700 dark:text-zinc-300">Nickname</span>
      <div className="flex items-center rounded-lg border border-black/[.08] px-3 focus-within:border-blue-500 dark:border-white/[.145]">
        <span className="text-zinc-400">@</span>
        <input
          type="text"
          value={value}
          maxLength={NICKNAME_MAX}
          autoCapitalize="none"
          autoCorrect="off"
          spellCheck={false}
          autoComplete="off"
          onChange={(e) => onChange(normalizeNicknameInput(e.target.value))}
          placeholder="contoh: budi_pintar"
          className="w-full bg-transparent px-1 py-2 text-black outline-none dark:text-zinc-50"
        />
        <span className="text-xs text-zinc-400">
          {value.length}/{NICKNAME_MAX}
        </span>
      </div>
      {message ? (
        <span className={`text-xs ${message.tone}`}>{message.text}</span>
      ) : (
        <span className="text-xs text-zinc-500">Minimal 6 karakter: huruf a-z, angka 0-9, _ dan titik (.)</span>
      )}
    </label>
  );
}
