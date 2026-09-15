"use client";

import Link from "next/link";
import { useRouter } from "next/navigation";
import { useEffect } from "react";
import { useAuthStore } from "@/store/auth-store";
import { useLogoutMutation } from "@/hooks/use-auth";
import {
  approveRegistration,
  rejectRegistration,
  useRegistrations,
} from "@/store/org-registrations-store";
import { PLATFORM_ROLE_LABEL } from "@/lib/dummy-accounts";

export default function AdminDashboardPage() {
  const router = useRouter();
  const session = useAuthStore((s) => s.session);
  const hasHydrated = useAuthStore((s) => s.hasHydrated);
  const registrations = useRegistrations();
  const logoutMutation = useLogoutMutation();

  useEffect(() => {
    if (!hasHydrated) return; // tunggu localStorage kebaca dulu, jangan buru-buru redirect
    if (session === null) {
      router.replace("/login");
      return;
    }
    if (session.area !== "admin") {
      router.replace("/org");
    }
  }, [hasHydrated, session, router]);

  if (!hasHydrated || !session || session.area !== "admin") {
    return (
      <div className="flex min-h-screen items-center justify-center bg-zinc-50 dark:bg-black">
        <p className="text-zinc-500 dark:text-zinc-500">Memuat…</p>
      </div>
    );
  }

  const pending = registrations.filter((r) => r.status === "pending");
  const decided = registrations.filter((r) => r.status !== "pending");

  return (
    <div className="flex min-h-screen flex-col bg-zinc-50 font-sans dark:bg-black">
      <header className="flex items-center justify-between px-6 py-5 sm:px-10">
        <div className="flex flex-col">
          <span className="text-lg font-semibold tracking-tight text-black dark:text-zinc-50">
            MathQuest Admin
          </span>
          <span className="text-xs text-zinc-500 dark:text-zinc-500">
            {session.displayName} · {PLATFORM_ROLE_LABEL[session.platformRole]}
          </span>
        </div>
        <button
          type="button"
          onClick={async () => {
            await logoutMutation.mutateAsync();
            router.push("/login");
          }}
          className="rounded-full border border-black/[.08] px-4 py-1.5 text-sm font-medium text-zinc-600 transition-colors hover:bg-black/[.04] dark:border-white/[.145] dark:text-zinc-400 dark:hover:bg-[#1a1a1a]"
        >
          Keluar
        </button>
      </header>

      <main className="mx-auto flex w-full max-w-3xl flex-1 flex-col gap-8 px-6 py-8">
        <div className="flex flex-col gap-1">
          <h1 className="text-2xl font-semibold tracking-tight text-black dark:text-zinc-50">
            Persetujuan Organisasi
          </h1>
          <p className="text-sm text-zinc-500 dark:text-zinc-500">
            Organisasi baru butuh persetujuan sebelum pemiliknya bisa login ke dashboard.
          </p>
        </div>

        <section className="flex flex-col gap-3">
          <h2 className="text-sm font-semibold tracking-wide text-zinc-500 uppercase dark:text-zinc-500">
            Menunggu persetujuan ({pending.length})
          </h2>
          {pending.length === 0 ? (
            <p className="rounded-2xl border border-dashed border-black/[.08] px-5 py-6 text-center text-sm text-zinc-500 dark:border-white/[.145] dark:text-zinc-500">
              Belum ada pendaftaran organisasi baru.
            </p>
          ) : (
            <div className="flex flex-col gap-3">
              {pending.map((reg) => (
                <div
                  key={reg.id}
                  className="flex flex-col gap-3 rounded-2xl border border-black/[.08] bg-white p-5 sm:flex-row sm:items-center sm:justify-between dark:border-white/[.145] dark:bg-zinc-900"
                >
                  <div className="flex flex-col">
                    <span className="font-semibold text-black dark:text-zinc-50">
                      {reg.organizationName}
                    </span>
                    <span className="text-sm text-zinc-500 dark:text-zinc-500">
                      {reg.ownerName} · {reg.ownerEmail}
                    </span>
                  </div>
                  <div className="flex gap-2">
                    <button
                      type="button"
                      onClick={() => approveRegistration(reg.id)}
                      className="rounded-full bg-green-600 px-4 py-1.5 text-sm font-medium text-white transition-colors hover:bg-green-700"
                    >
                      Setujui
                    </button>
                    <button
                      type="button"
                      onClick={() => rejectRegistration(reg.id)}
                      className="rounded-full border border-black/[.08] px-4 py-1.5 text-sm font-medium text-zinc-600 transition-colors hover:bg-black/[.04] dark:border-white/[.145] dark:text-zinc-400 dark:hover:bg-[#1a1a1a]"
                    >
                      Tolak
                    </button>
                  </div>
                </div>
              ))}
            </div>
          )}
        </section>

        {decided.length > 0 && (
          <section className="flex flex-col gap-3">
            <h2 className="text-sm font-semibold tracking-wide text-zinc-500 uppercase dark:text-zinc-500">
              Riwayat keputusan
            </h2>
            <div className="flex flex-col gap-2">
              {decided.map((reg) => (
                <div
                  key={reg.id}
                  className="flex items-center justify-between rounded-xl border border-black/[.08] bg-white px-4 py-3 text-sm dark:border-white/[.145] dark:bg-zinc-900"
                >
                  <span className="text-zinc-700 dark:text-zinc-300">
                    {reg.organizationName}{" "}
                    <span className="text-zinc-400 dark:text-zinc-600">· {reg.ownerEmail}</span>
                  </span>
                  <span
                    className={
                      reg.status === "approved"
                        ? "rounded-full bg-green-100 px-3 py-1 text-xs font-medium text-green-700 dark:bg-green-500/10 dark:text-green-400"
                        : "rounded-full bg-red-100 px-3 py-1 text-xs font-medium text-red-700 dark:bg-red-500/10 dark:text-red-400"
                    }
                  >
                    {reg.status === "approved" ? "Disetujui" : "Ditolak"}
                  </span>
                </div>
              ))}
            </div>
          </section>
        )}

        <Link
          href="/admin/kurikulum"
          className="flex flex-col gap-1 rounded-2xl border border-black/[.08] bg-white p-5 transition-colors hover:border-blue-400 dark:border-white/[.145] dark:bg-zinc-900"
        >
          <span className="font-semibold text-black dark:text-zinc-50">
            Kurikulum & Waktu Challenge →
          </span>
          <span className="text-sm text-zinc-500 dark:text-zinc-500">
            Lihat tier/batch/challenge SD, SMP, SMK/SMA dan atur waktu per soal (challenge biasa)
            atau waktu total (ujian).
          </span>
        </Link>

        <Link
          href="/admin/users"
          className="flex flex-col gap-1 rounded-2xl border border-black/[.08] bg-white p-5 transition-colors hover:border-blue-400 dark:border-white/[.145] dark:bg-zinc-900"
        >
          <span className="font-semibold text-black dark:text-zinc-50">Daftar User →</span>
          <span className="text-sm text-zinc-500 dark:text-zinc-500">
            Lihat murid terdaftar, streak, sisa nyawa, dan reset nyawa manual per user.
          </span>
        </Link>

        <section className="flex flex-col gap-3 rounded-2xl border border-dashed border-black/[.08] p-5 text-sm text-zinc-500 dark:border-white/[.145] dark:text-zinc-500">
          <span className="font-medium text-zinc-600 dark:text-zinc-400">Segera hadir</span>
          <span>Bank soal global, tambah tier baru, dan statistik platform.</span>
        </section>
      </main>
    </div>
  );
}
