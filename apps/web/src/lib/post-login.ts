// Tujuan setelah login: dipakai link undangan room multiplayer -- user yang
// belum login dilempar ke /login (dan mungkin /onboarding, atau muter lewat
// Google OAuth), abis itu harus balik ke room-nya, bukan ke /belajar.
// Disimpen di sessionStorage (bukan query) biar selamat lewat redirect Google.
const KEY = "mathquest_post_login_v1";
const TTL_MS = 30 * 60 * 1000;

export function rememberPostLoginPath(path: string) {
  try {
    sessionStorage.setItem(KEY, JSON.stringify({ path, at: Date.now() }));
  } catch {
    // storage diblok -> ya udah, balik ke default aja nanti
  }
}

// takePostLoginPath: cuma baca (onboarding bisa manggil 2x pas nickname
// kesimpen), dihapusnya sama halaman tujuan lewat clearPostLoginPath.
export function takePostLoginPath(fallback: string) {
  try {
    const raw = sessionStorage.getItem(KEY);
    if (!raw) return fallback;
    const { path, at } = JSON.parse(raw) as { path: string; at: number };
    if (Date.now() - at > TTL_MS || !path.startsWith("/") || path.startsWith("//")) return fallback;
    return path;
  } catch {
    return fallback;
  }
}

export function clearPostLoginPath() {
  try {
    sessionStorage.removeItem(KEY);
  } catch {
    // abaikan
  }
}
