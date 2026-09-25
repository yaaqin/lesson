// Aturan nickname -- sama persis sama NormalizeNickname di services/api
// internal/curriculumsvc/profile.go & CHECK constraint users.username.
export const NICKNAME_MIN = 6;
export const NICKNAME_MAX = 20;
const NICKNAME_PATTERN = /^[a-z0-9_.]+$/;

// Huruf besar otomatis dikecilin (server juga ngelakuin hal yang sama).
export function normalizeNicknameInput(raw: string) {
  return raw.trim().toLowerCase();
}

// Balikin pesan error, atau null kalau formatnya valid.
export function nicknameFormatError(nickname: string): string | null {
  if (nickname.length === 0) return "Nickname wajib diisi.";
  if (!NICKNAME_PATTERN.test(nickname)) return "Cuma boleh huruf a-z, angka 0-9, _ dan titik (.)";
  if (nickname.length < NICKNAME_MIN) return `Minimal ${NICKNAME_MIN} karakter.`;
  if (nickname.length > NICKNAME_MAX) return `Maksimal ${NICKNAME_MAX} karakter.`;
  return null;
}
