import type { Metadata } from "next";
import Link from "next/link";
import { headers } from "next/headers";
import { notFound } from "next/navigation";
import { buildShareContent, fetchShareProfile, isShareKind, sharePath } from "@/lib/share";

// Halaman undangan publik (tanpa login) dari link share ranking/streak:
// nampilin kartu share + ajakan gabung. og:image-nya kartu yang sama, jadi
// preview link di WA/X/FB/Telegram langsung keliatan peringkatnya.

async function siteOrigin() {
  const h = await headers();
  const host = h.get("x-forwarded-host") ?? h.get("host") ?? "localhost:9803";
  const proto = h.get("x-forwarded-proto") ?? (host.startsWith("localhost") ? "http" : "https");
  return `${proto}://${host}`;
}

export async function generateMetadata({ params }: PageProps<"/s/[nickname]/[kind]">): Promise<Metadata> {
  const { nickname, kind } = await params;
  if (!isShareKind(kind)) return {};
  const profile = await fetchShareProfile(nickname);
  if (!profile) return {};

  const content = buildShareContent(profile, kind);
  const origin = await siteOrigin();
  const title = `${profile.nickname} — ${content.badge.replace(/^\S+\s/, "")} ${content.big} | MathQuest`;
  const image = `${origin}${sharePath(profile.nickname, kind)}/image`;
  return {
    title,
    description: content.text,
    openGraph: {
      title,
      description: content.text,
      url: `${origin}${sharePath(profile.nickname, kind)}`,
      siteName: "MathQuest",
      images: [{ url: image, width: 1200, height: 630 }],
    },
    twitter: { card: "summary_large_image", title, description: content.text, images: [image] },
  };
}

export default async function SharePage({ params }: PageProps<"/s/[nickname]/[kind]">) {
  const { nickname, kind } = await params;
  if (!isShareKind(kind)) notFound();
  const profile = await fetchShareProfile(nickname);
  if (!profile) notFound();

  const content = buildShareContent(profile, kind);

  return (
    <div className="flex min-h-screen flex-col items-center justify-center gap-8 bg-zinc-50 px-4 py-10 font-sans dark:bg-black">
      <div className="flex flex-col items-center gap-2 text-center">
        <span className="text-sm font-semibold tracking-wide text-blue-600 uppercase dark:text-blue-400">
          Undangan MathQuest
        </span>
        <h1 className="max-w-xl text-2xl font-semibold tracking-tight text-black sm:text-3xl dark:text-zinc-50">
          {profile.nickname} nantang kamu adu matematika!
        </h1>
        <p className="max-w-md text-sm text-zinc-600 dark:text-zinc-400">{content.text}</p>
      </div>

      {/* eslint-disable-next-line @next/next/no-img-element -- PNG dinamis dari route /image */}
      <img
        src={`${sharePath(profile.nickname, kind)}/image`}
        alt={`${content.badge} ${profile.nickname}: ${content.big} ${content.caption}`}
        width={1200}
        height={630}
        className="h-auto w-full max-w-2xl rounded-2xl border border-black/[.08] shadow-lg dark:border-white/[.145]"
      />

      <div className="flex w-full max-w-sm flex-col gap-2">
        <Link
          href="/login"
          className="rounded-full bg-foreground px-6 py-3 text-center text-sm font-medium text-background transition-colors hover:bg-[#383838] dark:hover:bg-[#ccc]"
        >
          Gabung gratis &amp; mulai latihan
        </Link>
        <p className="text-center text-xs text-zinc-500">
          Latihan matematika jadi game: SD, SMP, SMK/SMA sampai Kampus, lengkap sama streak harian dan ranking.
        </p>
      </div>
    </div>
  );
}
