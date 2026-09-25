"use client";

import { useParams, useRouter } from "next/navigation";
import { useEffect, useState } from "react";
import { useAuthStore } from "@/store/auth-store";
import { useLogoutMutation } from "@/hooks/use-auth";
import {
  useAdminUserDetailQuery,
  useResetUserLivesMutation,
  useSetUserPremiumMutation,
  useSetUserRoomQuotaMutation,
  type AdminUserDetail,
} from "@/hooks/use-admin-users";
import { ConfirmDialog } from "@/components/confirm-dialog";
import { PLATFORM_ROLE_LABEL } from "@/lib/dummy-accounts";
import { BackButton } from "@/components/back-button";

export default function AdminUserDetailPage() {
  const router = useRouter();
  const params = useParams<{ id: string }>();
  const session = useAuthStore((s) => s.session);
  const hasHydrated = useAuthStore((s) => s.hasHydrated);
  const logoutMutation = useLogoutMutation();

  const userDetailQuery = useAdminUserDetailQuery(params.id);
  const resetLivesMutation = useResetUserLivesMutation(params.id);
  const premiumMutation = useSetUserPremiumMutation(params.id);

  const [showConfirm, setShowConfirm] = useState(false);
  const [justReset, setJustReset] = useState(false);

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

  const user = userDetailQuery.data;

  const confirmReset = () => {
    resetLivesMutation.mutate(undefined, {
      onSuccess: () => {
        setShowConfirm(false);
        setJustReset(true);
        window.setTimeout(() => setJustReset(false), 2000);
      },
    });
  };

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

      <main className="mx-auto flex w-full max-w-2xl flex-1 flex-col gap-6 px-6 py-8">
        <div className="-ml-2 flex items-center gap-1">
          <BackButton href="/admin/users" label="Daftar User" />
          <h1 className="text-2xl font-semibold tracking-tight text-black dark:text-zinc-50">Detail User</h1>
        </div>

        {userDetailQuery.isLoading && (
          <p className="text-sm text-zinc-500 dark:text-zinc-500">Memuat detail user…</p>
        )}

        {userDetailQuery.isError && (
          <p className="rounded-2xl border border-dashed border-black/[.08] px-5 py-6 text-center text-sm text-red-500 dark:border-white/[.145]">
            User gak ketemu atau gagal dimuat.
          </p>
        )}

        {user && (
          <>
            <div className="flex flex-col gap-1 rounded-2xl border border-black/[.08] bg-white p-5 dark:border-white/[.145] dark:bg-zinc-900">
              <span className="text-xl font-semibold text-black dark:text-zinc-50">
                {user.displayName}
              </span>
              <span className="text-sm text-zinc-500 dark:text-zinc-500">{user.email}</span>
              <span className="mt-2 w-fit rounded-full bg-zinc-100 px-3 py-1 text-xs font-medium text-zinc-600 dark:bg-zinc-800 dark:text-zinc-400">
                {user.role}
              </span>
              <span className="mt-1 text-xs text-zinc-400 dark:text-zinc-600">
                Terdaftar sejak{" "}
                {new Date(user.createdAt).toLocaleDateString("id-ID", { dateStyle: "long" })}
              </span>
            </div>

            <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
              <div className="flex flex-col gap-2 rounded-2xl border border-black/[.08] bg-white p-5 dark:border-white/[.145] dark:bg-zinc-900">
                <span className="text-sm font-semibold text-zinc-500 dark:text-zinc-500">Streak</span>
                <span className="text-3xl font-bold text-black dark:text-zinc-50">
                  🔥 {user.currentStreak}
                </span>
                <span className="text-xs text-zinc-500 dark:text-zinc-500">
                  Terpanjang: {user.longestStreak} hari
                </span>
                <span className="text-xs text-zinc-400 dark:text-zinc-600">
                  Terakhir aktif:{" "}
                  {user.lastActiveDate
                    ? new Date(user.lastActiveDate).toLocaleDateString("id-ID", { dateStyle: "medium" })
                    : "belum pernah"}
                </span>
              </div>

              <div className="flex flex-col gap-2 rounded-2xl border border-black/[.08] bg-white p-5 dark:border-white/[.145] dark:bg-zinc-900">
                <span className="text-sm font-semibold text-zinc-500 dark:text-zinc-500">Nyawa</span>
                <span className="text-3xl">
                  {Array.from({ length: 3 }).map((_, i) => (
                    <span key={i} className={i < user.livesRemaining ? "" : "opacity-20"}>
                      ❤️
                    </span>
                  ))}
                </span>
                <span className="text-xs text-zinc-500 dark:text-zinc-500">
                  {user.livesRemaining}/3 nyawa tersisa
                </span>
                <span className="text-xs text-zinc-400 dark:text-zinc-600">
                  Reset harian terakhir:{" "}
                  {new Date(user.livesLastResetAt).toLocaleString("id-ID", { dateStyle: "medium", timeStyle: "short" })}
                </span>
                <button
                  type="button"
                  onClick={() => setShowConfirm(true)}
                  disabled={user.livesRemaining >= 3}
                  className="mt-1 w-fit rounded-full bg-foreground px-4 py-1.5 text-xs font-medium text-background disabled:opacity-40"
                >
                  {justReset ? "Berhasil ✓" : "Reset Nyawa"}
                </button>
              </div>
            </div>

            <div className="flex items-center justify-between gap-4 rounded-2xl border border-black/[.08] bg-white p-5 dark:border-white/[.145] dark:bg-zinc-900">
              <div className="flex flex-col gap-1">
                <span className="text-sm font-semibold text-zinc-500 dark:text-zinc-500">
                  Premium {user.isPremium && "👑"}
                </span>
                <span className="text-xs text-zinc-500 dark:text-zinc-500">
                  {user.isPremium
                    ? "Aktif — tag spesial, user ini bisa bikin room multiplayer tanpa batas."
                    : "Belum premium — bikin room makai kuota di bawah."}
                </span>
                {premiumMutation.isError && (
                  <span className="text-xs text-red-500">Gagal ngubah status premium, coba lagi.</span>
                )}
              </div>
              <button
                type="button"
                onClick={() => premiumMutation.mutate(!user.isPremium)}
                disabled={premiumMutation.isPending}
                className={`shrink-0 rounded-full px-4 py-1.5 text-xs font-medium disabled:opacity-40 ${
                  user.isPremium
                    ? "border border-black/[.08] text-zinc-600 dark:border-white/[.145] dark:text-zinc-400"
                    : "bg-amber-500 text-white"
                }`}
              >
                {premiumMutation.isPending ? "Menyimpan…" : user.isPremium ? "Cabut premium" : "Jadikan premium"}
              </button>
            </div>

            <RoomQuotaCard key={user.roomQuota} user={user} />

            <div className="flex flex-col gap-2 rounded-2xl border border-black/[.08] bg-white p-5 dark:border-white/[.145] dark:bg-zinc-900">
              <span className="text-sm font-semibold text-zinc-500 dark:text-zinc-500">
                Riwayat Attempt
              </span>
              <span className="text-sm text-zinc-700 dark:text-zinc-300">
                {user.totalAttempts} total attempt · {user.passedAttempts} lulus
              </span>
            </div>
          </>
        )}
      </main>

      {showConfirm && user && (
        <ConfirmDialog
          title="Reset nyawa?"
          description={`Nyawa ${user.displayName} bakal langsung dibalikin ke penuh (3/3), gak perlu nunggu reset harian. Yakin lanjut?`}
          confirmLabel="Ya, reset"
          isPending={resetLivesMutation.isPending}
          onConfirm={confirmReset}
          onCancel={() => setShowConfirm(false)}
        />
      )}
    </div>
  );
}

