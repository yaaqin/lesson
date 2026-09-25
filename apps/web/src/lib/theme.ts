// Tema tampilan: light | dark | system. Class `dark` di <html> yang nyalain
// varian `dark:` Tailwind (lihat @custom-variant di globals.css). Pilihan
// disimpen di localStorage biar langsung kepake sebelum render (tanpa kedip),
// dan di akun (users.theme_preference) biar kebawa ke device lain.

export type ThemePreference = "light" | "dark" | "system";

export const THEME_STORAGE_KEY = "theme";

// Warna status bar HP (sama kayak viewport.themeColor di layout.tsx).
const THEME_COLOR = { light: "#fafafa", dark: "#000000" };

export function readStoredTheme(): ThemePreference {
  try {
    const v = localStorage.getItem(THEME_STORAGE_KEY);
    return v === "light" || v === "dark" ? v : "system";
  } catch {
    return "system";
  }
}

export function storeTheme(pref: ThemePreference) {
  try {
    localStorage.setItem(THEME_STORAGE_KEY, pref);
  } catch {
    // storage diblokir -- tema tetap kepake buat sesi ini
  }
}

export function applyTheme(pref: ThemePreference) {
  const dark = pref === "dark" || (pref === "system" && window.matchMedia("(prefers-color-scheme: dark)").matches);
  document.documentElement.classList.toggle("dark", dark);
  // Pilihan manual: semua meta theme-color (light & dark) disamain ke tema
  // aktif; "system" balikin ke masing-masing media query-nya.
  document.querySelectorAll<HTMLMetaElement>('meta[name="theme-color"]').forEach((meta) => {
    const media = meta.getAttribute("media") ?? "";
    meta.content =
      pref === "system"
        ? media.includes("dark")
          ? THEME_COLOR.dark
          : THEME_COLOR.light
        : THEME_COLOR[dark ? "dark" : "light"];
  });
}

// Dijalanin inline di <head> sebelum halaman kerender (lihat layout.tsx).
export const THEME_INIT_SCRIPT = `(function(){try{var t=localStorage.getItem("${THEME_STORAGE_KEY}");var d=t==="dark"||(t!=="light"&&window.matchMedia("(prefers-color-scheme: dark)").matches);if(d)document.documentElement.classList.add("dark")}catch(e){}})()`;
