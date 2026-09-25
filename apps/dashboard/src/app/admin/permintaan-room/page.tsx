"use client";

import Link from "next/link";
import { useRouter } from "next/navigation";
import { useEffect, useState } from "react";
import { useAuthStore } from "@/store/auth-store";
import {
  MAX_ROOM_QUOTA_GRANT,
  useApproveRoomQuotaRequestMutation,
  useRejectRoomQuotaRequestMutation,
  useRoomQuotaRequestsQuery,
  type AdminRoomQuotaRequest,
  type RoomQuotaRequestStatus,
} from "@/hooks/use-admin-room-quota";
import { BackButton } from "@/components/back-button";

const PAGE_SIZE = 20;

const TABS: { value: RoomQuotaRequestStatus | ""; label: string }[] = [
  { value: "pending", label: "Menunggu" },
  { value: "approved", label: "Disetujui" },
  { value: "rejected", label: "Ditolak" },
  { value: "", label: "Semua" },
];

export default function AdminRoomQuotaRequestsPage() {
  const router = useRouter();
  const session = useAuthStore((s) => s.session);
  const hasHydrated = useAuthStore((s) => s.hasHydrated);
  const [status, setStatus] = useState<RoomQuotaRequestStatus | "">("pending");
  const [page, setPage] = useState(1);
  const query = useRoomQuotaRequestsQuery({ status, page, pageSize: PAGE_SIZE });

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

  const data = query.data;
  const totalPages = data ? Math.max(1, Math.ceil(data.total / PAGE_SIZE)) : 1;

  return (
    <div className="flex min-h-screen flex-col bg-zinc-50 font-sans dark:bg-black">
      <main className="mx-auto flex w-full max-w-3xl flex-1 flex-col gap-6 px-6 py-8">
        <div className="flex flex-col gap-1">
          <div className="-ml-2 flex items-center gap-1">
            <BackButton href="/admin" label="Dashboard Admin" />
            <h1 className="text-2xl font-semibold tracking-tight text-black dark:text-zinc-50">Permintaan Bikin Room</h1>
          </div>
          <p className="text-sm text-zinc-500 dark:text-zinc-500">
            User minta tambahan kesempatan bikin room multiplayer. Setujui = tambah sejumlah kesempatan ke kuotanya.
          </p>
        </div>

        <div className="flex flex-wrap gap-2">
          {TABS.map((t) => (
            <button
              key={t.value}
              type="button"
              onClick={() => {
                setStatus(t.value);
                setPage(1);
              }}
              className={`rounded-full border px-4 py-1.5 text-sm font-medium transition-colors ${
                status === t.value
                  ? "border-blue-500 bg-blue-50 text-blue-700 dark:bg-blue-500/10 dark:text-blue-300"
                  : "border-black/[.08] text-zinc-600 hover:bg-black/[.04] dark:border-white/[.145] dark:text-zinc-400 dark:hover:bg-[#1a1a1a]"
              }`}
            >
              {t.label}
              {t.value === "pending" && data && data.pendingCount > 0 && ` (${data.pendingCount})`}
            </button>
          ))}
        </div>

        {query.isLoading && <p className="text-sm text-zinc-500 dark:text-zinc-500">Memuat permintaan…</p>}
        {query.isError && <p className="text-sm text-red-500">Gagal memuat permintaan.</p>}
        {data && data.items.length === 0 && (
          <p className="rounded-2xl border border-dashed border-black/[.08] px-5 py-6 text-center text-sm text-zinc-500 dark:border-white/[.145] dark:text-zinc-500">
            {status === "pending" ? "Gak ada permintaan yang nunggu." : "Belum ada permintaan."}
          </p>
        )}

        <div className="flex flex-col gap-3">
          {data?.items.map((req) => <RequestCard key={req.id} req={req} />)}
        </div>

        {data && totalPages > 1 && (
          <div className="flex items-center justify-center gap-3 text-sm">
            <button
              type="button"
              disabled={page <= 1}
              onClick={() => setPage((p) => p - 1)}
              className="rounded-full border border-black/[.08] px-3 py-1 disabled:opacity-40 dark:border-white/[.145] dark:text-zinc-300"
            >
              ←
            </button>
            <span className="text-zinc-500">
              {page} / {totalPages}
            </span>
            <button
              type="button"
              disabled={page >= totalPages}
              onClick={() => setPage((p) => p + 1)}
              className="rounded-full border border-black/[.08] px-3 py-1 disabled:opacity-40 dark:border-white/[.145] dark:text-zinc-300"
            >
              →
            </button>
          </div>
        )}
      </main>
    </div>
  );
}

const DECIDE_ERROR_TEXT: Record<string, string> = {
  already_decided: "Permintaan ini udah diproses admin lain.",
  invalid_amount: `Jumlah kesempatan harus 1–${MAX_ROOM_QUOTA_GRANT}.`,
};

