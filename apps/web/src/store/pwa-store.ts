import { create } from "zustand";

// Event beforeinstallprompt (Chrome/Edge/Samsung Internet di Android & desktop)
// -- belum ada di lib.dom TypeScript.
export type BeforeInstallPromptEvent = Event & {
  prompt: () => Promise<void>;
  userChoice: Promise<{ outcome: "accepted" | "dismissed"; platform: string }>;
};

type PwaState = {
  // Prompt install native yang di-"tahan" (preventDefault) biar bisa dipicu
  // dari tombol kita sendiri. null = browser gak/belum ngasih.
  deferredPrompt: BeforeInstallPromptEvent | null;
  installed: boolean;
  setDeferredPrompt: (e: BeforeInstallPromptEvent | null) => void;
  setInstalled: (value: boolean) => void;
};

export const usePwaStore = create<PwaState>()((set) => ({
  deferredPrompt: null,
  installed: false,
  setDeferredPrompt: (deferredPrompt) => set({ deferredPrompt }),
  setInstalled: (installed) => set({ installed }),
}));
