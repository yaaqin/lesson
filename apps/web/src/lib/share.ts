import type { UserAvatar } from "@/hooks/use-curriculum";

// Share ranking/streak ke sosmed. Link share = /s/{nickname}/{kind}: halaman
// undangan publik (tanpa login) + gambar kartu di /s/{nickname}/{kind}/image
// (dipakai juga sebagai og:image). Datanya diambil dari server lewat
// nickname, bukan dari URL, jadi rank gak bisa dipalsuin.
// Kontrak: services/api internal/curriculumsvc/share.go.

export type ShareKind = "sd" | "smp" | "smk" | "kampus" | "adventure" | "streak";

export const SHARE_KINDS: ShareKind[] = ["sd", "smp", "smk", "kampus", "adventure", "streak"];

type ShareRank = { rank: number; total: number };

export type ShareProfile = {
  nickname: string;
  avatar: UserAvatar;
  currentStreak: number;
  longestStreak: number;
  tiers: { code: string; name: string; points: number; rank: ShareRank | null }[];
  adventure: {
    checkpointsCleared: number;
    totalCheckpoints: number;
    completed: boolean;
    timePercent: number;
    rank: ShareRank;
  } | null;
};

// Isi kartu: `big` = angka utama, `caption` = keterangan di bawahnya, `text`
// = kalimat ajakan buat caption sosmed.
export type ShareContent = {
  badge: string;
  big: string;
  caption: string;
  detail: string;
  text: string;
};

const API_BASE_URL = process.env.NEXT_PUBLIC_API_BASE_URL ?? "http://localhost:9801/api/v1";

export function isShareKind(kind: string): kind is ShareKind {
  return (SHARE_KINDS as string[]).includes(kind);
}

export function sharePath(nickname: string, kind: ShareKind) {
  return `/s/${encodeURIComponent(nickname)}/${kind}`;
}

// Server-side (page & image route). null = nickname gak ketemu.
export async function fetchShareProfile(nickname: string): Promise<ShareProfile | null> {
  const res = await fetch(`${API_BASE_URL}/public/share/${encodeURIComponent(nickname)}`, {
    next: { revalidate: 60 },
  });
  if (res.status === 404) return null;
  if (!res.ok) throw new Error(`share profile: ${res.status}`);
  return res.json();
}

export function buildShareContent(profile: ShareProfile, kind: ShareKind): ShareContent {
  if (kind === "streak") {
    const days = profile.currentStreak;
    return {
      badge: "🔥 Streak harian",
      big: `${days} hari`,
      caption: days > 0 ? "belajar matematika tanpa bolong" : "siap mulai streak baru",
      detail: `Rekor terpanjang: ${profile.longestStreak} hari`,
      text:
        days > 0
          ? `🔥 Aku udah ${days} hari berturut-turut latihan matematika di MathQuest! Kuat nyusul streak-ku?`
          : `Yuk bareng latihan matematika tiap hari di MathQuest! 🔥`,
    };
  }

  if (kind === "adventure") {
    const adv = profile.adventure;
    if (!adv) {
      return {
        badge: "🗺️ Adventure",
        big: "4.000 soal",
        caption: "dari SD sampai Kampus",
        detail: "40 checkpoint, jatah salah 5x tiap checkpoint",
        text: "Aku lagi mulai Adventure 4.000 soal matematika di MathQuest. Ikutan yuk!",
      };
    }
    const progress = adv.completed ? "Tamat 🏆" : `Checkpoint ${adv.checkpointsCleared}/${adv.totalCheckpoints}`;
    return {
      badge: "🗺️ Ranking Adventure",
      big: `#${adv.rank.rank}`,
      caption: `dari ${adv.rank.total.toLocaleString("id-ID")} petualang`,
      detail: progress,
      text: `🏆 Aku peringkat #${adv.rank.rank} di Adventure MathQuest (${progress})! Berani nyusul?`,
    };
  }

  const tier = profile.tiers.find((t) => t.code === kind);
  const name = tier?.name ?? kind.toUpperCase();
  if (!tier?.rank) {
    return {
      badge: `📚 Jenjang ${name}`,
      big: name,
      caption: "latihan matematika jadi game",
      detail: "Kumpulin poin dari tiap jawaban benar",
      text: `Aku lagi latihan matematika jenjang ${name} di MathQuest. Ikutan yuk!`,
    };
  }
  return {
    badge: `🏆 Ranking ${name}`,
    big: `#${tier.rank.rank}`,
    caption: `dari ${tier.rank.total.toLocaleString("id-ID")} murid ${name}`,
    detail: `${tier.points.toLocaleString("id-ID")} poin`,
    text: `🏆 Aku peringkat #${tier.rank.rank} jenjang ${name} di MathQuest dengan ${tier.points.toLocaleString("id-ID")} poin! Berani nyusul?`,
  };
}

