"use client";

import Link from "next/link";
import { useRouter } from "next/navigation";
import { useEffect, useState } from "react";
import { useAuthStore } from "@/store/auth-store";
import { useRequireNickname } from "@/hooks/use-profile";
import {
  ROOM_QUOTA_REQUEST_MAX_LENGTH,
  errorCode,
  useRoomQuotaQuery,
  useRoomQuotaRequestMutation,
  type RoomQuotaStatus,
} from "@/hooks/use-multiplayer";
import { BackButton } from "@/components/back-button";
import { PremiumBadge } from "@/components/premium-badge";
import { MODE_INFO, normalizeRoomCode } from "@/lib/multiplayer";

export default function MultiplayerHomePage() {
  const router = useRouter();
  const session = useAuthStore((s) => s.session);
  const hasHydrated = useAuthStore((s) => s.hasHydrated);
  useRequireNickname();
  const [code, setCode] = useState("");

  useEffect(() => {
    if (!hasHydrated) return;
    if (session === null) router.replace("/login");
  }, [hasHydrated, session, router]);

  if (!hasHydrated || !session) {
    return (
      <div className="flex min-h-screen items-center justify-center bg-zinc-50 dark:bg-black">
        <p className="text-zinc-500 dark:text-zinc-500">Memuat…</p>
      </div>
    );
  }

  const canJoin = code.length === 6;

  return (
    <div className="flex min-h-screen flex-col bg-zinc-50 font-sans dark:bg-black">
      <header className="flex items-center justify-between px-4 py-5 sm:px-10">
        <BackButton href="/belajar" label="Kembali ke jenjang" />
        <span className="text-lg font-semibold tracking-tight text-black dark:text-zinc-50">Multiplayer</span>
        <span className="w-9" />
      </header>

      <main className="mx-auto flex w-full max-w-md flex-1 flex-col gap-6 px-4 pb-10">
        <form
          onSubmit={(e) => {
            e.preventDefault();
            if (canJoin) router.push(`/multiplayer/${code}`);
          }}
          className="flex flex-col gap-3 rounded-2xl border border-black/[.08] bg-white p-5 dark:border-white/[.145] dark:bg-zinc-900"
        >
          <label htmlFor="room-code" className="font-semibold text-black dark:text-zinc-50">
            Gabung room
          </label>
          <input
            id="room-code"
            value={code}
            onChange={(e) => setCode(normalizeRoomCode(e.target.value))}
            placeholder="KODE ROOM"
            autoComplete="off"
            autoCapitalize="characters"
            inputMode="text"
            className="rounded-xl border-2 border-black/[.08] bg-zinc-50 px-4 py-3 text-center font-mono text-2xl font-semibold tracking-[0.3em] text-black uppercase outline-none focus:border-violet-500 dark:border-white/[.145] dark:bg-black dark:text-zinc-50"
          />
          <button
            type="submit"
            disabled={!canJoin}
            className="rounded-full bg-violet-600 px-6 py-3 text-sm font-semibold text-white transition-colors hover:bg-violet-700 disabled:opacity-40"
          >
            Gabung
          </button>
        </form>

        <CreateRoomCard />

        <div className="flex flex-col gap-3">
          <span className="text-sm font-semibold text-zinc-500 dark:text-zinc-500">Mode permainan</span>
          {Object.values(MODE_INFO).map((mode) => (
            <div
              key={mode.label}
              className="flex gap-3 rounded-2xl border border-black/[.08] bg-white p-4 dark:border-white/[.145] dark:bg-zinc-900"
            >
              <span className="text-2xl">{mode.icon}</span>
              <span className="flex flex-col gap-0.5">
                <span className="font-semibold text-black dark:text-zinc-50">{mode.label}</span>
                <span className="text-sm text-zinc-500 dark:text-zinc-400">{mode.desc}</span>
              </span>
            </div>
          ))}
          <p className="text-xs text-zinc-400 dark:text-zinc-600">
            Multiplayer gak motong nyawa harian. Maksimal 10 pemain per room.
          </p>
        </div>
      </main>
    </div>
  );
}

const REQUEST_ERROR_TEXT: Record<string, string> = {
  request_pending: "Permintaan kamu sebelumnya masih ditinjau admin.",
  invalid_message: "Tulis alasan kamu dulu (maks. 500 karakter).",
  already_premium: "Kamu udah premium, bisa bikin room tanpa batas.",
};

