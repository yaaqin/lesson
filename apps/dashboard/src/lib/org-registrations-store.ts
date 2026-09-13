"use client";

import { useSyncExternalStore } from "react";

// Simulasi antrean approval pendaftaran organisasi — dummy, localStorage.
// Di backend beneran ini bagian dari POST /api/v1/org/organizations/register
// + admin approval flow (belum ada di FSD.md, ditambahkan di sesi ini).

const STORAGE_KEY = "mathquest_dashboard_org_registrations_v1";

export type OrgRegistration = {
  id: string;
  organizationName: string;
  ownerName: string;
  ownerEmail: string;
  password: string;
  status: "pending" | "approved" | "rejected";
  createdAt: string;
};

let cache: OrgRegistration[] | null = null;
const listeners = new Set<() => void>();

function read(): OrgRegistration[] {
  try {
    const raw = window.localStorage.getItem(STORAGE_KEY);
    return raw ? (JSON.parse(raw) as OrgRegistration[]) : [];
  } catch {
    return [];
  }
}

function write(next: OrgRegistration[]) {
  cache = next;
  window.localStorage.setItem(STORAGE_KEY, JSON.stringify(next));
  listeners.forEach((listener) => listener());
}

export function getRegistrations(): OrgRegistration[] {
  if (cache === null) cache = read();
  return cache;
}

function subscribe(listener: () => void) {
  listeners.add(listener);
  return () => listeners.delete(listener);
}

function getServerSnapshot(): OrgRegistration[] {
  return [];
}

function generateId() {
  return `org-req-${Math.random().toString(36).slice(2, 9)}`;
}

export function submitRegistration(input: {
  organizationName: string;
  ownerName: string;
  ownerEmail: string;
  password: string;
}): OrgRegistration {
  const registration: OrgRegistration = {
    id: generateId(),
    status: "pending",
    createdAt: new Date().toISOString(),
    ...input,
  };
  write([...getRegistrations(), registration]);
  return registration;
}

export function approveRegistration(id: string) {
  write(getRegistrations().map((r) => (r.id === id ? { ...r, status: "approved" } : r)));
}

export function rejectRegistration(id: string) {
  write(getRegistrations().map((r) => (r.id === id ? { ...r, status: "rejected" } : r)));
}

export function useRegistrations() {
  return useSyncExternalStore(subscribe, getRegistrations, getServerSnapshot);
}
