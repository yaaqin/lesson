import { ImageResponse } from "next/og";
import { findAvatarCharacter } from "@/lib/avatars";
import { buildShareContent, fetchShareProfile, isShareKind } from "@/lib/share";

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

  const content = buildShareContent(profile, kind);
  const accent = ACCENT[kind];
  const character = findAvatarCharacter(profile.avatar.key);
  const googleUrl = profile.avatar.type === "google" ? profile.avatar.googleUrl : null;

  return new ImageResponse(
    (
      <div
        style={{
          width: "100%",
          height: "100%",
          display: "flex",
          flexDirection: "column",
          justifyContent: "space-between",
          padding: "56px 64px",
          background: `radial-gradient(circle at 85% 20%, ${accent}33, transparent 55%), #09090b`,
          color: "#fafafa",
          fontFamily: "sans-serif",
        }}
      >
        <div style={{ display: "flex", alignItems: "center", justifyContent: "space-between" }}>
          <div style={{ display: "flex", alignItems: "center", gap: 20 }}>
            {googleUrl ? (
              // eslint-disable-next-line @next/next/no-img-element -- dirender Satori, bukan browser
              <img src={googleUrl} width={88} height={88} style={{ borderRadius: 44 }} alt="" />
            ) : (
              <div
                style={{
                  width: 88,
                  height: 88,
                  borderRadius: 44,
                  background: "#27272a",
                  display: "flex",
                  alignItems: "center",
                  justifyContent: "center",
                  fontSize: 52,
                }}
              >
                {character.emoji}
              </div>
            )}
            <div style={{ display: "flex", flexDirection: "column" }}>
              <span style={{ fontSize: 40, fontWeight: 700 }}>{profile.nickname}</span>
              <span style={{ fontSize: 26, color: "#a1a1aa" }}>{content.badge}</span>
            </div>
          </div>
          <span style={{ fontSize: 30, fontWeight: 700, color: accent }}>MathQuest</span>
        </div>

        <div style={{ display: "flex", flexDirection: "column" }}>
          <span style={{ fontSize: 148, fontWeight: 800, lineHeight: 1, color: accent }}>{content.big}</span>
          <span style={{ fontSize: 40, marginTop: 12 }}>{content.caption}</span>
          <span style={{ fontSize: 30, marginTop: 8, color: "#a1a1aa" }}>{content.detail}</span>
        </div>

        <div
          style={{
            display: "flex",
            alignItems: "center",
            justifyContent: "space-between",
            borderTop: "2px solid #27272a",
            paddingTop: 24,
            fontSize: 28,
          }}
        >
          <span style={{ color: "#a1a1aa" }}>Latihan matematika jadi game, SD sampai Kampus</span>
          <span
            style={{
              background: accent,
              color: "#09090b",
              fontWeight: 700,
              padding: "10px 24px",
              borderRadius: 999,
            }}
          >
            Tantang aku! →
          </span>
        </div>
      </div>
    ),
    { width: 1200, height: 630 },
  );
}
