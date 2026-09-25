import { buildShareContent, fetchShareProfile, isShareKind } from "@/lib/share";
import { renderShareCard } from "@/lib/share-card";

// Gambar kartu share (PNG 1200x630): dipakai sebagai og:image halaman
// /s/{nickname}/{kind}, dan juga yang di-download / dikirim lewat tombol share.
export const revalidate = 60;

const ACCENT: Record<string, string> = {
  sd: "#38bdf8",
  smp: "#34d399",
  smk: "#fbbf24",
  kampus: "#c084fc",
  adventure: "#60a5fa",
  streak: "#fb923c",
};

export async function GET(_req: Request, ctx: RouteContext<"/s/[nickname]/[kind]/image">) {
  const { nickname, kind } = await ctx.params;
  if (!isShareKind(kind)) return new Response("not found", { status: 404 });

  const profile = await fetchShareProfile(nickname);
  if (!profile) return new Response("not found", { status: 404 });

  return renderShareCard({
    nickname: profile.nickname,
    avatar: profile.avatar,
    content: buildShareContent(profile, kind),
    accent: ACCENT[kind],
  });
}