// CreateRoomCard: premium = bikin room tanpa batas; user biasa lihat sisa
// kesempatan + bisa minta tambahan ke admin.
function CreateRoomCard() {
  const quotaQuery = useRoomQuotaQuery();
  const quota = quotaQuery.data;

  return (
    <div className="flex flex-col gap-3 rounded-2xl border border-black/[.08] bg-white p-5 dark:border-white/[.145] dark:bg-zinc-900">
      <div className="flex items-center justify-between gap-2">
        <span className="font-semibold text-black dark:text-zinc-50">Bikin room</span>
        {quota?.isPremium && <PremiumBadge size="xs" />}
      </div>

      {quotaQuery.isLoading && <p className="text-sm text-zinc-500">Memuat…</p>}
      {quotaQuery.isError && <p className="text-sm text-red-500">Gagal memuat sisa kesempatan.</p>}

      {quota && (
        <>
          {quota.isPremium ? (
            <p className="text-sm text-zinc-500 dark:text-zinc-400">
              Kamu premium: bikin room sebanyak yang kamu mau. Racik soalnya, bagiin kodenya, tantang temen-temenmu.
            </p>
          ) : (
            <div className="flex items-center justify-between gap-3 rounded-xl bg-zinc-50 px-4 py-3 dark:bg-black">
              <span className="text-sm text-zinc-600 dark:text-zinc-400">Sisa kesempatan bikin room</span>
              <span
                className={`text-2xl font-bold tabular-nums ${
                  quota.roomQuota > 0 ? "text-violet-600 dark:text-violet-400" : "text-zinc-400 dark:text-zinc-600"
                }`}
              >
                {quota.roomQuota}×
              </span>
            </div>
          )}

          {quota.isPremium || quota.roomQuota > 0 ? (
            <Link
              href="/multiplayer/buat"
              className="rounded-full bg-foreground px-6 py-3 text-center text-sm font-semibold text-background"
            >
              Bikin room baru
            </Link>
          ) : (
            <p className="text-sm text-zinc-500 dark:text-zinc-400">
              Kesempatanmu udah habis. Kamu tetap bisa gabung ke room orang lain pakai kodenya, atau minta tambahan
              kesempatan ke admin.
            </p>
          )}

          {!quota.isPremium && <QuotaRequest quota={quota} />}
        </>
      )}
    </div>
  );
}

function QuotaRequest({ quota }: { quota: RoomQuotaStatus }) {
  const latest = quota.latestRequest;
  const [open, setOpen] = useState(quota.roomQuota === 0);
  const [message, setMessage] = useState("");
  const mutation = useRoomQuotaRequestMutation();

  if (latest?.status === "pending") {
    return (
      <div className="flex flex-col gap-1 rounded-xl border border-amber-300 bg-amber-50 px-4 py-3 dark:border-amber-500/30 dark:bg-amber-500/10">
        <span className="text-sm font-semibold text-amber-800 dark:text-amber-300">⏳ Permintaanmu lagi ditinjau admin</span>
        <span className="text-xs break-words text-amber-700/80 dark:text-amber-300/70">“{latest.message}”</span>
        <span className="text-xs text-amber-700/60 dark:text-amber-300/50">
          Dikirim {new Date(latest.createdAt).toLocaleString("id-ID", { dateStyle: "medium", timeStyle: "short" })}
        </span>
      </div>
    );
  }

  const submit = () => {
    const text = message.trim();
    if (!text) return;
    mutation.mutate(text, { onSuccess: () => setMessage("") });
  };
  const error = mutation.isError ? (REQUEST_ERROR_TEXT[errorCode(mutation.error) ?? ""] ?? "Gagal ngirim, coba lagi.") : null;

  return (
    <div className="flex flex-col gap-2">
      {latest?.status === "approved" && (
        <p className="text-xs text-green-600 dark:text-green-400">
          ✅ Permintaan terakhirmu disetujui: +{latest.grantedQuota ?? 0} kesempatan.
        </p>
      )}
      {latest?.status === "rejected" && (
        <p className="text-xs text-red-500">❌ Permintaan terakhirmu ditolak admin. Kamu bisa kirim lagi.</p>
      )}

      {!open ? (
        <button
          type="button"
          onClick={() => setOpen(true)}
          className="text-left text-sm font-medium text-violet-600 dark:text-violet-400"
        >
          Minta tambahan kesempatan ke admin →
        </button>
      ) : (
        <form
          onSubmit={(e) => {
            e.preventDefault();
            submit();
          }}
          className="flex flex-col gap-2"
        >
          <label htmlFor="quota-request" className="text-sm font-medium text-zinc-700 dark:text-zinc-300">
            Minta tambahan kesempatan
          </label>
          <textarea
            id="quota-request"
            value={message}
            onChange={(e) => setMessage(e.target.value.slice(0, ROOM_QUOTA_REQUEST_MAX_LENGTH))}
            rows={3}
            placeholder="Ceritain buat apa, mis. mau bikin kuis bareng teman sekelas tiap minggu."
            className="resize-none rounded-xl border-2 border-black/[.08] bg-zinc-50 px-3 py-2 text-sm text-black outline-none focus:border-violet-500 dark:border-white/[.145] dark:bg-black dark:text-zinc-50"
          />
          <div className="flex items-center justify-between gap-2">
            <span className="text-xs text-zinc-400 tabular-nums">
              {message.length}/{ROOM_QUOTA_REQUEST_MAX_LENGTH}
            </span>
            <button
              type="submit"
              disabled={!message.trim() || mutation.isPending}
              className="rounded-full bg-violet-600 px-5 py-2 text-sm font-semibold text-white transition-colors hover:bg-violet-700 disabled:opacity-40"
            >
              {mutation.isPending ? "Mengirim…" : "Kirim ke admin"}
            </button>
          </div>
          {error && <p className="text-xs text-red-500">{error}</p>}
        </form>
      )}
    </div>
  );
}
