"use client";

import { useState, type ReactNode } from "react";
import type { InstallPlatform } from "@/hooks/use-pwa-install";

// Tur "cara install" pakai mockup layar HP -- buat browser yang gak ngasih
// prompt install native (Safari iOS, browser bawaan aplikasi, atau Chrome
// Android yang belum nembak beforeinstallprompt). PWA emang gak "ke-install"
// kayak APK, jadi user perlu dituntun ke menu browser-nya.

type Step = { title: string; body: string; screen: ReactNode };

function stepsFor(platform: InstallPlatform, host: string): Step[] {
  if (platform === "ios") {
    return [
      {
        title: "Tap tombol Bagikan",
        body: "Buka MathQuest di Safari, lalu tap ikon Bagikan (kotak dengan panah ke atas) di bawah layar.",
        screen: <IosSafariScreen host={host} />,
      },
      {
        title: "Pilih “Tambah ke Layar Utama”",
        body: "Geser daftar menu ke bawah sampai ketemu “Tambah ke Layar Utama”, lalu tap.",
        screen: <IosShareSheetScreen />,
      },
      {
        title: "Tap “Tambah”",
        body: "Nama app-nya udah keisi “MathQuest”. Tap “Tambah” di pojok kanan atas.",
        screen: <IosAddScreen />,
      },
      {
        title: "Selesai! Buka dari layar utama",
        body: "Ikon MathQuest sekarang ada di layar utama. Buka dari situ biar tampil penuh kayak aplikasi.",
        screen: <HomeScreen />,
      },
    ];
  }
  if (platform === "in_app") {
    return [
      {
        title: "Buka di browser dulu",
        body: "Kamu lagi buka MathQuest dari dalam aplikasi lain (Instagram, TikTok, dll) yang gak bisa install app. Tap menu ⋯ di pojok kanan atas.",
        screen: <InAppScreen host={host} />,
      },
      {
        title: "Pilih “Buka di browser”",
        body: "Pilih “Buka di Chrome” / “Buka di browser eksternal”, lalu tap lagi tombol Install aplikasi di sana.",
        screen: <InAppMenuScreen host={host} />,
      },
    ];
  }
  return [
    {
      title: "Buka menu browser",
      body: "Di Chrome, tap ikon titik tiga ⋮ di pojok kanan atas.",
      screen: <AndroidBrowserScreen host={host} />,
    },
    {
      title: "Pilih “Instal aplikasi”",
      body: "Tap “Instal aplikasi” atau “Tambahkan ke layar utama” (namanya beda-beda tergantung versi Chrome).",
      screen: <AndroidMenuScreen host={host} />,
    },
    {
      title: "Tap “Instal”",
      body: "Konfirmasi di jendela yang muncul. Tunggu sebentar sampai ikon MathQuest nongol.",
      screen: <AndroidInstallDialogScreen host={host} />,
    },
    {
      title: "Selesai! Buka dari layar utama",
      body: "Ikon MathQuest sekarang ada di layar utama / laci aplikasi. Buka dari situ biar tampil penuh tanpa address bar.",
      screen: <HomeScreen />,
    },
  ];
}

