import type { UserAvatar as UserAvatarData } from "@/hooks/use-curriculum";
import { findAvatarCharacter } from "@/lib/avatars";

const SIZE = {
  sm: { box: "h-8 w-8", text: "text-lg" },
  md: { box: "h-12 w-12", text: "text-2xl" },
  lg: { box: "h-24 w-24", text: "text-5xl" },
};

export function UserAvatar({ avatar, size = "md" }: { avatar: UserAvatarData; size?: keyof typeof SIZE }) {
  const s = SIZE[size];
  if (avatar.type === "google" && avatar.googleUrl) {
    return (
      // eslint-disable-next-line @next/next/no-img-element -- foto Google eksternal, gak perlu optimasi next/image
      <img
        src={avatar.googleUrl}
        alt="Foto profil"
        referrerPolicy="no-referrer"
        className={`${s.box} shrink-0 rounded-full object-cover`}
      />
    );
  }
  const character = findAvatarCharacter(avatar.key);
  return (
    <span
      role="img"
      aria-label={character.label}
      className={`${s.box} ${s.text} ${character.bg} flex shrink-0 items-center justify-center rounded-full`}
    >
      {character.emoji}
    </span>
  );
}
