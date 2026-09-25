"use client";

import { useState } from "react";
import { useQuery } from "@tanstack/react-query";
import { HugeiconsIcon } from "@hugeicons/react";
import {
  Cancel01Icon,
  Copy01Icon,
  Download04Icon,
  Facebook01Icon,
  NewTwitterIcon,
  Share08Icon,
  TelegramIcon,
  Tick02Icon,
  WhatsappIcon,
} from "@hugeicons/core-free-icons";
import { publicApi } from "@/lib/http";
import { buildShareContent, sharePath, type ShareKind, type ShareProfile } from "@/lib/share";

const DEFAULT_BUTTON_CLASS =
  "flex items-center gap-1.5 rounded-full border border-black/[.08] px-3 py-1.5 text-sm font-medium text-zinc-600 transition-colors hover:bg-black/[.04] dark:border-white/[.145] dark:text-zinc-400 dark:hover:bg-[#1a1a1a]";

// Tombol "Bagikan" + sheet pilihan share (sosmed, salin link, download
// gambar). Link yang dibagiin ngarah ke halaman undangan /s/{nickname}/{kind}.
export function ShareButton({
  nickname,
  kind,
  label = "Bagikan",
  className,
}: {
  nickname: string;
  kind: ShareKind;
  label?: string;
  className?: string;
}) {
  const [open, setOpen] = useState(false);
  return (
    <>
      <button type="button" onClick={() => setOpen(true)} className={className ?? DEFAULT_BUTTON_CLASS}>
        <HugeiconsIcon icon={Share08Icon} size={16} strokeWidth={1.8} />
        {label}
      </button>
      {open && <ProfileShareSheet nickname={nickname} kind={kind} onClose={() => setOpen(false)} />}
    </>
  );
}

// PathShareButton: share halaman undangan apa aja yang punya route /image
// (mis. hasil multiplayer /s/{nickname}/multiplayer/{matchId}).
export function PathShareButton({
  path,
  text,
  fileName,
  label = "Bagikan",
  className,
}: {
  path: string;
  text: string;
  fileName: string;
  label?: string;
  className?: string;
}) {
  const [open, setOpen] = useState(false);
  return (
    <>
      <button type="button" onClick={() => setOpen(true)} className={className ?? DEFAULT_BUTTON_CLASS}>
        <HugeiconsIcon icon={Share08Icon} size={16} strokeWidth={1.8} />
        {label}
      </button>
      {open && <ShareSheet path={path} text={text} fileName={fileName} onClose={() => setOpen(false)} />}
    </>
  );
}

function ProfileShareSheet({ nickname, kind, onClose }: { nickname: string; kind: ShareKind; onClose: () => void }) {
  const profileQuery = useQuery({
    queryKey: ["share-profile", nickname],
    queryFn: async () =>
      (await publicApi.get<ShareProfile>(`/public/share/${encodeURIComponent(nickname)}`)).data,
    staleTime: 0,
  });
  const text = profileQuery.data ? buildShareContent(profileQuery.data, kind).text : "Ayo latihan matematika bareng di MathQuest!";
  return (
    <ShareSheet path={sharePath(nickname, kind)} text={text} fileName={`mathquest-${nickname}-${kind}.png`} onClose={onClose} />
  );
}