// ---------- Hasil multiplayer ----------
// Link: /s/{nickname}/multiplayer/{matchId}. Server cuma nyimpen siapa aja
// yang main & rank-nya (tanpa skor), jadi kartunya cuma "menang lawan siapa".

export type MatchShare = {
  id: string;
  mode: "classic" | "race";
  questionCount: number;
  finishedAt: string;
  players: { rank: number; nickname: string; avatar: UserAvatar }[];
};

const MATCH_MODE_LABEL: Record<MatchShare["mode"], string> = { classic: "Klasik", race: "Adu Cepat" };

export function matchSharePath(nickname: string, matchId: string) {
  return `/s/${encodeURIComponent(nickname)}/multiplayer/${matchId}`;
}

export async function fetchMatchShare(matchId: string): Promise<MatchShare | null> {
  const res = await fetch(`${API_BASE_URL}/public/multiplayer/matches/${encodeURIComponent(matchId)}`, {
    next: { revalidate: 300 },
  });
  if (res.status === 404 || res.status === 400) return null;
  if (!res.ok) throw new Error(`match share: ${res.status}`);
  return res.json();
}

export function findMatchPlayer(match: MatchShare, nickname: string) {
  const n = nickname.toLowerCase();
  return match.players.find((p) => p.nickname.toLowerCase() === n) ?? null;
}

// "@budi, @cici, @dodi +2 lainnya"
function opponentNames(nicknames: string[]) {
  const shown = nicknames.slice(0, 3).map((n) => `@${n}`).join(", ");
  const rest = nicknames.length - 3;
  return rest > 0 ? `${shown} +${rest} lainnya` : shown;
}

// Dipakai juga di klien (tombol share di halaman hasil) -- bentuk datanya
// cukup rank + nickname tiap pemain.
export function buildMatchShareContent(
  mode: MatchShare["mode"],
  players: { rank: number; nickname: string }[],
  nickname: string,
): ShareContent {
  const n = nickname.toLowerCase();
  const me = players.find((p) => p.nickname.toLowerCase() === n);
  const rank = me?.rank ?? players.length;
  const opponents = opponentNames(players.filter((p) => p !== me).map((p) => p.nickname));
  const badge = `⚔️ Multiplayer ${MATCH_MODE_LABEL[mode]}`;

  if (rank === 1) {
    return {
      badge,
      big: "MENANG",
      caption: `lawan ${opponents}`,
      detail: `Juara 1 dari ${players.length} pemain`,
      text: `🏆 Aku menang lawan ${opponents} di Multiplayer MathQuest! Siapa berani nantang?`,
    };
  }
  return {
    badge,
    big: `#${rank}`,
    caption: `lawan ${opponents}`,
    detail: `dari ${players.length} pemain`,
    text: `⚔️ Aku peringkat #${rank} dari ${players.length} pemain lawan ${opponents} di Multiplayer MathQuest. Revans yuk!`,
  };
}
