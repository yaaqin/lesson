import type { ReactNode } from "react";
import type { MaterialBlock } from "@/hooks/use-course";

// Teks materi: **tebal** dan `kode inline`. Sengaja parser kecil sendiri --
// konten dari seed (bukan input user), gak perlu markdown lengkap.
export function RichText({ text }: { text: string }) {
  // `` x `` = kode inline yang isinya ada backtick (mis. template literal).
  const parts = text.split(/(\*\*[^*]+\*\*|`` .+? ``|`[^`]+`)/g);
  return (
    <>
      {parts.map((part, i): ReactNode => {
        if (part.startsWith("**") && part.endsWith("**") && part.length > 4) {
          return (
            <strong key={i} className="font-semibold text-black dark:text-zinc-50">
              {part.slice(2, -2)}
            </strong>
          );
        }
        const isDouble = part.startsWith("`` ") && part.endsWith(" ``") && part.length > 6;
        if (isDouble || (part.startsWith("`") && part.endsWith("`") && part.length > 2)) {
          return (
            <code
              key={i}
              className="rounded bg-black/[.06] px-1.5 py-0.5 font-mono text-[0.88em] text-black dark:bg-white/[.1] dark:text-zinc-100"
            >
              {isDouble ? part.slice(3, -3) : part.slice(1, -1)}
            </code>
          );
        }
        return part;
      })}
    </>
  );
}

export function CourseCode({ code, output, title }: { code: string; output?: string; title?: string }) {
  return (
    <div className="overflow-hidden rounded-xl bg-zinc-900 dark:bg-zinc-950 dark:ring-1 dark:ring-white/[.1]">
      {title && (
        <div className="border-b border-white/[.08] px-4 py-2 text-xs font-medium text-zinc-400">{title}</div>
      )}
      <pre className="overflow-x-auto px-4 py-3 font-mono text-sm text-zinc-100">
        <code>{code}</code>
      </pre>
      {output && (
        <div className="border-t border-white/[.08] px-4 py-3">
          <div className="mb-1 text-[10px] font-semibold tracking-wide text-zinc-500 uppercase">Output</div>
          <pre className="overflow-x-auto font-mono text-sm text-emerald-400">{output}</pre>
        </div>
      )}
    </div>
  );
}

function Block({ block }: { block: MaterialBlock }) {
  switch (block.type) {
    case "heading":
      return (
        <h2 className="mt-4 text-xl font-semibold tracking-tight text-black first:mt-0 dark:text-zinc-50">
          {block.text}
        </h2>
      );
    case "paragraph":
      return (
        <p className="leading-relaxed text-zinc-700 dark:text-zinc-300">
          <RichText text={block.text} />
        </p>
      );
    case "code":
      return <CourseCode code={block.code} output={block.output} title={block.title} />;
    case "example":
      return (
        <div className="flex flex-col gap-3 rounded-2xl border border-emerald-500/30 bg-emerald-50/60 p-4 dark:bg-emerald-500/[.06]">
          <div className="flex flex-col gap-1">
            <span className="text-xs font-semibold tracking-wide text-emerald-700 uppercase dark:text-emerald-400">
              💡 {block.title ?? "Contoh"}
            </span>
            {block.text && (
              <p className="text-sm text-zinc-700 dark:text-zinc-300">
                <RichText text={block.text} />
              </p>
            )}
          </div>
          <CourseCode code={block.code} output={block.output} />
        </div>
      );
    case "tip":
      return (
        <div className="rounded-2xl border border-amber-400/40 bg-amber-50 px-4 py-3 text-sm leading-relaxed text-amber-900 dark:bg-amber-500/[.08] dark:text-amber-200">
          <span className="mr-1">⚠️</span>
          <RichText text={block.text} />
        </div>
      );
    case "list":
      return (
        <ul className="flex list-disc flex-col gap-1.5 pl-5 text-zinc-700 dark:text-zinc-300">
          {block.items.map((item, i) => (
            <li key={i}>
              <RichText text={item} />
            </li>
          ))}
        </ul>
      );
    case "table":
      return (
        <div className="overflow-x-auto rounded-xl border border-black/[.08] dark:border-white/[.145]">
          <table className="w-full text-left text-sm">
            <thead className="bg-black/[.03] dark:bg-white/[.04]">
              <tr>
                {block.headers.map((h) => (
                  <th key={h} className="px-3 py-2 font-semibold text-black dark:text-zinc-50">
                    {h}
                  </th>
                ))}
              </tr>
            </thead>
            <tbody>
              {block.rows.map((row, i) => (
                <tr key={i} className="border-t border-black/[.06] dark:border-white/[.08]">
                  {row.map((cell, j) => (
                    <td key={j} className="px-3 py-2 align-top text-zinc-700 dark:text-zinc-300">
                      <RichText text={cell} />
                    </td>
                  ))}
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      );
  }
}

export function CourseMaterial({ blocks }: { blocks: MaterialBlock[] }) {
  return (
    <div className="flex flex-col gap-4">
      {blocks.map((block, i) => (
        <Block key={i} block={block} />
      ))}
    </div>
  );
}
