"use client";

import Link from "next/link";
import { useRouter } from "next/navigation";
import { useState, type FormEvent } from "react";
import { useLoginMutation } from "@/hooks/use-auth";

export default function LoginPage() {
  const router = useRouter();
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [error, setError] = useState<string | null>(null);
  const loginMutation = useLoginMutation();

  const handleSubmit = async (event: FormEvent) => {
    event.preventDefault();
    setError(null);
    try {
      await loginMutation.mutateAsync({ identifier: email, password });
      router.push("/belajar");
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
            Login SSO Google belum jalan, pakai email/password dulu.
          </p>
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
