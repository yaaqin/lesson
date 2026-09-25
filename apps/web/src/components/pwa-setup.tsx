"use client";

import { useEffect } from "react";
import { usePwaStore, type BeforeInstallPromptEvent } from "@/store/pwa-store";

// Dipasang sekali di root (providers.tsx): daftarin service worker & nangkep
// beforeinstallprompt sedini mungkin -- event ini cuma nembak sekali per load,
// jadi harus udah didengerin sebelum user sampai ke halaman yang ada tombol
// install-nya.
export function PwaSetup() {
  const setDeferredPrompt = usePwaStore((s) => s.setDeferredPrompt);
  const setInstalled = usePwaStore((s) => s.setInstalled);

  useEffect(() => {
    if ("serviceWorker" in navigator) {
      navigator.serviceWorker.register("/sw.js", { scope: "/", updateViaCache: "none" }).catch(() => {
        // gagal daftar (mis. http non-localhost) -- app tetap jalan, cuma gak installable
      });
    }

    const onBeforeInstallPrompt = (e: Event) => {
      e.preventDefault();
      setDeferredPrompt(e as BeforeInstallPromptEvent);
    };
    const onInstalled = () => {
      setDeferredPrompt(null);
      setInstalled(true);
    };
    window.addEventListener("beforeinstallprompt", onBeforeInstallPrompt);
    window.addEventListener("appinstalled", onInstalled);
    return () => {
      window.removeEventListener("beforeinstallprompt", onBeforeInstallPrompt);
      window.removeEventListener("appinstalled", onInstalled);
    };
  }, [setDeferredPrompt, setInstalled]);

  return null;
}
