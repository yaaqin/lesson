"use client";

import { usePathname, useRouter } from "next/navigation";
import { useCallback, useEffect } from "react";

// Riwayat navigasi in-app (stack pathname) biar tombol back bisa pakai
// router.back() -- bukan push ke path tetap yang bikin history muter-muter
// (A -> B -> back(push A) -> back(push B) -> ...). Kalau gak ada riwayat
// in-app (buka link langsung / abis login), baru jatuh ke `fallback` pakai
// replace, jadi tetap gak nambah entry baru.
// Disimpen di sessionStorage biar selamat pas refresh (history browser juga
// tetap ada pas refresh).
const KEY = "mathquest_nav_stack_v1";
const MAX = 50;

// Halaman auth gak dihitung: masuknya lewat replace (redirect) jadi isi
// history aslinya gak ketebak -- stack direset biar back gak nyasar ke /login.
const isAuthPath = (path: string) =>
  path === "/login" || path === "/onboarding" || path.startsWith("/auth/");

function readStack(): string[] {
  try {
    const raw = sessionStorage.getItem(KEY);
    return raw ? (JSON.parse(raw) as string[]) : [];
  } catch {
    return [];
  }
}

function writeStack(stack: string[]) {
  try {
    sessionStorage.setItem(KEY, JSON.stringify(stack.slice(-MAX)));
  } catch {
    // storage diblok -> back selalu pakai fallback
  }
}

let poppedByBrowser = false;

// Dipasang sekali di Providers.
export function NavHistoryTracker() {
  const pathname = usePathname();

  useEffect(() => {
    const onPopState = () => {
      poppedByBrowser = true;
    };
    window.addEventListener("popstate", onPopState);
    return () => window.removeEventListener("popstate", onPopState);
  }, []);

  useEffect(() => {
    const popped = poppedByBrowser;
    poppedByBrowser = false;
    if (isAuthPath(pathname)) {
      writeStack([]);
      return;
    }
    const stack = readStack();
    if (stack[stack.length - 1] === pathname) return;
    if (popped && stack[stack.length - 2] === pathname) stack.pop();
    else stack.push(pathname);
    writeStack(stack);
  }, [pathname]);

  return null;
}

export function useGoBack(fallback: string) {
  const router = useRouter();
  return useCallback(() => {
    if (readStack().length > 1) router.back();
    else router.replace(fallback);
  }, [router, fallback]);
}
