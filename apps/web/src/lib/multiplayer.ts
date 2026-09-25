import type { MultiplayerFormat, MultiplayerMode, MultiplayerTierCode } from "@/hooks/use-multiplayer";

export const MODE_INFO: Record<MultiplayerMode, { label: string; icon: string; desc: string }> = {
  classic: {
    label: "Klasik",
    icon: "🎯",
    desc: "Semua jawab soal yang sama barengan. Benar + makin cepat = poin makin gede. Soal lanjut pas waktunya habis.",
  },
  race: {
    label: "Adu Cepat",
    icon: "⚡",
    desc: "Siapa yang duluan jawab benar dapet poin, soal langsung lanjut. Salah = kekunci di soal itu.",
  },
};

export const FORMAT_LABEL: Record<MultiplayerFormat, string> = {
  mc: "Pilihan ganda",
  essay: "Isian",
  mixed: "Campuran",
};

export const TIER_LABEL: Record<MultiplayerTierCode, string> = {
  sd: "SD",
  smp: "SMP",
  smk: "SMK / SMA",
  kampus: "Kampus",
};

export const TIER_BADGE: Record<MultiplayerTierCode, string> = {
  sd: "bg-sky-100 text-sky-700 dark:bg-sky-500/10 dark:text-sky-400",
  smp: "bg-emerald-100 text-emerald-700 dark:bg-emerald-500/10 dark:text-emerald-400",
  smk: "bg-amber-100 text-amber-700 dark:bg-amber-500/10 dark:text-amber-400",
  kampus: "bg-purple-100 text-purple-700 dark:bg-purple-500/10 dark:text-purple-400",
};

// Kode room: 6 karakter dari alfabet tanpa 0/O/1/I/L (sama kayak server).
export function normalizeRoomCode(raw: string) {
  return raw.toUpperCase().replace(/[^A-Z0-9]/g, "").slice(0, 6);
}

export const JOIN_ERROR_TEXT: Record<string, string> = {
  room_not_found: "Room gak ketemu. Cek lagi kodenya, atau room-nya udah ditutup.",
  room_full: "Room udah penuh.",
  game_started: "Game di room ini udah mulai.",
  kicked: "Kamu dikeluarin dari room ini sama host.",
  nickname_required: "Bikin nickname dulu sebelum main.",
};