export function InstallTour({ platform, onClose }: { platform: InstallPlatform; onClose: () => void }) {
  const [step, setStep] = useState(0);
  const steps = stepsFor(platform, typeof window === "undefined" ? "" : window.location.host);
  const current = steps[step];
  const isLast = step === steps.length - 1;

  return (
    <div className="fixed inset-0 z-50 flex items-end justify-center bg-black/50 sm:items-center sm:p-4" onClick={onClose}>
      <div
        role="dialog"
        aria-modal="true"
        aria-label="Cara install MathQuest"
        onClick={(e) => e.stopPropagation()}
        className="flex max-h-[95dvh] w-full max-w-md flex-col gap-4 overflow-y-auto rounded-t-3xl bg-white p-5 pb-6 sm:rounded-3xl dark:bg-zinc-900"
      >
        <div className="flex items-center justify-between">
          <span className="text-xs font-semibold tracking-wide text-blue-600 uppercase dark:text-blue-400">
            Install MathQuest · {step + 1}/{steps.length}
          </span>
          <button
            type="button"
            onClick={onClose}
            aria-label="Tutup"
            className="flex h-8 w-8 items-center justify-center rounded-full text-zinc-500 hover:bg-black/[.06] dark:hover:bg-white/[.08]"
          >
            ✕
          </button>
        </div>

        <div className="flex justify-center">
          <PhoneFrame>{current.screen}</PhoneFrame>
        </div>

        <div className="flex flex-col gap-1 text-center">
          <h3 className="text-lg font-semibold text-black dark:text-zinc-50">{current.title}</h3>
          <p className="text-sm leading-relaxed text-zinc-600 dark:text-zinc-400">{current.body}</p>
        </div>

        <div className="flex justify-center gap-1.5">
          {steps.map((_, i) => (
            <span
              key={i}
              className={`h-1.5 rounded-full transition-all ${i === step ? "w-6 bg-blue-600" : "w-1.5 bg-zinc-300 dark:bg-zinc-700"}`}
            />
          ))}
        </div>

        <div className="grid grid-cols-2 gap-2">
          <button
            type="button"
            onClick={() => (step === 0 ? onClose() : setStep((s) => s - 1))}
            className="rounded-full border border-black/[.08] py-3 text-sm font-medium text-zinc-600 dark:border-white/[.145] dark:text-zinc-400"
          >
            {step === 0 ? "Nanti aja" : "← Balik"}
          </button>
          <button
            type="button"
            onClick={() => (isLast ? onClose() : setStep((s) => s + 1))}
            className="rounded-full bg-blue-600 py-3 text-sm font-semibold text-white hover:bg-blue-700"
          >
            {isLast ? "Siap!" : "Lanjut →"}
          </button>
        </div>
      </div>
    </div>
  );
}

// ---------- mockup layar ----------

function PhoneFrame({ children }: { children: ReactNode }) {
  return (
    <div className="relative h-[340px] w-[180px] rounded-[2rem] border-[6px] border-zinc-800 bg-zinc-800 shadow-xl dark:border-zinc-700">
      <div className="absolute top-1.5 left-1/2 z-10 h-3 w-14 -translate-x-1/2 rounded-full bg-zinc-800 dark:bg-zinc-700" />
      <div className="relative h-full w-full overflow-hidden rounded-[1.5rem] bg-zinc-50 text-[9px] text-zinc-800">
        {children}
      </div>
    </div>
  );
}

// Penanda bagian yang harus di-tap: ring biru + titik yang berdenyut.
function Tap({ children, className = "" }: { children: ReactNode; className?: string }) {
  return (
    <span className={`relative inline-flex items-center justify-center rounded-md ring-2 ring-blue-500 ${className}`}>
      {children}
      <span className="absolute -right-1 -bottom-1 flex h-3 w-3">
        <span className="absolute inline-flex h-full w-full animate-ping rounded-full bg-blue-400 opacity-75" />
        <span className="relative inline-flex h-3 w-3 rounded-full bg-blue-600" />
      </span>
    </span>
  );
}

function AppIcon({ size = 32 }: { size?: number }) {
  return (
    // eslint-disable-next-line @next/next/no-img-element -- ikon statis kecil di mockup
    <img src="/icons/icon-192.png" alt="" width={size} height={size} className="rounded-[22%]" />
  );
}

