import type { QuestionOption } from "@/hooks/use-curriculum";

// Prompt soal coding = teks + blok kode yang dibungkus ``` (lihat
// services/api cmd/seed-curriculum/coding.go). Prompt matematika gak punya
// blok kode, jadi tetep dirender persis kayak sebelumnya.
export function splitPrompt(prompt: string): { text: string; code: string | null } {
  const match = prompt.match(/^([\s\S]*?)\n?```\n([\s\S]*?)\n?```\s*$/);
  if (!match) return { text: prompt, code: null };
  return { text: match[1].trim(), code: match[2] };
}

// `kode` di dalam teks soal dirender sebagai inline code.
function InlineText({ text }: { text: string }) {
  return (
    <>
      {text.split(/(`[^`]+`)/g).map((part, i) =>
        part.startsWith("`") && part.endsWith("`") && part.length > 2 ? (
          <code
            key={i}
            className="rounded bg-black/[.06] px-1.5 py-0.5 font-mono text-[0.9em] dark:bg-white/[.1]"
          >
            {part.slice(1, -1)}
          </code>
        ) : (
          part
        ),
      )}
    </>
  );
}

export function CodeBlock({ code, compact = false }: { code: string; compact?: boolean }) {
  return (
    <pre
      className={`overflow-x-auto rounded-xl bg-zinc-900 text-left font-mono text-zinc-100 dark:bg-zinc-950 dark:ring-1 dark:ring-white/[.1] ${
        compact ? "mt-1 px-3 py-2 text-xs" : "px-4 py-3 text-sm sm:text-base"
      }`}
    >
      <code>{code}</code>
    </pre>
  );
}

// className = kelas heading prompt matematika (beda-beda per tampilan). Prompt
// yang ada kodenya pakai heading lebih kecil + blok kode di bawahnya.
export function QuestionPrompt({ prompt, className }: { prompt: string; className: string }) {
  const { text, code } = splitPrompt(prompt);
  if (code === null) {
    return (
      <h2 className={className}>
        <InlineText text={text} />
      </h2>
    );
  }
  return (
    <div className="flex flex-col gap-3">
      <h2 className="text-lg font-semibold tracking-tight text-black dark:text-zinc-50">
        <InlineText text={text} />
      </h2>
      <CodeBlock code={code} />
    </div>
  );
}

// Ringkasan prompt di daftar hasil: teks + blok kode kecil.
export function QuestionPromptSummary({ prompt }: { prompt: string }) {
  const { text, code } = splitPrompt(prompt);
  return (
    <>
      <InlineText text={text} />
      {code !== null && <CodeBlock code={code} compact />}
    </>
  );
}

export function hasTextOptions(options: QuestionOption[] | undefined): boolean {
  return !!options?.some((o) => o.label !== undefined);
}

// Isi tombol opsi: label teks (monospace, boleh multi-baris kayak output
// console.log) atau angka.
export function OptionText({ option }: { option: QuestionOption }) {
  if (option.label === undefined) return <>{option.value}</>;
  return <span className="font-mono text-base font-medium whitespace-pre-wrap">{option.label}</span>;
}
