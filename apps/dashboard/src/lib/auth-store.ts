"use client";

import { useSyncExternalStore } from "react";
import { apiLogin, apiLogout } from "./api-client";
import { DUMMY_ACCOUNTS, type OrgRole, type PlatformRole } from "./dummy-accounts";
import { getRegistrations } from "./org-registrations-store";

const SESSION_KEY = "mathquest_dashboard_session_v1";

export type Session =
  | {
      area: "admin";
      email: string;
      displayName: string;
      platformRole: PlatformRole;
      accessToken: string;
      refreshToken: string;
    }
  | {
      area: "org";
      email: string;
      displayName: string;
      orgRole: OrgRole;
      organizationId: string;
      organizationName: string;
    };

const UNSET = Symbol("unset");
let cache: Session | null | typeof UNSET = UNSET;
const listeners = new Set<() => void>();

function readSession(): Session | null {
  try {
    const raw = window.localStorage.getItem(SESSION_KEY);
    return raw ? (JSON.parse(raw) as Session) : null;
  } catch {
    return null;
  }
}

function getSnapshot(): Session | null {
  if (cache === UNSET) cache = readSession();
  return cache;
}

function getServerSnapshot(): Session | null {
  return null;
}

function subscribe(listener: () => void) {
  listeners.add(listener);
  return () => listeners.delete(listener);
}

function setSession(next: Session | null) {
  cache = next;
  if (next) {
    window.localStorage.setItem(SESSION_KEY, JSON.stringify(next));
  } else {
    window.localStorage.removeItem(SESSION_KEY);
  }
  listeners.forEach((listener) => listener());
}

export type LoginResult =
  | { ok: true; session: Session }
  | { ok: false; reason: "invalid" | "pending" | "rejected" };

export async function login(email: string, password: string): Promise<LoginResult> {
  const normalizedEmail = email.trim().toLowerCase();

  // 1) coba lewat backend Go beneran dulu — ini jalur akun admin platform.
  const apiResult = await apiLogin(normalizedEmail, password);
  if (apiResult.ok) {
    const { user, accessToken, refreshToken } = apiResult.data;
    const session: Session = {
      area: "admin",
      email: user.email,
      displayName: user.displayName,
      platformRole: user.role === "student" ? "admin" : user.role,
      accessToken,
      refreshToken,
    };
    setSession(session);
    return { ok: true, session };
  }

  // 2) fallback ke akun dummy organisasi (backend org belum ada)
  const seedMatch = DUMMY_ACCOUNTS.find(
    (acc) => acc.email.toLowerCase() === normalizedEmail && acc.password === password,
  );
  if (seedMatch) {
    const session: Session = {
      area: "org",
      email: seedMatch.email,
      displayName: seedMatch.displayName,
      orgRole: seedMatch.orgRole,
      organizationId: seedMatch.organizationId,
      organizationName: seedMatch.organizationName,
    };
    setSession(session);
    return { ok: true, session };
  }

  const registration = getRegistrations().find(
    (r) => r.ownerEmail.toLowerCase() === normalizedEmail && r.password === password,
  );
  if (!registration) return { ok: false, reason: "invalid" };
  if (registration.status === "pending") return { ok: false, reason: "pending" };
  if (registration.status === "rejected") return { ok: false, reason: "rejected" };

  const session: Session = {
    area: "org",
    email: registration.ownerEmail,
    displayName: registration.ownerName,
    orgRole: "owner",
    organizationId: registration.id,
    organizationName: registration.organizationName,
  };
  setSession(session);
  return { ok: true, session };
}

export function logout() {
  const current = getSnapshot();
  setSession(null);
  if (current?.area === "admin") {
    apiLogout(current.accessToken);
  }
}

export function useSession() {
  return useSyncExternalStore(subscribe, getSnapshot, getServerSnapshot);
}
