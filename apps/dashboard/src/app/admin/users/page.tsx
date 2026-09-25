"use client";

import Link from "next/link";
import { useRouter } from "next/navigation";
import { useEffect, useMemo, useState } from "react";
import {
  createColumnHelper,
  flexRender,
  getCoreRowModel,
  useReactTable,
} from "@tanstack/react-table";
import { useAuthStore } from "@/store/auth-store";
import { useLogoutMutation } from "@/hooks/use-auth";
import { useAdminUsersQuery, type AdminUserListItem } from "@/hooks/use-admin-users";
import { PLATFORM_ROLE_LABEL } from "@/lib/dummy-accounts";
import { BackButton } from "@/components/back-button";

const PAGE_SIZE = 20;

const columnHelper = createColumnHelper<AdminUserListItem>();

const columns = [
  columnHelper.display({
    id: "user",
    header: "Nama & Email",
    cell: ({ row }) => (
      <div className="flex flex-col">
        <span className="font-medium text-black dark:text-zinc-50">{row.original.displayName}</span>
        <span className="text-xs text-zinc-500 dark:text-zinc-500">{row.original.email}</span>
      </div>
    ),
  }),
  columnHelper.display({
    id: "streak",
    header: "Streak",
    cell: ({ row }) => (
      <span className="text-sm text-zinc-700 dark:text-zinc-300">
        🔥 {row.original.currentStreak}{" "}
        <span className="text-xs text-zinc-400 dark:text-zinc-600">
          (terpanjang {row.original.longestStreak})
        </span>
      </span>
    ),
  }),
  columnHelper.accessor("livesRemaining", {
    header: "Nyawa",
    cell: (info) => (
      <span className="text-sm">
        {Array.from({ length: 3 }).map((_, i) => (
          <span key={i} className={i < info.getValue() ? "" : "opacity-20"}>
            ❤️
          </span>
        ))}
      </span>
    ),
  }),
  columnHelper.accessor("createdAt", {
    header: "Terdaftar",
    cell: (info) => (
      <span className="text-xs text-zinc-500 dark:text-zinc-500">
        {new Date(info.getValue()).toLocaleDateString("id-ID", { dateStyle: "medium" })}
      </span>
    ),
  }),
  columnHelper.display({
    id: "actions",
    header: "",
    cell: ({ row }) => (
      <Link
        href={`/admin/users/${row.original.id}`}
        className="text-xs font-medium text-blue-600 dark:text-blue-400"
      >
        Detail →
      </Link>
    ),
  }),
];

export default function AdminUsersPage() {
  const router = useRouter();
  const session = useAuthStore((s) => s.session);
  const hasHydrated = useAuthStore((s) => s.hasHydrated);
  const logoutMutation = useLogoutMutation();

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

  const usersQuery = useAdminUsersQuery({ page, pageSize: PAGE_SIZE, q: debouncedSearch });

  const data = useMemo(() => usersQuery.data?.items ?? [], [usersQuery.data]);
  const total = usersQuery.data?.total ?? 0;
  const pageCount = Math.max(1, Math.ceil(total / PAGE_SIZE));

  const table = useReactTable({
    data,
    columns,
    getCoreRowModel: getCoreRowModel(),
    manualPagination: true,
    pageCount,
  });

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

      <main className="mx-auto flex w-full max-w-4xl flex-1 flex-col gap-6 px-6 py-8">
        <div className="flex flex-col gap-1">
          <div className="-ml-2 flex items-center gap-1">
            <BackButton href="/admin" label="Kembali" />
            <h1 className="text-2xl font-semibold tracking-tight text-black dark:text-zinc-50">Daftar User</h1>
          </div>
          <p className="text-sm text-zinc-500 dark:text-zinc-500">
            Murid terdaftar — klik &ldquo;Detail&rdquo; buat lihat streak, nyawa, dan reset nyawa manual.
          </p>
        </div>

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
              {table.getHeaderGroups().map((headerGroup) => (
                <tr key={headerGroup.id} className="border-b border-black/[.08] dark:border-white/[.145]">
                  {headerGroup.headers.map((header) => (
                    <th
                      key={header.id}
                      className="px-4 py-3 text-xs font-semibold uppercase tracking-wide text-zinc-500 dark:text-zinc-500"
                    >
                      {flexRender(header.column.columnDef.header, header.getContext())}
                    </th>
                  ))}
                </tr>
              ))}
            </thead>
            <tbody>
              {usersQuery.isLoading ? (
                <tr>
                  <td colSpan={columns.length} className="px-4 py-8 text-center text-sm text-zinc-500 dark:text-zinc-500">
                    Memuat…
                  </td>
                </tr>
              ) : data.length === 0 ? (
                <tr>
                  <td colSpan={columns.length} className="px-4 py-8 text-center text-sm text-zinc-500 dark:text-zinc-500">
                    Gak ada user yang cocok.
                  </td>
                </tr>
              ) : (
                table.getRowModel().rows.map((row) => (
                  <tr
                    key={row.id}
                    className="border-b border-black/[.06] last:border-0 dark:border-white/[.08]"
                  >
                    {row.getVisibleCells().map((cell) => (
                      <td key={cell.id} className="px-4 py-3">
                        {flexRender(cell.column.columnDef.cell, cell.getContext())}
                      </td>
                    ))}
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
            Halaman {page} dari {pageCount} · {total} user
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
      </main>
    </div>
  );
}
