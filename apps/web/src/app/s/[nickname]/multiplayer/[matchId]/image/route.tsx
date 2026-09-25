import { buildMatchShareContent, fetchMatchShare, findMatchPlayer } from "@/lib/share";
import { renderShareCard } from "@/lib/share-card";

// Kartu share hasil multiplayer (og:image + yang di-download/dikirim).
export const revalidate = 300;

export async function GET(_req: Request, ctx: RouteContext<"/s/[nickname]/multiplayer/[matchId]/image">) {
  const { nickname, matchId } = await ctx.params;
  const match = await fetchMatchShare(matchId);
  const me = match && findMatchPlayer(match, nickname);
  if (!match || !me) return new Response("not found", { status: 404 });

  return renderShareCard({
    nickname: me.nickname,
    avatar: me.avatar,
    content: buildMatchShareContent(match.mode, match.players, me.nickname),
    accent: me.rank === 1 ? "#facc15" : "#a78bfa",
  });
}
