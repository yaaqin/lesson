"use client";

const KEYPAD_KEYS = ["7", "8", "9", "4", "5", "6", "1", "2", "3", "-", "0", "."];

// Keypad angka on-screen buat soal essay_numeric -- sengaja bukan <input type="number">
// biar konsisten di semua device (gak gantung keyboard native OS) & gampang dikunci
// (disabled) begitu jawaban udah disubmit, kayak tombol opsi pilihan ganda.
export function NumericKeypad({
  value,
  onChange,
  onSubmit,
  disabled,
  feedback,
}: {
  value: string;
  onChange: (value: string) => void;
  onSubmit?: () => void;
  disabled?: boolean;
  feedback?: "correct" | "incorrect" | null;
}) {
  const press = (key: string) => {
    if (disabled) return;
    if (key === "-") {
      onChange(value.startsWith("-") ? value.slice(1) : "-" + value);
      return;
    }
    if (key === "." && value.includes(".")) return;
    onChange(value + key);
  };

  return (
    <div className="flex w-full max-w-xs flex-col gap-3">
      <div
        className={`flex h-16 items-center justify-center rounded-2xl border-2 text-3xl font-semibold transition-colors ${
          feedback === "correct"
            ? "border-green-500 bg-green-50 text-green-700 dark:bg-green-500/10 dark:text-green-400"
            : feedback === "incorrect"
              ? "border-red-500 bg-red-50 text-red-700 dark:bg-red-500/10 dark:text-red-400"
              : "border-black/[.08] bg-white text-black dark:border-white/[.145] dark:bg-zinc-900 dark:text-zinc-50"
        }`}
      >
        {value || <span className="text-zinc-300 dark:text-zinc-700">0</span>}
      </div>

      <div className="grid grid-cols-3 gap-2">
        {KEYPAD_KEYS.map((key) => (
          <button
            key={key}
            type="button"
            disabled={disabled}
            onClick={() => press(key)}
            className="rounded-xl border border-black/[.08] bg-white py-4 text-xl font-semibold text-black transition-colors hover:border-blue-400 disabled:opacity-40 dark:border-white/[.145] dark:bg-zinc-900 dark:text-zinc-50"
          >
            {key}
          </button>
        ))}
      </div>

      <div className={onSubmit ? "grid grid-cols-2 gap-2" : "grid grid-cols-1 gap-2"}>
        <button
          type="button"
          disabled={disabled}
          onClick={() => onChange(value.slice(0, -1))}
          className="rounded-xl border border-black/[.08] py-3 text-sm font-medium text-zinc-600 disabled:opacity-40 dark:border-white/[.145] dark:text-zinc-400"
        >
          ⌫ Hapus
        </button>
        {onSubmit && (
          <button
            type="button"
            disabled={disabled || value === "" || value === "-"}
            onClick={onSubmit}
            className="rounded-xl bg-foreground py-3 text-sm font-semibold text-background disabled:opacity-40"
          >
            Jawab
          </button>
        )}
      </div>
    </div>
  );
}