function ShareSheet({
  path,
  text,
  fileName,
  onClose,
}: {
  path: string;
  text: string;
  fileName: string;
  onClose: () => void;
}) {
  const [copied, setCopied] = useState(false);
  const [busy, setBusy] = useState(false);

  const url = `${window.location.origin}${path}`;
  const imageUrl = `${path}/image`;

  const fetchImageFile = async () => {
    const blob = await (await fetch(imageUrl)).blob();
    return new File([blob], fileName, { type: "image/png" });
  };

  // Share bawaan HP (bisa langsung ke IG/WA story dll) -- kirim gambarnya
  // kalau browser dukung share file, kalau gak cukup teks + link.
  const nativeShare = async () => {
    setBusy(true);
    try {
      const file = await fetchImageFile().catch(() => null);
      if (file && navigator.canShare?.({ files: [file] })) {
        await navigator.share({ files: [file], text: `${text} ${url}` });
      } else {
        await navigator.share({ text, url });
      }
    } catch {
      // user batal share -- abaikan
    } finally {
      setBusy(false);
    }
  };

  const download = async () => {
    setBusy(true);
    try {
      const file = await fetchImageFile();
      const a = document.createElement("a");
      a.href = URL.createObjectURL(file);
      a.download = file.name;
      a.click();
      URL.revokeObjectURL(a.href);
    } finally {
      setBusy(false);
    }
  };

  const copyLink = async () => {
    await navigator.clipboard.writeText(`${text} ${url}`);
    setCopied(true);
    window.setTimeout(() => setCopied(false), 2000);
  };

  const enc = encodeURIComponent;
  const links = [
    { label: "WhatsApp", icon: WhatsappIcon, href: `https://wa.me/?text=${enc(`${text} ${url}`)}` },
    { label: "X", icon: NewTwitterIcon, href: `https://twitter.com/intent/tweet?text=${enc(text)}&url=${enc(url)}` },
    { label: "Facebook", icon: Facebook01Icon, href: `https://www.facebook.com/sharer/sharer.php?u=${enc(url)}` },
    { label: "Telegram", icon: TelegramIcon, href: `https://t.me/share/url?url=${enc(url)}&text=${enc(text)}` },
  ];
  const canNativeShare = typeof navigator !== "undefined" && typeof navigator.share === "function";

  return (
    <div className="fixed inset-0 z-50 flex items-end justify-center bg-black/50 sm:items-center sm:p-4" onClick={onClose}>
      <div
        role="dialog"
        aria-modal="true"
        aria-label="Bagikan"
        onClick={(e) => e.stopPropagation()}
        className="flex max-h-[90vh] w-full max-w-md flex-col gap-4 overflow-y-auto rounded-t-2xl bg-white p-5 sm:rounded-2xl dark:bg-zinc-900"
      >
        <div className="flex items-center justify-between">
          <h3 className="text-base font-semibold text-black dark:text-zinc-50">Bagikan & ajak teman</h3>
          <button
            type="button"
            onClick={onClose}
            aria-label="Tutup"
            className="flex h-8 w-8 items-center justify-center rounded-full text-zinc-500 hover:bg-black/[.06] dark:hover:bg-white/[.08]"
          >
            <HugeiconsIcon icon={Cancel01Icon} size={18} />
          </button>
        </div>

        {/* eslint-disable-next-line @next/next/no-img-element -- PNG dinamis dari route /image */}
        <img
          src={imageUrl}
          alt="Kartu share"
          width={1200}
          height={630}
          className="h-auto w-full rounded-xl border border-black/[.08] bg-zinc-100 dark:border-white/[.145] dark:bg-zinc-800"
        />

        {canNativeShare && (
          <button
            type="button"
            disabled={busy}
            onClick={nativeShare}
            className="flex items-center justify-center gap-2 rounded-full bg-foreground px-6 py-3 text-sm font-medium text-background disabled:opacity-50"
          >
            <HugeiconsIcon icon={Share08Icon} size={18} strokeWidth={1.8} />
            Bagikan gambar
          </button>
        )}

        <div className="grid grid-cols-4 gap-2">
          {links.map((l) => (
            <a
              key={l.label}
              href={l.href}
              target="_blank"
              rel="noopener noreferrer"
              className="flex flex-col items-center gap-1.5 rounded-xl py-3 text-xs font-medium text-zinc-600 hover:bg-black/[.04] dark:text-zinc-400 dark:hover:bg-white/[.06]"
            >
              <HugeiconsIcon icon={l.icon} size={24} strokeWidth={1.6} />
              {l.label}
            </a>
          ))}
        </div>

        <div className="grid grid-cols-2 gap-2">
          <button
            type="button"
            onClick={copyLink}
            className="flex items-center justify-center gap-2 rounded-full border border-black/[.08] py-2.5 text-sm font-medium text-zinc-600 dark:border-white/[.145] dark:text-zinc-400"
          >
            <HugeiconsIcon icon={copied ? Tick02Icon : Copy01Icon} size={16} strokeWidth={1.8} />
            {copied ? "Tersalin" : "Salin link"}
          </button>
          <button
            type="button"
            disabled={busy}
            onClick={download}
            className="flex items-center justify-center gap-2 rounded-full border border-black/[.08] py-2.5 text-sm font-medium text-zinc-600 disabled:opacity-50 dark:border-white/[.145] dark:text-zinc-400"
          >
            <HugeiconsIcon icon={Download04Icon} size={16} strokeWidth={1.8} />
            Simpan gambar
          </button>
        </div>
        <p className="text-center text-xs text-zinc-500">
          Buat Instagram / TikTok: simpan gambarnya, terus upload ke story atau postingan kamu.
        </p>
      </div>
    </div>
  );
}
