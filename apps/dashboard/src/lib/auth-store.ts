"use client";

import { useSyncExternalStore } from "react";
import { DUMMY_ACCOUNTS, type DummyAccount, type OrgRole, type PlatformRole } from "./dummy-accounts";
import { getRegistrations } from "./org-registrations-store";

const SESSION_KEY = "mathquest_dashboard_session_v1";

export type Session =
  | {
      area: "admin";
      email: string;
      displayName: string;
      platformRole: PlatformRole;
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

function toSession(account: DummyAccount): Session {
  if (account.area === "admin") {
    return {
      area: "admin",
      email: account.email,
      displayName: account.displayName,
      platformRole: account.platformRole,
    };
  }
  return {
    area: "org",
    email: account.email,
    displayName: account.displayName,
    orgRole: account.orgRole,
    organizationId: account.organizationId,
    organizationName: account.organizationName,
  };
}

export type LoginResult =
  | { ok: true; session: Session }
  | { ok: false; reason: "invalid" | "pending" | "rejected" };

export function login(email: string, password: string): LoginResult {
  const normalizedEmail = email.trim().toLowerCase();

  const seedMatch = DUMMY_ACCOUNTS.find(
    (acc) => acc.email.toLowerCase() === normalizedEmail && acc.password === password,
  );
  if (seedMatch) {
    const session = toSession(seedMatch);
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
  setSession(null);
}

export function useSession() {
  return useSyncExternalStore(subscribe, getSnapshot, getServerSnapshot);
}
