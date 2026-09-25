"use client";

import { useState, type FormEvent } from "react";
import type { MeInfo } from "@/hooks/use-curriculum";
import { useUpdateProfileMutation, type ProfileUpdate } from "@/hooks/use-profile";
import { AvatarPicker, type AvatarChoice } from "@/components/avatar-picker";
import { NicknameField, type NicknameStatus } from "@/components/nickname-field";
import { UserAvatar } from "@/components/user-avatar";

const SAVE_ERROR: Record<string, string> = {
  nickname_taken: "Nickname ini udah dipakai user lain.",
  invalid_nickname: "Format nickname gak valid.",
  invalid_avatar: "Avatar gak valid.",
  no_google_avatar: "Akun ini gak punya foto Google.",
};

// Form nickname + avatar -- dipakai di /onboarding (nickname wajib) dan
// /profile (edit).
export function ProfileForm({
  me,
  submitLabel,
  onSaved,
}: {
  me: MeInfo;
  submitLabel: string;
  onSaved?: (me: MeInfo) => void;
}) {
  const updateMutation = useUpdateProfileMutation();
  const [nickname, setNickname] = useState(me.nickname ?? "");
  const [nicknameStatus, setNicknameStatus] = useState<NicknameStatus>("invalid");
  const [avatar, setAvatar] = useState<AvatarChoice>(
    me.avatar.type === "google" && me.avatar.googleUrl ? { type: "google" } : { type: "character", key: me.avatar.key },
  );
  const [error, setError] = useState<string | null>(null);
  const [saved, setSaved] = useState(false);

  const avatarChanged =
    avatar.type !== me.avatar.type || (avatar.type === "character" && avatar.key !== me.avatar.key);
  const nicknameChanged = nickname !== (me.nickname ?? "");
  const nicknameOk = nicknameStatus === "available" || nicknameStatus === "unchanged";
  const canSubmit = nicknameOk && (nicknameChanged || avatarChanged || me.nickname === null) && !updateMutation.isPending;

  const handleSubmit = (e: FormEvent) => {
    e.preventDefault();
    if (!canSubmit) return;
    setError(null);
    setSaved(false);
    const update: ProfileUpdate = {};
    if (nicknameChanged) update.nickname = nickname;
    if (avatarChanged || me.nickname === null) {
      update.avatarType = avatar.type;
      if (avatar.type === "character") update.avatarKey = avatar.key;
    }
    updateMutation.mutate(update, {
      onSuccess: (next) => {
        setSaved(true);
        onSaved?.(next);
      },
      onError: (err: unknown) => {
        const code = (err as { response?: { data?: { error?: string } } })?.response?.data?.error ?? "";
        setError(SAVE_ERROR[code] ?? "Gagal nyimpen profil, coba lagi.");
      },
    });
  };

  const preview = avatar.type === "google" ? { ...me.avatar, type: "google" as const } : { ...me.avatar, type: "character" as const, key: avatar.key };

  return (
    <form onSubmit={handleSubmit} className="flex flex-col gap-6">
      <div className="flex items-center gap-4">
        <UserAvatar avatar={preview} size="lg" />
        <div className="flex min-w-0 flex-col">
          <span className="truncate text-lg font-semibold text-black dark:text-zinc-50">
            {nickname ? `@${nickname}` : "@nickname"}
          </span>
          <span className="truncate text-sm text-zinc-500">{me.email}</span>
        </div>
      </div>

      <NicknameField value={nickname} onChange={setNickname} current={me.nickname} onStatus={setNicknameStatus} />

      <div className="flex flex-col gap-2">
        <span className="text-sm font-medium text-zinc-700 dark:text-zinc-300">Karakter profil</span>
        <AvatarPicker value={avatar} googleUrl={me.avatar.googleUrl} onChange={setAvatar} />
      </div>

      {error && (
        <p className="rounded-lg bg-red-50 px-3 py-2 text-sm text-red-600 dark:bg-red-500/10 dark:text-red-400">{error}</p>
      )}
      {saved && !error && !onSaved && (
        <p className="rounded-lg bg-green-50 px-3 py-2 text-sm text-green-700 dark:bg-green-500/10 dark:text-green-400">
          Profil tersimpan.
        </p>
      )}

      <button
        type="submit"
        disabled={!canSubmit}
        className="rounded-full bg-foreground px-6 py-3 text-sm font-medium text-background transition-colors hover:bg-[#383838] disabled:opacity-40 dark:hover:bg-[#ccc]"
      >
        {updateMutation.isPending ? "Menyimpan…" : submitLabel}
      </button>
    </form>
  );
}