function RequestCard({ req }: { req: AdminRoomQuotaRequest }) {
  const [grant, setGrant] = useState("3");
  const approve = useApproveRoomQuotaRequestMutation();
  const reject = useRejectRoomQuotaRequestMutation();
  const parsed = Number(grant);
  const validGrant = Number.isInteger(parsed) && parsed >= 1 && parsed <= MAX_ROOM_QUOTA_GRANT;
  const busy = approve.isPending || reject.isPending;
  const failed = approve.error ?? reject.error;
  const errCode = (failed as { response?: { data?: { error?: string } } } | null)?.response?.data?.error;
  const fmt = (d: string) => new Date(d).toLocaleString("id-ID", { dateStyle: "medium", timeStyle: "short" });

  return (
    <div className="flex flex-col gap-3 rounded-2xl border border-black/[.08] bg-white p-5 dark:border-white/[.145] dark:bg-zinc-900">
      <div className="flex flex-wrap items-start justify-between gap-2">
        <div className="flex flex-col">
          <Link
            href={`/admin/users/${req.user.id}`}
            className="font-semibold text-black hover:text-blue-600 dark:text-zinc-50 dark:hover:text-blue-400"
          >
            {req.user.displayName}
            {req.user.nickname && <span className="font-normal text-zinc-500"> · @{req.user.nickname}</span>}
          </Link>
          <span className="text-xs text-zinc-500 dark:text-zinc-500">{req.user.email}</span>
        </div>
        <div className="flex flex-col items-end gap-1">
          <StatusPill req={req} />
          <span className="text-xs text-zinc-400 dark:text-zinc-600">
            Sisa kuota: {req.user.isPremium ? "∞ (premium)" : `${req.user.roomQuota}×`}
          </span>
        </div>
      </div>

      <p className="rounded-xl bg-zinc-50 px-4 py-3 text-sm whitespace-pre-wrap break-words text-zinc-700 dark:bg-black dark:text-zinc-300">
        {req.message}
      </p>

      <span className="text-xs text-zinc-400 dark:text-zinc-600">
        Dikirim {fmt(req.createdAt)}
        {req.decidedAt && ` · diproses ${fmt(req.decidedAt)}`}
      </span>

      {req.status === "pending" && (
        <div className="flex flex-wrap items-center gap-2">
          <label className="flex items-center gap-2 text-sm text-zinc-600 dark:text-zinc-400">
            Kasih
            <input
              type="number"
              min={1}
              max={MAX_ROOM_QUOTA_GRANT}
              value={grant}
              onChange={(e) => setGrant(e.target.value)}
              className="w-20 rounded-xl border border-black/[.08] bg-zinc-50 px-3 py-1.5 text-sm text-black outline-none focus:border-blue-400 dark:border-white/[.145] dark:bg-black dark:text-zinc-50"
            />
            kesempatan
          </label>
          <button
            type="button"
            disabled={!validGrant || busy}
            onClick={() => approve.mutate({ id: req.id, grant: parsed })}
            className="rounded-full bg-green-600 px-4 py-1.5 text-sm font-medium text-white transition-colors hover:bg-green-700 disabled:opacity-40"
          >
            {approve.isPending ? "Menyimpan…" : validGrant ? `Setujui +${parsed}` : "Setujui"}
          </button>
          <button
            type="button"
            disabled={busy}
            onClick={() => reject.mutate(req.id)}
            className="rounded-full border border-black/[.08] px-4 py-1.5 text-sm font-medium text-zinc-600 transition-colors hover:bg-black/[.04] disabled:opacity-40 dark:border-white/[.145] dark:text-zinc-400 dark:hover:bg-[#1a1a1a]"
          >
            {reject.isPending ? "Menyimpan…" : "Tolak"}
          </button>
        </div>
      )}
      {failed && <p className="text-xs text-red-500">{DECIDE_ERROR_TEXT[errCode ?? ""] ?? "Gagal memproses, coba lagi."}</p>}
    </div>
  );
}

function StatusPill({ req }: { req: AdminRoomQuotaRequest }) {
  if (req.status === "approved") {
    return (
      <span className="rounded-full bg-green-100 px-3 py-1 text-xs font-medium text-green-700 dark:bg-green-500/10 dark:text-green-400">
        Disetujui +{req.grantedQuota ?? 0}
      </span>
    );
  }
  if (req.status === "rejected") {
    return (
      <span className="rounded-full bg-red-100 px-3 py-1 text-xs font-medium text-red-700 dark:bg-red-500/10 dark:text-red-400">
        Ditolak
      </span>
    );
  }
  return (
    <span className="rounded-full bg-amber-100 px-3 py-1 text-xs font-medium text-amber-700 dark:bg-amber-500/10 dark:text-amber-400">
      Menunggu
    </span>
  );
}
