import type { Metadata } from "next";
import { notFound } from "next/navigation";
import { ShareInvite } from "@/components/share-invite";
import { buildShareContent, fetchShareProfile, isShareKind, sharePath } from "@/lib/share";
import { siteOrigin } from "@/lib/site-origin";

// Halaman undangan publik (tanpa login) dari link share ranking/streak:
// nampilin kartu share + ajakan gabung. og:image-nya kartu yang sama, jadi
// preview link di WA/X/FB/Telegram langsung keliatan peringkatnya.

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
    <ShareInvite
      headline={`${profile.nickname} nantang kamu adu matematika!`}
      text={content.text}
      imageSrc={`${sharePath(profile.nickname, kind)}/image`}
      imageAlt={`${content.badge} ${profile.nickname}: ${content.big} ${content.caption}`}
    />
  );
}
