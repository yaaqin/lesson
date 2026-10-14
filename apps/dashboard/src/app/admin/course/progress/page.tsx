"use client";

import Link from "next/link";
import { useRouter } from "next/navigation";
import { useEffect, useState } from "react";
import { useAuthStore } from "@/store/auth-store";
import { useCourseProgressQuery, type CourseModuleFunnel } from "@/hooks/use-admin-course";
import { BackButton } from "@/components/back-button";
import { PositionText, ProgressBar, formatDateTime } from "@/components/course-progress";

const PAGE_SIZE = 20;

export default function AdminCourseProgressPage() {
  const router = useRouter();
  const session = useAuthStore((s) => s.session);
  const hasHydrated = useAuthStore((s) => s.hasHydrated);

  const [page, setPage] = useState(1);
  const [search, setSearch] = useState("");
  const [debouncedSearch, setDebouncedSearch] = useState("");

  useEffect(() => {
    const timeout = window.setTimeout(() => {
      setDebouncedSearch(search);
      setPage(1);
    }, 300);
    return () => window.clearTimeout(timeout);
  }, [search]);

  const progressQuery = useCourseProgressQuery({ page, pageSize: PAGE_SIZE, q: debouncedSearch });

  useEffect(() => {
    if (!hasHydrated) return;
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

  const data = progressQuery.data;
  const users = data?.users.items ?? [];
  const total = data?.users.total ?? 0;
  const pageCount = Math.max(1, Math.ceil(total / PAGE_SIZE));

  return (
    <div className="flex min-h-screen flex-col bg-zinc-50 font-sans dark:bg-black">
      <main className="mx-auto flex w-full max-w-4xl flex-1 flex-col gap-6 px-6 py-8">
        <div className="flex flex-col gap-1">
          <div className="-ml-2 flex items-center gap-1">
            <BackButton href="/admin" label="Dashboard Admin" />
            <h1 className="text-2xl font-semibold tracking-tight text-black dark:text-zinc-50">Progress Course</h1>
          </div>
          <p className="text-sm text-zinc-500 dark:text-zinc-500">
            Pantau murid Course Coding Dasar: udah lulus level berapa, lagi di mana, dan kapan terakhir aktif.
          </p>
        </div>

        {progressQuery.isError && <p className="text-sm text-red-500">Gagal memuat progress course.</p>}

        {data && (
          <>
            <div className="grid grid-cols-2 gap-4">
              <Stat label="Murid yang udah mulai" value={data.startedUsers} />
              <Stat label="Selesai semua module" value={data.completedUsers} />
            </div>

            <section className="flex flex-col gap-3">
              <h2 className="text-sm font-semibold tracking-wide text-zinc-500 uppercase dark:text-zinc-500">
                Lulus per level
              </h2>
              {data.modules
                .filter((m) => !m.comingSoon)
                .map((m) => (
                  <ModuleFunnel key={m.code} module={m} started={data.startedUsers} />
                ))}
            </section>
          </>
        )}

        <section className="flex flex-col gap-3">
          <h2 className="text-sm font-semibold tracking-wide text-zinc-500 uppercase dark:text-zinc-500">Murid</h2>
          <input
            type="text"
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            placeholder="Cari nama atau email…"
            className="w-full max-w-xs rounded-lg border border-black/[.08] bg-white px-3 py-2 text-sm text-black outline-none focus:border-blue-500 dark:border-white/[.145] dark:bg-zinc-900 dark:text-zinc-50"
          />

          <div className="overflow-x-auto rounded-2xl border border-black/[.08] bg-white dark:border-white/[.145] dark:bg-zinc-900">
            <table className="w-full text-left text-sm">
              <thead>
                <tr className="border-b border-black/[.08] dark:border-white/[.145]">
                  {["Nama & Email", "Level lulus", "Lagi di", "Attempt", "Aktif terakhir", ""].map((h) => (
                    <th
                      key={h}
                      className="px-4 py-3 text-xs font-semibold tracking-wide text-zinc-500 uppercase dark:text-zinc-500"
                    >
                      {h}
                    </th>
                  ))}
                </tr>
              </thead>
              <tbody>
                {progressQuery.isLoading ? (
                  <tr>
                    <td colSpan={6} className="px-4 py-8 text-center text-sm text-zinc-500 dark:text-zinc-500">
                      Memuat…
                    </td>
                  </tr>
                ) : users.length === 0 ? (
                  <tr>
                    <td colSpan={6} className="px-4 py-8 text-center text-sm text-zinc-500 dark:text-zinc-500">
                      {debouncedSearch ? "Gak ada murid yang cocok." : "Belum ada murid yang mulai course."}
                    </td>
                  </tr>
                ) : (
                  users.map((u) => (
                    <tr key={u.userId} className="border-b border-black/[.06] last:border-0 dark:border-white/[.08]">
                      <td className="px-4 py-3">
                        <div className="flex flex-col">
                          <span className="font-medium text-black dark:text-zinc-50">{u.displayName}</span>
                          <span className="text-xs text-zinc-500 dark:text-zinc-500">{u.email}</span>
                        </div>
                      </td>
                      <td className="px-4 py-3">
                        <div className="flex w-28 flex-col gap-1">
                          <span className="text-xs text-zinc-600 tabular-nums dark:text-zinc-400">
                            {u.levelsPassed}/{u.totalLevels} · {u.modulesCompleted}/{u.totalModules} module
                          </span>
                          <ProgressBar value={u.levelsPassed} max={u.totalLevels} />
                        </div>
                      </td>
                      <td className="px-4 py-3">
                        <PositionText position={u} />
                      </td>
                      <td className="px-4 py-3 text-sm text-zinc-700 tabular-nums dark:text-zinc-300">{u.attempts}</td>
                      <td className="px-4 py-3 text-xs text-zinc-500 dark:text-zinc-500">
                        {formatDateTime(u.lastActivityAt)}
                      </td>
                      <td className="px-4 py-3">
                        <Link
                          href={`/admin/users/${u.userId}`}
                          className="text-xs font-medium text-blue-600 dark:text-blue-400"
                        >
                          Detail →
                        </Link>
                      </td>
                    </tr>
                  ))
                )}
              </tbody>
            </table>
          </div>

          <div className="flex items-center justify-between">
            <button
              type="button"
              disabled={page <= 1}
              onClick={() => setPage((p) => Math.max(1, p - 1))}
              className="rounded-full border border-black/[.08] px-4 py-2 text-sm font-medium text-zinc-600 disabled:opacity-40 dark:border-white/[.145] dark:text-zinc-400"
            >
              ← Sebelumnya
            </button>
            <span className="text-xs text-zinc-500 dark:text-zinc-500">
              Halaman {page} dari {pageCount} · {total} murid
            </span>
            <button
              type="button"
              disabled={page >= pageCount}
              onClick={() => setPage((p) => Math.min(pageCount, p + 1))}
              className="rounded-full border border-black/[.08] px-4 py-2 text-sm font-medium text-zinc-600 disabled:opacity-40 dark:border-white/[.145] dark:text-zinc-400"
            >
              Berikutnya →
            </button>
          </div>
        </section>
      </main>
    </div>
  );
}

function Stat({ label, value }: { label: string; value: number }) {
  return (
    <div className="flex flex-col gap-1 rounded-2xl border border-black/[.08] bg-white p-5 dark:border-white/[.145] dark:bg-zinc-900">
      <span className="text-sm font-semibold text-zinc-500 dark:text-zinc-500">{label}</span>
      <span className="text-3xl font-bold text-black tabular-nums dark:text-zinc-50">{value}</span>
    </div>
  );
}

// ModuleFunnel: berapa murid (dari yang udah mulai) yang lulus tiap level &
// final test -- level yang angkanya anjlok = tempat murid pada nyangkut.
function ModuleFunnel({ module: m, started }: { module: CourseModuleFunnel; started: number }) {
  const rows = [
    ...m.levels.map((l, i) => ({ key: l.id, label: `${i + 1}. ${l.title}`, n: l.passedUsers })),
    { key: "exam", label: "📝 Final test", n: m.examPassedUsers },
  ];
  return (
    <div className="flex flex-col gap-3 rounded-2xl border border-black/[.08] bg-white p-5 dark:border-white/[.145] dark:bg-zinc-900">
      <span className="font-medium text-black dark:text-zinc-50">{m.title}</span>
      <ul className="flex flex-col gap-2.5">
        {rows.map((r) => (
          <li key={r.key} className="flex flex-col gap-1">
            <div className="flex items-center justify-between gap-3 text-sm">
              <span className="truncate text-zinc-700 dark:text-zinc-300">{r.label}</span>
              <span className="shrink-0 text-xs text-zinc-500 tabular-nums dark:text-zinc-500">
                {r.n} murid{started > 0 ? ` · ${Math.round((r.n / started) * 100)}%` : ""}
              </span>
            </div>
            <ProgressBar value={r.n} max={started} />
          </li>
        ))}
      </ul>
    </div>
  );
}
