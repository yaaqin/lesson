import { useCallback, useEffect, useRef, useSyncExternalStore } from "react";

function subscribeFullscreen(onChange: () => void) {
  document.addEventListener("fullscreenchange", onChange);
  return () => document.removeEventListener("fullscreenchange", onChange);
}

const noopSubscribe = () => () => {};

// useFullscreenGuard: challenge dikerjain dalam mode fullscreen (Fullscreen
// API) -- nyusahin split-screen sama tab/window lain. Keluar fullscreen selagi
// `active` dilaporin lewat onExit (halaman yang mutusin mau pause timer + nyatet
// event). Browser yang gak dukung (mis. Safari iPhone) -> `supported` false dan
// halaman gak maksa fullscreen; tab switch tetap ketangkep useVisibilityTracker.
export function useFullscreenGuard({ active, onExit }: { active: boolean; onExit: () => void }) {
  const supported = useSyncExternalStore(
    noopSubscribe,
    () => document.fullscreenEnabled && typeof document.documentElement.requestFullscreen === "function",
    () => false,
  );
  const isFullscreen = useSyncExternalStore(
    subscribeFullscreen,
    () => document.fullscreenElement !== null,
    () => false,
  );

  const activeRef = useRef(active);
  const onExitRef = useRef(onExit);
  useEffect(() => {
    activeRef.current = active;
    onExitRef.current = onExit;
  }, [active, onExit]);

  useEffect(() => {
    const onChange = () => {
      if (document.fullscreenElement === null && activeRef.current) onExitRef.current();
    };
    document.addEventListener("fullscreenchange", onChange);
    return () => document.removeEventListener("fullscreenchange", onChange);
  }, []);

  // Navigasi keluar halaman challenge (SPA) gak otomatis keluar fullscreen.
  useEffect(() => {
    return () => {
      if (document.fullscreenElement) document.exitFullscreen().catch(() => {});
    };
  }, []);

  // enter: harus dipanggil langsung dari handler klik (browser nolak
  // requestFullscreen tanpa user gesture).
  const enter = useCallback(async () => {
    if (!supported || document.fullscreenElement) return;
    try {
      await document.documentElement.requestFullscreen({ navigationUI: "hide" });
    } catch {
      // ditolak browser -- biarin, overlay "kembali ke fullscreen" bakal nongol
    }
  }, [supported]);

  const exit = useCallback(() => {
    if (document.fullscreenElement) document.exitFullscreen().catch(() => {});
  }, []);

  return { supported, isFullscreen, enter, exit };
}
