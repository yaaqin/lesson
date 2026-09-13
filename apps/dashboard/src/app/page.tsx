import Link from "next/link";

export default function Home() {
  return (
    <div className="flex min-h-screen flex-col items-center justify-center bg-zinc-50 font-sans dark:bg-black">
      <main className="flex flex-col items-center gap-4 text-center">
        <span className="rounded-full bg-black/[.06] px-3 py-1 text-sm font-medium text-zinc-600 dark:bg-white/[.08] dark:text-zinc-400">
          dashboard
        </span>
        <h1 className="text-3xl font-semibold tracking-tight text-black dark:text-zinc-50">
          MathQuest Dashboard
        </h1>
        <p className="max-w-md text-zinc-600 dark:text-zinc-400">
          Satu app, dua ruang: <code className="font-mono text-[0.9em]">/org</code> (dashboard
          organisasi) dan <code className="font-mono text-[0.9em]">/admin</code> (dashboard
          admin platform).
        </p>
        <div className="mt-2 flex items-center gap-3">
          <Link
            href="/login"
            className="rounded-full bg-foreground px-6 py-2.5 text-sm font-medium text-background transition-colors hover:bg-[#383838] dark:hover:bg-[#ccc]"
          >
            Masuk
          </Link>
          <Link
            href="/register-organisasi"
            className="rounded-full border border-black/[.08] px-6 py-2.5 text-sm font-medium text-zinc-600 transition-colors hover:bg-black/[.04] dark:border-white/[.145] dark:text-zinc-400 dark:hover:bg-[#1a1a1a]"
          >
            Daftarkan Organisasi
          </Link>
        </div>
      </main>
    </div>
  );
}
