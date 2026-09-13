// Akun dummy — belum tersambung ke Postgres/Go API.
// Password sengaja plaintext & sederhana karena ini cuma buat demo dashboard.

export type PlatformRole = "superadmin" | "admin";
export type OrgRole = "owner" | "admin" | "teacher";

export type DummyAccount =
  | {
      area: "admin";
      email: string;
      password: string;
      displayName: string;
      platformRole: PlatformRole;
    }
  | {
      area: "org";
      email: string;
      password: string;
      displayName: string;
      orgRole: OrgRole;
      organizationId: string;
      organizationName: string;
    };

export const DUMMY_ACCOUNTS: DummyAccount[] = [
  {
    area: "admin",
    email: "superadmin@mathquest.dev",
    password: "password123",
    displayName: "Bimo",
    platformRole: "superadmin",
  },
  {
    area: "org",
    email: "owner@sekolahsatu.sch.id",
    password: "password123",
    displayName: "Rina",
    orgRole: "owner",
    organizationId: "org-1",
    organizationName: "SD Sekolah Satu",
  },
  {
    area: "org",
    email: "admin@sekolahsatu.sch.id",
    password: "password123",
    displayName: "Budi",
    orgRole: "admin",
    organizationId: "org-1",
    organizationName: "SD Sekolah Satu",
  },
  {
    area: "org",
    email: "guru@sekolahsatu.sch.id",
    password: "password123",
    displayName: "Sari",
    orgRole: "teacher",
    organizationId: "org-1",
    organizationName: "SD Sekolah Satu",
  },
];

export const ORG_ROLE_LABEL: Record<OrgRole, string> = {
  owner: "Pemilik",
  admin: "Admin Organisasi",
  teacher: "Guru",
};

export const PLATFORM_ROLE_LABEL: Record<PlatformRole, string> = {
  superadmin: "Superadmin",
  admin: "Admin Platform",
};