// Konten dummy halaman MathQuest di dalam mockup.
function FakePage({ dim = false }: { dim?: boolean }) {
  return (
    <div className={`flex flex-col gap-2 p-3 ${dim ? "opacity-40" : ""}`}>
      <div className="flex items-center gap-1.5">
        <AppIcon size={16} />
        <span className="font-semibold">MathQuest</span>
      </div>
      <div className="rounded-lg border-2 border-blue-500 bg-white p-2 font-semibold">🗺️ Adventure</div>
      {["SD", "SMP", "SMK / SMA", "Kampus"].map((t) => (
        <div key={t} className="rounded-lg border border-black/10 bg-white p-2">
          {t}
        </div>
      ))}
    </div>
  );
}

function ChromeBar({ host, tapMenu = false }: { host: string; tapMenu?: boolean }) {
  return (
    <div className="flex items-center gap-1.5 border-b border-black/10 bg-white px-2 pt-5 pb-1.5">
      <span className="flex-1 truncate rounded-full bg-zinc-100 px-2 py-1 text-zinc-500">🔒 {host || "mathquest"}</span>
      {tapMenu ? (
        <Tap className="h-5 w-4 font-bold">⋮</Tap>
      ) : (
        <span className="flex h-5 w-4 items-center justify-center font-bold">⋮</span>
      )}
    </div>
  );
}

function AndroidBrowserScreen({ host }: { host: string }) {
  return (
    <>
      <ChromeBar host={host} tapMenu />
      <FakePage />
    </>
  );
}

function AndroidMenuScreen({ host }: { host: string }) {
  return (
    <>
      <ChromeBar host={host} />
      <FakePage dim />
      <div className="absolute top-7 right-1.5 flex w-[120px] flex-col rounded-lg bg-white py-1 shadow-lg ring-1 ring-black/10">
        {["Tab baru", "Riwayat", "Download"].map((item) => (
          <span key={item} className="px-2.5 py-1.5">
            {item}
          </span>
        ))}
        <span className="px-1.5 py-0.5">
          <Tap className="w-full justify-start bg-blue-50 px-1 py-1 font-semibold text-blue-700">📲 Instal aplikasi</Tap>
        </span>
        <span className="px-2.5 py-1.5">Setelan</span>
      </div>
    </>
  );
}

function AndroidInstallDialogScreen({ host }: { host: string }) {
  return (
    <>
      <ChromeBar host={host} />
      <FakePage dim />
      <div className="absolute inset-x-3 top-1/3 flex flex-col gap-2 rounded-xl bg-white p-3 shadow-xl ring-1 ring-black/10">
        <span className="font-semibold">Instal aplikasi?</span>
        <div className="flex items-center gap-2">
          <AppIcon size={24} />
          <div className="flex flex-col">
            <span className="font-semibold">MathQuest</span>
            <span className="text-zinc-500">{host || "mathquest"}</span>
          </div>
        </div>
        <div className="flex justify-end gap-2">
          <span className="px-1.5 py-1 text-blue-700">Batal</span>
          <Tap className="bg-blue-600 px-2 py-1 font-semibold text-white">Instal</Tap>
        </div>
      </div>
    </>
  );
}

function SafariBottomBar({ tapShare = false }: { tapShare?: boolean }) {
  return (
    <div className="absolute inset-x-0 bottom-0 flex items-center justify-around border-t border-black/10 bg-white/95 px-2 pt-1.5 pb-3 text-[11px] text-blue-600">
      <span>‹</span>
      <span>›</span>
      {tapShare ? <Tap className="h-5 w-5">⍐</Tap> : <span>⍐</span>}
      <span>📖</span>
      <span>⧉</span>
    </div>
  );
}

function IosSafariScreen({ host }: { host: string }) {
  return (
    <>
      <div className="px-2 pt-5 pb-1.5">
        <span className="block truncate rounded-lg bg-zinc-100 px-2 py-1 text-center text-zinc-500">
          {host || "mathquest"}
        </span>
      </div>
      <FakePage />
      <SafariBottomBar tapShare />
    </>
  );
}

