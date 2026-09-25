import type { Metadata } from "next";
import { notFound } from "next/navigation";
import { ShareInvite } from "@/components/share-invite";
import { buildMatchShareContent, fetchMatchShare, findMatchPlayer, matchSharePath } from "@/lib/share";
import { siteOrigin } from "@/lib/site-origin";

// Halaman undangan dari share hasil game multiplayer: "aku menang lawan
// siapa". Data diambil dari server (bukan URL), jadi gak bisa dipalsuin.

async function load(nickname: string, matchId: string) {
  const match = await fetchMatchShare(matchId);
  if (!match) return null;
  const me = findMatchPlayer(match, nickname);
  if (!me) return null;
  return { match, me, content: buildMatchShareContent(match.mode, match.players, me.nickname) };
}

export async function generateMetadata({
  params,
}: PageProps<"/s/[nickname]/multiplayer/[matchId]">): Promise<Metadata> {
  const { nickname, matchId } = await params;
  const data = await load(nickname, matchId);
  if (!data) return {};

  const { me, content } = data;
  const origin = await siteOrigin();
  const path = matchSharePath(me.nickname, matchId);
  const title = `${me.nickname} — ${content.big} ${content.caption} | MathQuest`;
  const image = `${origin}${path}/image`;
  return {
    title,
    description: content.text,
    openGraph: {
      title,
      description: content.text,
      url: `${origin}${path}`,
      siteName: "MathQuest",
      images: [{ url: image, width: 1200, height: 630 }],
    },
    twitter: { card: "summary_large_image", title, description: content.text, images: [image] },
  };
}

export default async function MatchSharePage({ params }: PageProps<"/s/[nickname]/multiplayer/[matchId]">) {
  const { nickname, matchId } = await params;
  const data = await load(nickname, matchId);
  if (!data) notFound();

  const { me, content } = data;
  return (
    <ShareInvite
      headline={`${me.nickname} nantang kamu adu matematika bareng!`}
      text={content.text}
      imageSrc={`${matchSharePath(me.nickname, matchId)}/image`}
      imageAlt={`${content.badge} ${me.nickname}: ${content.big} ${content.caption}`}
    />
  );
}
