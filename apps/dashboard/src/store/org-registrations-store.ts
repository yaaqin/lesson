import { create } from "zustand";
import { createJSONStorage, persist } from "zustand/middleware";

// Simulasi antrean approval pendaftaran organisasi — dummy, localStorage (Zustand persist).
// Di backend beneran ini bagian dari POST /api/v1/org/organizations/register
// + admin approval flow (belum ada di FSD.md, ditambahkan di sesi ini).

export type OrgRegistration = {
  id: string;
  organizationName: string;
  ownerName: string;
  ownerEmail: string;
  password: string;
  status: "pending" | "approved" | "rejected";
  createdAt: string;
};

type State = {
  registrations: OrgRegistration[];
  submit: (input: {
    organizationName: string;
    ownerName: string;
    ownerEmail: string;
    password: string;
  }) => OrgRegistration;
  approve: (id: string) => void;
  reject: (id: string) => void;
};

function generateId() {
  return `org-req-${Math.random().toString(36).slice(2, 9)}`;
}

const useOrgRegistrationsStore = create<State>()(
  persist(
    (set, get) => ({
      registrations: [],
      submit: (input) => {
        const registration: OrgRegistration = {
          id: generateId(),
          status: "pending",
          createdAt: new Date().toISOString(),
          ...input,
        };
        set({ registrations: [...get().registrations, registration] });
        return registration;
      },
      approve: (id) =>
        set({
          registrations: get().registrations.map((r) =>
            r.id === id ? { ...r, status: "approved" } : r,
          ),
        }),
      reject: (id) =>
        set({
          registrations: get().registrations.map((r) =>
            r.id === id ? { ...r, status: "rejected" } : r,
          ),
        }),
    }),
    {
      name: "mathquest_dashboard_org_registrations_v1",
      storage: createJSONStorage(() => localStorage),
    },
  ),
);

export function useRegistrations() {
  return useOrgRegistrationsStore((s) => s.registrations);
}

export function getRegistrations(): OrgRegistration[] {
  return useOrgRegistrationsStore.getState().registrations;
}

export function submitRegistration(input: {
  organizationName: string;
  ownerName: string;
  ownerEmail: string;
  password: string;
}): OrgRegistration {
  return useOrgRegistrationsStore.getState().submit(input);
}

export function approveRegistration(id: string) {
  useOrgRegistrationsStore.getState().approve(id);
}

export function rejectRegistration(id: string) {
  useOrgRegistrationsStore.getState().reject(id);
}
