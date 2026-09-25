"use client";

import Link from "next/link";
import { useRouter, useSearchParams } from "next/navigation";
import { Suspense, useState, type FormEvent } from "react";
import { useLoginMutation } from "@/hooks/use-auth";
import { takePostLoginPath } from "@/lib/post-login";

const API_BASE_URL = process.env.NEXT_PUBLIC_API_BASE_URL ?? "http://localhost:9801/api/v1";

const GOOGLE_ERROR_MESSAGE: Record<string, string> = {
  google_state: "Sesi login Google kedaluwarsa, coba lagi.",
  google_missing_code: "Login Google dibatalkan.",
  google_failed: "Login Google gagal, coba lagi atau pakai email/password.",
};

export default function LoginPage() {
  return (
    <Suspense>
      <LoginForm />
    </Suspense>
  );
}

function LoginForm() {
  const router = useRouter();
  const searchParams = useSearchParams();
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [error, setError] = useState<string | null>(
    () => GOOGLE_ERROR_MESSAGE[searchParams.get("error") ?? ""] ?? null,
  );
  const loginMutation = useLoginMutation();

  const handleSubmit = async (event: FormEvent) => {
    event.preventDefault();
    setError(null);
    try {
      await loginMutation.mutateAsync({ identifier: email, password });
      router.push(takePostLoginPath("/belajar"));
    } catch {
      setError("Email atau password salah.");
    }
  };

  return (
    <div className="flex min-h-screen flex-col items-center justify-center gap-8 bg-zinc-50 px-6 py-12 font-sans dark:bg-black">
      <Link href="/" className="text-lg font-semibold tracking-tight text-black dark:text-zinc-50">
        MathQuest
      </Link>

      <div className="flex w-full max-w-sm flex-col gap-6 rounded-2xl border border-black/[.08] bg-white p-6 dark:border-white/[.145] dark:bg-zinc-900">
        <div className="flex flex-col gap-1">
          <h1 className="text-xl font-semibold text-black dark:text-zinc-50">Masuk</h1>
          <p className="text-sm text-zinc-500 dark:text-zinc-500">
            Pakai akun Google atau email/password.
          </p>
        </div>

        <a
          href={`${API_BASE_URL}/auth/google`}
          className="flex items-center justify-center gap-2 rounded-full border border-black/[.08] px-6 py-2.5 text-sm font-medium text-zinc-700 transition-colors hover:bg-black/[.04] dark:border-white/[.145] dark:text-zinc-300 dark:hover:bg-[#1a1a1a]"
        >
          <svg width="16" height="16" viewBox="0 0 48 48" aria-hidden="true">
            <path
              fill="#FFC107"
              d="M43.6 20.5H42V20H24v8h11.3c-1.6 4.7-6.1 8-11.3 8-6.6 0-12-5.4-12-12s5.4-12 12-12c3.1 0 5.8 1.1 8 3l5.7-5.7C34.6 6 29.6 4 24 4 12.9 4 4 12.9 4 24s8.9 20 20 20 20-8.9 20-20c0-1.3-.1-2.7-.4-3.5z"
            />
            <path
              fill="#FF3D00"
              d="M6.3 14.7l6.6 4.8C14.6 15.9 18.9 13 24 13c3.1 0 5.8 1.1 8 3l5.7-5.7C34.6 6 29.6 4 24 4 16.3 4 9.7 8.3 6.3 14.7z"
            />
            <path
              fill="#4CAF50"
              d="M24 44c5.5 0 10.4-1.9 14.3-5.1l-6.6-5.6C29.7 34.9 27 36 24 36c-5.2 0-9.6-3.3-11.3-8l-6.5 5C9.6 39.6 16.3 44 24 44z"
            />
            <path
              fill="#1976D2"
              d="M43.6 20.5H42V20H24v8h11.3c-.8 2.3-2.2 4.2-4.1 5.6l6.6 5.6C40.9 36.5 44 30.9 44 24c0-1.3-.1-2.7-.4-3.5z"
            />
          </svg>
          Masuk dengan Google
        </a>

        <div className="flex items-center gap-3 text-xs text-zinc-400 dark:text-zinc-600">
          <span className="h-px flex-1 bg-black/[.08] dark:bg-white/[.145]" />
          atau
          <span className="h-px flex-1 bg-black/[.08] dark:bg-white/[.145]" />
        </div>

        <form onSubmit={handleSubmit} className="flex flex-col gap-4">
          <label className="flex flex-col gap-1.5 text-sm">
            <span className="font-medium text-zinc-700 dark:text-zinc-300">Email</span>
            <input
              type="email"
              required
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              className="rounded-lg border border-black/[.08] bg-transparent px-3 py-2 text-black outline-none focus:border-blue-500 dark:border-white/[.145] dark:text-zinc-50"
              placeholder="siswa@mathquest.dev"
            />
          </label>
          <label className="flex flex-col gap-1.5 text-sm">
            <span className="font-medium text-zinc-700 dark:text-zinc-300">Password</span>
            <input
              type="password"
              required
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              className="rounded-lg border border-black/[.08] bg-transparent px-3 py-2 text-black outline-none focus:border-blue-500 dark:border-white/[.145] dark:text-zinc-50"
              placeholder="••••••••"
            />
          </label>

          {error && (
            <p className="rounded-lg bg-red-50 px-3 py-2 text-sm text-red-600 dark:bg-red-500/10 dark:text-red-400">
              {error}
            </p>
          )}

          <button
            type="submit"
            disabled={loginMutation.isPending}
            className="rounded-full bg-foreground px-6 py-2.5 text-sm font-medium text-background transition-colors hover:bg-[#383838] disabled:opacity-50 dark:hover:bg-[#ccc]"
          >
            {loginMutation.isPending ? "Memproses…" : "Masuk"}
          </button>
        </form>
      </div>

      <div className="flex w-full max-w-sm flex-col gap-1 rounded-2xl border border-dashed border-black/[.08] p-4 text-xs text-zinc-500 dark:border-white/[.145] dark:text-zinc-500">
        <span className="font-medium text-zinc-600 dark:text-zinc-400">Akun contoh</span>
        <div className="flex items-center justify-between">
          <span className="font-mono">siswa@mathquest.dev</span>
        </div>
        <span className="font-mono text-zinc-400 dark:text-zinc-600">password: siswa12345</span>
      </div>
    </div>
  );
}
