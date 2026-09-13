"use client";

import { useRouter } from "next/navigation";
import { useEffect, useState, type FormEvent } from "react";
import { logout, useSession, type Session } from "@/lib/auth-store";
import { useOrgContent } from "@/lib/org-content-store";
import { ORG_ROLE_LABEL } from "@/lib/dummy-accounts";

export default function OrgDashboardPage() {
  const router = useRouter();
  const session = useSession();

  useEffect(() => {
    if (session === null) {
      router.replace("/login");
      return;
    }
    if (session.area !== "org") {
      router.replace("/admin");
    }
  }, [session, router]);

  if (!session || session.area !== "org") {
    return (
      <div className="flex min-h-screen items-center justify-center bg-zinc-50 dark:bg-black">
        <p className="text-zinc-500 dark:text-zinc-500">Memuat…</p>
      </div>
    );
  }

  return <OrgDashboard session={session} />;
}

type OrgSession = Extract<Session, { area: "org" }>;

function OrgDashboard({ session }: { session: OrgSession }) {
  const router = useRouter();
  const { kelasList, addKelas, addGroup, addChallenge } = useOrgContent(session.organizationId);

  const [newKelasName, setNewKelasName] = useState("");
  const [showKelasForm, setShowKelasForm] = useState(false);

  const [groupFormForKelas, setGroupFormForKelas] = useState<string | null>(null);
  const [newGroupName, setNewGroupName] = useState("");

  const [challengeFormFor, setChallengeFormFor] = useState<{
    kelasId: string;
    groupId: string;
  } | null>(null);
  const [newChallengeName, setNewChallengeName] = useState("");
  const [newChallengeIsExam, setNewChallengeIsExam] = useState(false);

  const submitKelas = (event: FormEvent) => {
    event.preventDefault();
    if (!newKelasName.trim()) return;
    addKelas(newKelasName.trim());
    setNewKelasName("");
    setShowKelasForm(false);
  };

  const submitGroup = (event: FormEvent, kelasId: string) => {
    event.preventDefault();
    if (!newGroupName.trim()) return;
    addGroup(kelasId, newGroupName.trim());
    setNewGroupName("");
    setGroupFormForKelas(null);
  };

  const submitChallenge = (event: FormEvent, kelasId: string, groupId: string) => {
    event.preventDefault();
    if (!newChallengeName.trim()) return;
    addChallenge(kelasId, groupId, newChallengeName.trim(), newChallengeIsExam);
    setNewChallengeName("");
    setNewChallengeIsExam(false);
    setChallengeFormFor(null);
  };

  return (
    <div className="flex min-h-screen flex-col bg-zinc-50 font-sans dark:bg-black">
      <header className="flex items-center justify-between px-6 py-5 sm:px-10">
        <div className="flex flex-col">
          <span className="text-lg font-semibold tracking-tight text-black dark:text-zinc-50">
            {session.organizationName}
          </span>
          <span className="text-xs text-zinc-500 dark:text-zinc-500">
            {session.displayName} · {ORG_ROLE_LABEL[session.orgRole]}
          </span>
        </div>
        <button
          type="button"
          onClick={() => {
            logout();
            router.push("/login");
          }}
          className="rounded-full border border-black/[.08] px-4 py-1.5 text-sm font-medium text-zinc-600 transition-colors hover:bg-black/[.04] dark:border-white/[.145] dark:text-zinc-400 dark:hover:bg-[#1a1a1a]"
        >
          Keluar
        </button>
      </header>

      <main className="mx-auto flex w-full max-w-3xl flex-1 flex-col gap-6 px-6 py-8">
        <div className="flex flex-col gap-1">
          <h1 className="text-2xl font-semibold tracking-tight text-black dark:text-zinc-50">
            Kurikulum
          </h1>
          <p className="text-sm text-zinc-500 dark:text-zinc-500">
            Kelas → Group → Challenge. Satu challenge di tiap group bisa ditandai jadi ujian
            akhir.
          </p>
        </div>

        {kelasList.length === 0 && !showKelasForm && (
          <p className="rounded-2xl border border-dashed border-black/[.08] px-5 py-6 text-center text-sm text-zinc-500 dark:border-white/[.145] dark:text-zinc-500">
            Belum ada kelas. Mulai dengan menambah kelas pertama.
          </p>
        )}

        <div className="flex flex-col gap-4">
          {kelasList.map((kelas) => (
            <div
              key={kelas.id}
              className="flex flex-col gap-3 rounded-2xl border border-black/[.08] bg-white p-5 dark:border-white/[.145] dark:bg-zinc-900"
            >
              <div className="flex items-center justify-between">
                <span className="font-semibold text-black dark:text-zinc-50">{kelas.name}</span>
                <button
                  type="button"
                  onClick={() =>
                    setGroupFormForKelas(groupFormForKelas === kelas.id ? null : kelas.id)
                  }
                  className="text-sm font-medium text-blue-600 dark:text-blue-400"
                >
                  + Tambah Group
                </button>
              </div>

              {groupFormForKelas === kelas.id && (
                <form
                  onSubmit={(e) => submitGroup(e, kelas.id)}
                  className="flex items-center gap-2"
                >
                  <input
                    autoFocus
                    value={newGroupName}
                    onChange={(e) => setNewGroupName(e.target.value)}
                    placeholder="mis. Bab 2 — Geometri"
                    className="flex-1 rounded-lg border border-black/[.08] bg-transparent px-3 py-1.5 text-sm text-black outline-none focus:border-blue-500 dark:border-white/[.145] dark:text-zinc-50"
                  />
                  <button
                    type="submit"
                    className="rounded-lg bg-foreground px-3 py-1.5 text-sm font-medium text-background"
                  >
                    Simpan
                  </button>
                </form>
              )}

              {kelas.groups.length === 0 ? (
                <p className="pl-4 text-sm text-zinc-400 dark:text-zinc-600">
                  Belum ada group di kelas ini.
                </p>
              ) : (
                <div className="flex flex-col gap-3 border-l-2 border-black/[.06] pl-4 dark:border-white/[.08]">
                  {kelas.groups.map((group) => {
                    const isChallengeFormOpen =
                      challengeFormFor?.kelasId === kelas.id &&
                      challengeFormFor.groupId === group.id;
                    return (
                      <div key={group.id} className="flex flex-col gap-2">
                        <div className="flex items-center justify-between">
                          <span className="text-sm font-semibold text-zinc-700 dark:text-zinc-300">
                            {group.name}
                          </span>
                          <button
                            type="button"
                            onClick={() =>
                              setChallengeFormFor(
                                isChallengeFormOpen
                                  ? null
                                  : { kelasId: kelas.id, groupId: group.id },
                              )
                            }
                            className="text-xs font-medium text-blue-600 dark:text-blue-400"
                          >
                            + Tambah Challenge
                          </button>
                        </div>

                        {isChallengeFormOpen && (
                          <form
                            onSubmit={(e) => submitChallenge(e, kelas.id, group.id)}
                            className="flex flex-col gap-2 sm:flex-row sm:items-center"
                          >
                            <input
                              autoFocus
                              value={newChallengeName}
                              onChange={(e) => setNewChallengeName(e.target.value)}
                              placeholder="mis. Latihan Bangun Ruang"
                              className="flex-1 rounded-lg border border-black/[.08] bg-transparent px-3 py-1.5 text-sm text-black outline-none focus:border-blue-500 dark:border-white/[.145] dark:text-zinc-50"
                            />
                            <label className="flex items-center gap-1.5 text-xs text-zinc-600 dark:text-zinc-400">
                              <input
                                type="checkbox"
                                checked={newChallengeIsExam}
                                onChange={(e) => setNewChallengeIsExam(e.target.checked)}
                              />
                              Jadikan ujian akhir group ini
                            </label>
                            <button
                              type="submit"
                              className="rounded-lg bg-foreground px-3 py-1.5 text-sm font-medium text-background"
                            >
                              Simpan
                            </button>
                          </form>
                        )}

                        {group.challenges.length === 0 ? (
                          <p className="pl-3 text-xs text-zinc-400 dark:text-zinc-600">
                            Belum ada challenge.
                          </p>
                        ) : (
                          <ul className="flex flex-col gap-1 pl-3">
                            {group.challenges.map((challenge) => (
                              <li
                                key={challenge.id}
                                className="flex items-center gap-2 text-sm text-zinc-600 dark:text-zinc-400"
                              >
                                <span>{challenge.isExam ? "🏁" : "•"}</span>
                                <span>{challenge.name}</span>
                                {challenge.isExam && (
                                  <span className="rounded-full bg-amber-100 px-2 py-0.5 text-[10px] font-semibold text-amber-700 dark:bg-amber-500/10 dark:text-amber-400">
                                    UJIAN
                                  </span>
                                )}
                              </li>
                            ))}
                          </ul>
                        )}
                      </div>
                    );
                  })}
                </div>
              )}
            </div>
          ))}
        </div>

        {showKelasForm ? (
          <form
            onSubmit={submitKelas}
            className="flex items-center gap-2 rounded-2xl border border-black/[.08] bg-white p-4 dark:border-white/[.145] dark:bg-zinc-900"
          >
            <input
              autoFocus
              value={newKelasName}
              onChange={(e) => setNewKelasName(e.target.value)}
              placeholder="mis. Kelas 10 B"
              className="flex-1 rounded-lg border border-black/[.08] bg-transparent px-3 py-1.5 text-sm text-black outline-none focus:border-blue-500 dark:border-white/[.145] dark:text-zinc-50"
            />
            <button
              type="submit"
              className="rounded-lg bg-foreground px-4 py-1.5 text-sm font-medium text-background"
            >
              Simpan
            </button>
            <button
              type="button"
              onClick={() => setShowKelasForm(false)}
              className="text-sm text-zinc-500 dark:text-zinc-500"
            >
              Batal
            </button>
          </form>
        ) : (
          <button
            type="button"
            onClick={() => setShowKelasForm(true)}
            className="self-start rounded-full bg-foreground px-5 py-2 text-sm font-medium text-background transition-colors hover:bg-[#383838] dark:hover:bg-[#ccc]"
          >
            + Tambah Kelas
          </button>
        )}
      </main>
    </div>
  );
}
