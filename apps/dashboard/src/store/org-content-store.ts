import { create } from "zustand";
import { createJSONStorage, persist } from "zustand/middleware";

// Kurikulum organisasi — dummy, localStorage (Zustand persist), per organizationId.
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

function generateId(prefix: string) {
  return `${prefix}-${Math.random().toString(36).slice(2, 9)}`;
}

type State = {
  contentByOrg: Record<string, OrgKelas[]>;
  addKelas: (organizationId: string, name: string) => void;
  addGroup: (organizationId: string, kelasId: string, name: string) => void;
  addChallenge: (
    organizationId: string,
    kelasId: string,
    groupId: string,
    name: string,
    isExam: boolean,
  ) => void;
};

const useOrgContentStore = create<State>()(
  persist(
    (set, get) => ({
      contentByOrg: {},
      addKelas: (organizationId, name) => {
        const current = get().contentByOrg[organizationId] ?? seedContent[organizationId] ?? [];
        set({
          contentByOrg: {
            ...get().contentByOrg,
            [organizationId]: [...current, { id: generateId("kelas"), name, groups: [] }],
          },
        });
      },
      addGroup: (organizationId, kelasId, name) => {
        const current = get().contentByOrg[organizationId] ?? seedContent[organizationId] ?? [];
        set({
          contentByOrg: {
            ...get().contentByOrg,
            [organizationId]: current.map((kelas) =>
              kelas.id === kelasId
                ? {
                    ...kelas,
                    groups: [
                      ...kelas.groups,
                      { id: generateId("group"), name, challenges: [] },
                    ],
                  }
                : kelas,
            ),
          },
        });
      },
      addChallenge: (organizationId, kelasId, groupId, name, isExam) => {
        const current = get().contentByOrg[organizationId] ?? seedContent[organizationId] ?? [];
        set({
          contentByOrg: {
            ...get().contentByOrg,
            [organizationId]: current.map((kelas) =>
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
          },
        });
      },
    }),
    {
      name: "mathquest_dashboard_org_content_v1",
      storage: createJSONStorage(() => localStorage),
    },
  ),
);

export function useOrgContent(organizationId: string) {
  const kelasList = useOrgContentStore(
    (s) => s.contentByOrg[organizationId] ?? seedContent[organizationId] ?? [],
  );
  const addKelas = useOrgContentStore((s) => s.addKelas);
  const addGroup = useOrgContentStore((s) => s.addGroup);
  const addChallenge = useOrgContentStore((s) => s.addChallenge);

  return {
    kelasList,
    addKelas: (name: string) => addKelas(organizationId, name),
    addGroup: (kelasId: string, name: string) => addGroup(organizationId, kelasId, name),
    addChallenge: (kelasId: string, groupId: string, name: string, isExam: boolean) =>
      addChallenge(organizationId, kelasId, groupId, name, isExam),
  };
}