// RoomQuotaCard: sisa kesempatan bikin room multiplayer -- admin bisa set
// langsung (gak ngaruh kalau user-nya premium).
function RoomQuotaCard({ user }: { user: AdminUserDetail }) {
  const [value, setValue] = useState(String(user.roomQuota));
  const mutation = useSetUserRoomQuotaMutation(user.id);
  const parsed = Number(value);
  const valid = value !== "" && Number.isInteger(parsed) && parsed >= 0 && parsed <= 1000;

  return (
    <div className="flex flex-col gap-3 rounded-2xl border border-black/[.08] bg-white p-5 dark:border-white/[.145] dark:bg-zinc-900">
      <div className="flex flex-col gap-1">
        <span className="text-sm font-semibold text-zinc-500 dark:text-zinc-500">Kuota Bikin Room</span>
        <span className="text-3xl font-bold text-black tabular-nums dark:text-zinc-50">
          {user.isPremium ? "∞" : `${user.roomQuota}×`}
        </span>
        <span className="text-xs text-zinc-500 dark:text-zinc-500">
          {user.isPremium
            ? `User premium gak kepotong kuota (sisa kuota biasanya: ${user.roomQuota}×).`
            : "Berkurang 1 tiap user bikin room multiplayer."}
        </span>
      </div>
      <form
        onSubmit={(e) => {
          e.preventDefault();
          if (valid) mutation.mutate(parsed);
        }}
        className="flex items-center gap-2"
      >
        <input
          type="number"
          min={0}
          max={1000}
          value={value}
          onChange={(e) => setValue(e.target.value)}
          aria-label="Kuota bikin room"
          className="w-24 rounded-xl border border-black/[.08] bg-zinc-50 px-3 py-1.5 text-sm text-black outline-none focus:border-blue-400 dark:border-white/[.145] dark:bg-black dark:text-zinc-50"
        />
        <button
          type="submit"
          disabled={!valid || parsed === user.roomQuota || mutation.isPending}
          className="rounded-full bg-foreground px-4 py-1.5 text-xs font-medium text-background disabled:opacity-40"
        >
          {mutation.isPending ? "Menyimpan…" : "Set kuota"}
        </button>
        {mutation.isError && <span className="text-xs text-red-500">Gagal, coba lagi.</span>}
      </form>
    </div>
  );
}
