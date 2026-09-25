"use client";

import { useState } from "react";
import { usePwaInstall } from "@/hooks/use-pwa-install";
import { InstallTour } from "@/components/install-tour";

const BANNER_DISMISS_KEY = "mathquest_install_banner_dismissed_at";
const BANNER_SNOOZE_MS = 7 * 24 * 60 * 60 * 1000;

function readDismissedRecently() {
  try {
    const at = Number(localStorage.getItem(BANNER_DISMISS_KEY));
    return Number.isFinite(at) && at > 0 && Date.now() - at < BANNER_SNOOZE_MS;
  } catch {
    return false;
  }
}

// Klik install: kalau browser ngasih prompt native (Chrome Android) langsung
// pakai itu, kalau gak ada -> tampilin tur langkah-langkah.
function useInstallAction() {
  const pwa = usePwaInstall();
  const [showTour, setShowTour] = useState(false);

  const install = async () => {
    const prompted = await pwa.promptInstall();
    if (!prompted) setShowTour(true);
  };

  const tour =
    showTour && pwa.platform ? <InstallTour platform={pwa.platform} onClose={() => setShowTour(false)} /> : null;

  return { pwa, install, tour };
}

// Banner di /belajar -- bisa ditutup, muncul lagi 7 hari kemudian.
export function InstallAppBanner() {
  const { pwa, install, tour } = useInstallAction();
  const [dismissed, setDismissed] = useState(readDismissedRecently);

  if (!pwa.shouldOfferInstall || dismissed) return tour;

  const dismiss = () => {
    setDismissed(true);
    try {
      localStorage.setItem(BANNER_DISMISS_KEY, String(Date.now()));
    } catch {
      // storage diblokir -- banner cukup ketutup buat sesi ini
    }
  };

  return (
    <>
      <div className="flex items-center gap-3 rounded-2xl bg-gradient-to-r from-blue-600 to-indigo-600 p-4 text-white shadow-sm">
        {/* eslint-disable-next-line @next/next/no-img-element -- ikon statis */}
        <img src="/icons/icon-192.png" alt="" width={44} height={44} className="shrink-0 rounded-[22%] ring-2 ring-white/40" />
        <div className="flex min-w-0 flex-1 flex-col">
          <span className="text-sm font-semibold">Install aplikasi MathQuest</span>
          <span className="text-xs text-white/80">
            {pwa.platform === "in_app"
              ? "Buka di Chrome/Safari dulu biar bisa di-install."
              : "Buka langsung dari layar utama, tampil penuh kayak aplikasi."}
          </span>
        </div>
        <button
          type="button"
          onClick={install}
          className="shrink-0 rounded-full bg-white px-4 py-2 text-xs font-semibold text-blue-700 hover:bg-blue-50"
        >
          Install
        </button>
        <button
          type="button"
          onClick={dismiss}
          aria-label="Tutup"
          className="-mr-1 shrink-0 self-start text-white/70 hover:text-white"
        >
          ✕
        </button>
      </div>
      {tour}
    </>
  );
}

// Tombol biasa (mis. di halaman profil) -- selalu muncul selama belum
// dibuka sebagai app, gak ikut snooze banner.
export function InstallAppButton() {
  const { pwa, install, tour } = useInstallAction();
  if (!pwa.shouldOfferInstall) return tour;

  return (
    <>
      <button
        type="button"
        onClick={install}
        className="flex items-center justify-center gap-2 rounded-2xl border border-black/[.08] bg-white px-4 py-3 text-sm font-medium text-black transition-colors hover:border-blue-400 dark:border-white/[.145] dark:bg-zinc-900 dark:text-zinc-50"
      >
        📲 Install aplikasi MathQuest
      </button>
      {tour}
    </>
  );
}
