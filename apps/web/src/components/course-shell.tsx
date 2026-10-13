"use client";

import Link from "next/link";
import { usePathname, useRouter } from "next/navigation";
import { useEffect, type ReactNode } from "react";
import { useAuthStore } from "@/store/auth-store";
import { useRequireNickname } from "@/hooks/use-profile";
import { BackButton } from "@/components/back-button";
import { clearPostLoginPath, rememberPostLoginPath } from "@/lib/post-login";

// Kerangka halaman Course: wajib login (balik lagi ke halaman ini abis login),
// header, tombol back + judul.
export function CourseShell({
  backHref,
  backLabel,
  title,
  subtitle,
  children,
}: {
  backHref: string;
  backLabel: string;
  title: ReactNode;
  subtitle?: ReactNode;
  children: ReactNode;
}) {
  const router = useRouter();
  const pathname = usePathname();
  const session = useAuthStore((s) => s.session);
  const hasHydrated = useAuthStore((s) => s.hasHydrated);
  useRequireNickname();

  useEffect(() => {
    if (!hasHydrated) return;
    if (session === null) {
      rememberPostLoginPath(pathname);
      router.replace("/login");
    }
  }, [hasHydrated, session, router, pathname]);

  useEffect(() => clearPostLoginPath(), []);

  if (!hasHydrated || !session) {
    return (
      <div className="flex min-h-screen items-center justify-center bg-zinc-50 dark:bg-black">
        <p className="text-zinc-500 dark:text-zinc-500">Memuat…</p>
      </div>
    );
  }

  return (
    <div className="flex min-h-screen flex-col bg-zinc-50 font-sans dark:bg-black">
      <header className="flex items-center justify-between px-6 py-5 sm:px-10">
        <Link href="/coding" className="text-lg font-semibold tracking-tight text-black dark:text-zinc-50">
          MathQuest
        </Link>
        <span className="rounded-full bg-emerald-100 px-3 py-1 text-xs font-semibold text-emerald-700 dark:bg-emerald-500/15 dark:text-emerald-400">
          💻 Course
        </span>
      </header>
      <main className="mx-auto flex w-full max-w-3xl flex-1 flex-col gap-6 px-6 py-8">
        <div className="-ml-2 flex items-start gap-1">
          <BackButton href={backHref} label={backLabel} />
          <div className="flex min-w-0 flex-col gap-1 pt-1">
            <h1 className="text-2xl font-semibold tracking-tight text-black dark:text-zinc-50">{title}</h1>
            {subtitle && <p className="text-sm text-zinc-500 dark:text-zinc-500">{subtitle}</p>}
          </div>
        </div>
        {children}
      </main>
    </div>
  );
}
