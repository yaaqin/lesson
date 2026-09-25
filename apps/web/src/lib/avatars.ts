// Karakter kartun buat avatar profil. Key-nya harus sinkron sama
// AvatarCharacters di services/api internal/curriculumsvc/profile.go.
export type AvatarCharacter = { key: string; emoji: string; label: string; bg: string };

export const AVATAR_CHARACTERS: AvatarCharacter[] = [
  { key: "fox", emoji: "🦊", label: "Rubah", bg: "bg-orange-100 dark:bg-orange-500/20" },
  { key: "panda", emoji: "🐼", label: "Panda", bg: "bg-zinc-200 dark:bg-zinc-500/30" },
  { key: "tiger", emoji: "🐯", label: "Harimau", bg: "bg-amber-100 dark:bg-amber-500/20" },
  { key: "frog", emoji: "🐸", label: "Katak", bg: "bg-green-100 dark:bg-green-500/20" },
  { key: "monkey", emoji: "🐵", label: "Monyet", bg: "bg-yellow-100 dark:bg-yellow-500/20" },
  { key: "penguin", emoji: "🐧", label: "Penguin", bg: "bg-sky-100 dark:bg-sky-500/20" },
  { key: "lion", emoji: "🦁", label: "Singa", bg: "bg-yellow-100 dark:bg-yellow-500/20" },
  { key: "koala", emoji: "🐨", label: "Koala", bg: "bg-slate-200 dark:bg-slate-500/30" },
  { key: "rabbit", emoji: "🐰", label: "Kelinci", bg: "bg-pink-100 dark:bg-pink-500/20" },
  { key: "bear", emoji: "🐻", label: "Beruang", bg: "bg-amber-100 dark:bg-amber-500/20" },
  { key: "unicorn", emoji: "🦄", label: "Unicorn", bg: "bg-fuchsia-100 dark:bg-fuchsia-500/20" },
  { key: "octopus", emoji: "🐙", label: "Gurita", bg: "bg-rose-100 dark:bg-rose-500/20" },
  { key: "cat", emoji: "🐱", label: "Kucing", bg: "bg-orange-100 dark:bg-orange-500/20" },
  { key: "dog", emoji: "🐶", label: "Anjing", bg: "bg-amber-100 dark:bg-amber-500/20" },
  { key: "owl", emoji: "🦉", label: "Burung Hantu", bg: "bg-stone-200 dark:bg-stone-500/30" },
  { key: "dino", emoji: "🦖", label: "Dino", bg: "bg-emerald-100 dark:bg-emerald-500/20" },
];

export function findAvatarCharacter(key: string) {
  return AVATAR_CHARACTERS.find((c) => c.key === key) ?? AVATAR_CHARACTERS[0];
}
