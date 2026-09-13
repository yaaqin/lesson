"use client";

import { useCallback, useSyncExternalStore } from "react";

// Kurikulum organisasi — dummy, localStorage, per organizationId.
// Hierarki: Kelas -> Group -> Challenge (salah satu challenge per group bisa ditandai "ujian").

export type OrgChallenge = {
  id: string;
  name: string;
  isExam: boolean;
};

export type OrgGroup = {
  id: string;
  name: string;
  challenges: OrgChallenge[];
};

export type OrgKelas = {
  id: string;
  name: string;
  groups: OrgGroup[];
};

const STORAGE_PREFIX = "mathquest_dashboard_org_content_v1::";

const seedContent: Record<string, OrgKelas[]> = {
  "org-1": [
    {
      id: "kelas-1",
      name: "Kelas 10 A",
      groups: [
        {
          id: "group-1",
          name: "Bab 1 — Aljabar Dasar",
          challenges: [
            { id: "ch-1", name: "Latihan Persamaan Linear", isExam: false },
            { id: "ch-2", name: "Latihan Pertidaksamaan", isExam: false },
            { id: "ch-3", name: "Ujian Bab 1", isExam: true },
          ],
        },
      ],
    },
  ],
};

const cache = new Map<string, OrgKelas[]>();
const listeners = new Map<string, Set<() => void>>();

function storageKey(organizationId: string) {
  return `${STORAGE_PREFIX}${organizationId}`;
}

function readFromStorage(organizationId: string): OrgKelas[] {
  try {
    const raw = window.localStorage.getItem(storageKey(organizationId));
    if (raw) return JSON.parse(raw) as OrgKelas[];
  } catch {
    // ignore, fall through to seed/empty
  }
  return seedContent[organizationId] ?? [];
}

function getSnapshot(organizationId: string): OrgKelas[] {
  if (!cache.has(organizationId)) {
    cache.set(organizationId, readFromStorage(organizationId));
  }
  return cache.get(organizationId) as OrgKelas[];
}

function getServerSnapshot(organizationId: string): OrgKelas[] {
  return seedContent[organizationId] ?? [];
}

function writeState(organizationId: string, next: OrgKelas[]) {
  cache.set(organizationId, next);
  window.localStorage.setItem(storageKey(organizationId), JSON.stringify(next));
  listeners.get(organizationId)?.forEach((listener) => listener());
}

function subscribe(organizationId: string, listener: () => void) {
  if (!listeners.has(organizationId)) listeners.set(organizationId, new Set());
  const set = listeners.get(organizationId) as Set<() => void>;
  set.add(listener);
  return () => set.delete(listener);
}

function generateId(prefix: string) {
  return `${prefix}-${Math.random().toString(36).slice(2, 9)}`;
}

export function useOrgContent(organizationId: string) {
  const kelasList = useSyncExternalStore(
    useCallback((listener: () => void) => subscribe(organizationId, listener), [organizationId]),
    useCallback(() => getSnapshot(organizationId), [organizationId]),
    useCallback(() => getServerSnapshot(organizationId), [organizationId]),
  );

  const addKelas = useCallback(
    (name: string) => {
      const current = getSnapshot(organizationId);
      writeState(organizationId, [
        ...current,
        { id: generateId("kelas"), name, groups: [] },
      ]);
    },
    [organizationId],
  );

  const addGroup = useCallback(
    (kelasId: string, name: string) => {
      const current = getSnapshot(organizationId);
      writeState(
        organizationId,
        current.map((kelas) =>
          kelas.id === kelasId
            ? {
                ...kelas,
                groups: [...kelas.groups, { id: generateId("group"), name, challenges: [] }],
              }
            : kelas,
        ),
      );
    },
    [organizationId],
  );

  const addChallenge = useCallback(
    (kelasId: string, groupId: string, name: string, isExam: boolean) => {
      const current = getSnapshot(organizationId);
      writeState(
        organizationId,
        current.map((kelas) =>
          kelas.id !== kelasId
            ? kelas
            : {
                ...kelas,
                groups: kelas.groups.map((group) =>
                  group.id !== groupId
                    ? group
                    : {
                        ...group,
                        // kalau challenge baru dijadikan ujian, lepas status ujian dari yang lama
                        // (cuma boleh 1 ujian per group)
                        challenges: [
                          ...group.challenges.map((c) =>
                            isExam ? { ...c, isExam: false } : c,
                          ),
                          { id: generateId("ch"), name, isExam },
                        ],
                      },
                ),
              },
        ),
      );
    },
    [organizationId],
  );

  return { kelasList, addKelas, addGroup, addChallenge };
}
