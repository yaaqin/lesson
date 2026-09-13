"use client";

import Link from "next/link";
import { useState, type FormEvent } from "react";
import { submitRegistration } from "@/store/org-registrations-store";

export default function RegisterOrganisasiPage() {
  const [organizationName, setOrganizationName] = useState("");
  const [ownerName, setOwnerName] = useState("");
  const [ownerEmail, setOwnerEmail] = useState("");
  const [password, setPassword] = useState("");
  const [submitted, setSubmitted] = useState(false);

  const handleSubmit = (event: FormEvent) => {
    event.preventDefault();
    submitRegistration({ organizationName, ownerName, ownerEmail, password });
    setSubmitted(true);
  };

  if (submitted) {
    return (
      <div className="flex min-h-screen flex-col items-center justify-center gap-6 bg-zinc-50 px-6 py-12 text-center font-sans dark:bg-black">
        <span className="text-4xl">📨</span>
        <div className="flex max-w-sm flex-col gap-2">
          <h1 className="text-xl font-semibold text-black dark:text-zinc-50">
            Pendaftaran terkirim
          </h1>
          <p className="text-sm text-zinc-600 dark:text-zinc-400">
            Organisasi <strong>{organizationName}</strong> sedang menunggu persetujuan admin
            platform. Kamu bisa login pakai <strong>{ownerEmail}</strong> setelah disetujui.
          </p>
        </div>
        <Link
          href="/login"
          className="rounded-full bg-foreground px-6 py-2.5 text-sm font-medium text-background transition-colors hover:bg-[#383838] dark:hover:bg-[#ccc]"
        >
          Kembali ke halaman login
        </Link>
      </div>
    );
  }

  return (
    <div className="flex min-h-screen flex-col items-center justify-center gap-8 bg-zinc-50 px-6 py-12 font-sans dark:bg-black">
      <Link href="/" className="text-lg font-semibold tracking-tight text-black dark:text-zinc-50">
        MathQuest
      </Link>

      <div className="flex w-full max-w-sm flex-col gap-6 rounded-2xl border border-black/[.08] bg-white p-6 dark:border-white/[.145] dark:bg-zinc-900">
        <div className="flex flex-col gap-1">
          <h1 className="text-xl font-semibold text-black dark:text-zinc-50">
            Daftarkan Organisasi
          </h1>
          <p className="text-sm text-zinc-500 dark:text-zinc-500">
            Pendaftaran perlu disetujui admin platform dulu sebelum kamu bisa login.
          </p>
        </div>

        <form onSubmit={handleSubmit} className="flex flex-col gap-4">
          <label className="flex flex-col gap-1.5 text-sm">
            <span className="font-medium text-zinc-700 dark:text-zinc-300">Nama Organisasi</span>
            <input
              type="text"
              required
              value={organizationName}
              onChange={(e) => setOrganizationName(e.target.value)}
              className="rounded-lg border border-black/[.08] bg-transparent px-3 py-2 text-black outline-none focus:border-blue-500 dark:border-white/[.145] dark:text-zinc-50"
              placeholder="SD Harapan Bangsa"
            />
          </label>
          <label className="flex flex-col gap-1.5 text-sm">
            <span className="font-medium text-zinc-700 dark:text-zinc-300">Nama Pemilik</span>
            <input
              type="text"
              required
              value={ownerName}
              onChange={(e) => setOwnerName(e.target.value)}
              className="rounded-lg border border-black/[.08] bg-transparent px-3 py-2 text-black outline-none focus:border-blue-500 dark:border-white/[.145] dark:text-zinc-50"
              placeholder="Nama kamu"
            />
          </label>
          <label className="flex flex-col gap-1.5 text-sm">
            <span className="font-medium text-zinc-700 dark:text-zinc-300">Email Pemilik</span>
            <input
              type="email"
              required
              value={ownerEmail}
              onChange={(e) => setOwnerEmail(e.target.value)}
              className="rounded-lg border border-black/[.08] bg-transparent px-3 py-2 text-black outline-none focus:border-blue-500 dark:border-white/[.145] dark:text-zinc-50"
              placeholder="kamu@sekolah.sch.id"
            />
          </label>
          <label className="flex flex-col gap-1.5 text-sm">
            <span className="font-medium text-zinc-700 dark:text-zinc-300">Password</span>
            <input
              type="password"
              required
              minLength={6}
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              className="rounded-lg border border-black/[.08] bg-transparent px-3 py-2 text-black outline-none focus:border-blue-500 dark:border-white/[.145] dark:text-zinc-50"
              placeholder="••••••••"
            />
          </label>

          <button
            type="submit"
            className="rounded-full bg-foreground px-6 py-2.5 text-sm font-medium text-background transition-colors hover:bg-[#383838] dark:hover:bg-[#ccc]"
          >
            Kirim Pendaftaran
          </button>
        </form>

        <p className="text-center text-sm text-zinc-500 dark:text-zinc-500">
          Sudah punya akun?{" "}
          <Link href="/login" className="font-medium text-blue-600 dark:text-blue-400">
            Masuk
          </Link>
        </p>
      </div>
    </div>
  );
}