function IosShareSheetScreen() {
  return (
    <>
      <FakePage dim />
      <div className="absolute inset-x-0 bottom-0 flex flex-col gap-1 rounded-t-2xl bg-zinc-100 p-2 pb-3 shadow-xl">
        <div className="mx-auto mb-1 h-1 w-8 rounded-full bg-zinc-300" />
        <div className="flex items-center gap-1.5 rounded-lg bg-white p-1.5">
          <AppIcon size={18} />
          <span className="font-semibold">MathQuest</span>
        </div>
        <div className="flex flex-col rounded-lg bg-white">
          <span className="border-b border-black/5 px-2 py-1.5">Salin</span>
          <span className="border-b border-black/5 px-2 py-1.5">Tambah ke Daftar Bacaan</span>
          <span className="px-1 py-1">
            <Tap className="w-full justify-between bg-blue-50 px-1.5 py-1 font-semibold text-blue-700">
              Tambah ke Layar Utama <span>⊞</span>
            </Tap>
          </span>
        </div>
      </div>
    </>
  );
}

function IosAddScreen() {
  return (
    <div className="flex h-full flex-col bg-zinc-100">
      <div className="flex items-center justify-between px-2 pt-6 pb-2">
        <span className="text-blue-600">Batal</span>
        <span className="font-semibold">Tambah ke Layar Utama</span>
        <Tap className="px-1 font-semibold text-blue-600">Tambah</Tap>
      </div>
      <div className="mx-2 flex items-center gap-2 rounded-lg bg-white p-2">
        <AppIcon size={28} />
        <div className="flex flex-col">
          <span className="font-semibold">MathQuest</span>
          <span className="text-zinc-400">mathquest</span>
        </div>
      </div>
    </div>
  );
}

function HomeScreen() {
  const dummies = ["bg-green-400", "bg-yellow-400", "bg-pink-400", "bg-sky-400", "bg-orange-400", "bg-violet-400", "bg-red-400"];
  return (
    <div className="grid h-full grid-cols-4 content-start gap-x-2 gap-y-3 bg-gradient-to-b from-indigo-400 to-purple-500 px-3 pt-8">
      {dummies.map((c, i) => (
        <div key={i} className="flex flex-col items-center gap-0.5">
          <span className={`h-7 w-7 rounded-[22%] ${c}`} />
          <span className="h-1 w-6 rounded bg-white/50" />
        </div>
      ))}
      <div className="flex flex-col items-center gap-0.5">
        <Tap className="rounded-[22%]">
          <AppIcon size={28} />
        </Tap>
        <span className="text-[7px] text-white">MathQuest</span>
      </div>
    </div>
  );
}

function InAppBar({ host, tapMenu = false }: { host: string; tapMenu?: boolean }) {
  return (
    <div className="flex items-center gap-1.5 border-b border-black/10 bg-white px-2 pt-5 pb-1.5">
      <span className="font-bold">✕</span>
      <span className="flex-1 truncate text-center text-zinc-500">{host || "mathquest"}</span>
      {tapMenu ? <Tap className="h-5 w-5 font-bold">⋯</Tap> : <span className="font-bold">⋯</span>}
    </div>
  );
}

function InAppScreen({ host }: { host: string }) {
  return (
    <>
      <InAppBar host={host} tapMenu />
      <FakePage />
    </>
  );
}

function InAppMenuScreen({ host }: { host: string }) {
  return (
    <>
      <InAppBar host={host} />
      <FakePage dim />
      <div className="absolute inset-x-0 bottom-0 flex flex-col gap-1 rounded-t-2xl bg-white p-2 pb-4 shadow-xl">
        <span className="px-2 py-1.5">Salin link</span>
        <span className="px-1 py-0.5">
          <Tap className="w-full justify-start bg-blue-50 px-1.5 py-1.5 font-semibold text-blue-700">
            🌐 Buka di browser eksternal
          </Tap>
        </span>
        <span className="px-2 py-1.5">Laporkan</span>
      </div>
    </>
  );
}
