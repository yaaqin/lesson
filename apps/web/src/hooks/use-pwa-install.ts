import { useCallback, useSyncExternalStore } from "react";
import { usePwaStore } from "@/store/pwa-store";

// android = Android dengan browser biasa, ios = iPhone/iPad, in_app = browser
// bawaan aplikasi (Instagram, TikTok, FB, Line, WebView) yang gak bisa install
// PWA -- user harus buka di Chrome/Safari dulu. desktop = sisanya.
export type InstallPlatform = "android" | "ios" | "in_app" | "desktop";

function detectPlatform(): InstallPlatform {
  const ua = navigator.userAgent;
  if (/FBAN|FBAV|Instagram|Line\/|TikTok|musical_ly|Twitter|Snapchat|; wv\)/i.test(ua)) return "in_app";
  const isIOS = /iPad|iPhone|iPod/.test(ua) || (navigator.platform === "MacIntel" && navigator.maxTouchPoints > 1);
  if (isIOS) return "ios";
  if (/Android/i.test(ua)) return "android";
  return "desktop";
}

function detectStandalone() {
  return (
    window.matchMedia("(display-mode: standalone)").matches ||
    (navigator as Navigator & { standalone?: boolean }).standalone === true
  );
}

const noopSubscribe = () => () => {};

function subscribeDisplayMode(onChange: () => void) {
  const mq = window.matchMedia("(display-mode: standalone)");
  mq.addEventListener("change", onChange);
  return () => mq.removeEventListener("change", onChange);
}

export function usePwaInstall() {
  const deferredPrompt = usePwaStore((s) => s.deferredPrompt);
  const installed = usePwaStore((s) => s.installed);
  const setDeferredPrompt = usePwaStore((s) => s.setDeferredPrompt);

  // null di server (SSR) -- komponen install nunggu nilai klien dulu.
  const platform = useSyncExternalStore<InstallPlatform | null>(noopSubscribe, detectPlatform, () => null);
  const standalone = useSyncExternalStore(subscribeDisplayMode, detectStandalone, () => false);

  // Munculin prompt native. false = gak ada prompt native (tampilin tur manual).
  const promptInstall = useCallback(async () => {
    if (!deferredPrompt) return false;
    await deferredPrompt.prompt();
    await deferredPrompt.userChoice;
    // prompt cuma bisa dipakai sekali
    setDeferredPrompt(null);
    return true;
  }, [deferredPrompt, setDeferredPrompt]);

  const isAppMode = standalone || installed;
  return {
    platform,
    isAppMode,
    canPromptNatively: deferredPrompt !== null,
    // Tombol install ditampilin di HP (termasuk yang butuh tur manual) atau
    // di desktop kalau browsernya ngasih prompt native.
    shouldOfferInstall:
      platform !== null && !isAppMode && (platform !== "desktop" || deferredPrompt !== null),
    promptInstall,
  };
}
