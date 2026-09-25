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
