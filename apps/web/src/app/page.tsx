import Link from "next/link";

const tiers = [
  {
    step: "1",
    name: "SD",
    range: "Kelas 1–6",
    desc: "Penjumlahan, pengurangan, perkalian, pembagian dasar.",
  },
  {
    step: "2",
    name: "SMP",
    range: "Kelas 7–9",
    desc: "Aljabar dasar, pecahan, dan geometri dasar.",
  },
  {
    step: "3",
    name: "SMK / SMA",
    range: "Kelas 10–12",
    desc: "Materi lanjutan sesuai kurikulum SMK/SMA.",
  },
  {
    step: "4",
    name: "Kuliah",
    range: "Mahasiswa",
    desc: "Aljabar lanjut dan kalkulus.",
  },
  {
    step: "5",
    name: "Umum",
    range: "Bebas jenjang",
    desc: "Topik non-formal buat hobi, mis. crypto arithmetic.",
  },
];

const features = [
  {
    icon: "❤️",
    title: "Sistem Nyawa",
    desc: "Tiap hari dapet 3 nyawa. Salah dan gagal di satu sesi ngurangin nyawa — jadi tetap harus fokus, bukan asal coba-coba.",
  },
  {
    icon: "⏱️",
    title: "Waktu Jawab",
    desc: "Tiap soal punya batas waktu. Melatih bukan cuma bisa jawab, tapi juga jawab dengan cepat.",
  },
  {
    icon: "🔥",
    title: "Streak Harian",
    desc: "Selesaikan minimal satu tantangan tiap hari buat jaga streak kamu tetap nyala.",
  },
  {
    icon: "🔓",
    title: "Naik Level Berjenjang",
    desc: "Lulus ujian di akhir tiap batch buat buka batch berikutnya — progresmu selalu tersimpan.",
  },
];

export default function Home() {
  return (
    <div className="flex min-h-screen flex-col bg-zinc-50 font-sans dark:bg-black">
      <header className="flex items-center justify-between px-6 py-5 sm:px-10">
        <span className="text-lg font-semibold tracking-tight text-black dark:text-zinc-50">
          MathQuest
        </span>
        <button
          type="button"
          disabled
          className="cursor-not-allowed rounded-full border border-black/[.08] px-4 py-1.5 text-sm font-medium text-zinc-400 dark:border-white/[.145] dark:text-zinc-600"
        >
          Masuk
        </button>
      </header>

      <main className="flex flex-1 flex-col gap-24 px-6 pb-24 sm:px-10">
        {/* Hero */}
        <section className="flex flex-col items-center gap-8 pt-12 text-center sm:pt-20">
          <h1 className="max-w-2xl text-4xl font-semibold tracking-tight text-balance text-black sm:text-5xl dark:text-zinc-50">
            Latihan matematika, dibikin{" "}
            <span className="text-blue-600 dark:text-blue-400">se-adiktif</span> game.
          </h1>
          <p className="max-w-lg text-lg text-zinc-600 dark:text-zinc-400">
            Dari penjumlahan dasar SD sampai topik matematika umum — kerjain
            tantangan, jaga nyawa, dan kejar streak harianmu.
          </p>

          {/* Subject selector */}
          <div className="grid w-full max-w-md grid-cols-2 gap-3">
            <Link
              href="/belajar"
              className="flex flex-col items-center gap-1.5 rounded-2xl border-2 border-blue-600 bg-white px-4 py-5 shadow-sm transition-colors dark:border-blue-500 dark:bg-zinc-900"
            >
              <span className="text-2xl">➗</span>
              <span className="font-semibold text-black dark:text-zinc-50">
                Matematika
              </span>
              <span className="text-xs text-blue-600 dark:text-blue-400">
                Pilih pelajaran ini
              </span>
            </Link>
            <div className="relative flex flex-col items-center gap-1.5 rounded-2xl border-2 border-dashed border-black/[.08] bg-white/50 px-4 py-5 text-zinc-400 dark:border-white/[.145] dark:bg-zinc-900/40 dark:text-zinc-600">
              <span className="absolute -top-2.5 right-3 rounded-full bg-zinc-200 px-2 py-0.5 text-[10px] font-semibold tracking-wide text-zinc-600 uppercase dark:bg-zinc-700 dark:text-zinc-300">
                Soon
              </span>
              <span className="text-2xl grayscale">🧪</span>
              <span className="font-semibold">Sains</span>
              <span className="text-xs">Segera hadir</span>
            </div>
          </div>

          <a
            href="#jenjang"
            className="rounded-full bg-foreground px-6 py-3 text-sm font-medium text-background transition-colors hover:bg-[#383838] dark:hover:bg-[#ccc]"
          >
            Lihat Jenjang Belajar
          </a>
        </section>

        {/* Jenjang */}
        <section id="jenjang" className="flex flex-col gap-8 scroll-mt-10">
          <div className="flex flex-col gap-2 text-center">
            <h2 className="text-2xl font-semibold tracking-tight text-black dark:text-zinc-50">
              Satu jalur, dari dasar sampai umum
            </h2>
            <p className="text-zinc-600 dark:text-zinc-400">
              Tiap jenjang tersusun jadi batch-batch kecil, lulus ujian batch
              buat lanjut ke batch berikutnya.
            </p>
          </div>

          <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-5">
            {tiers.map((tier) => (
              <div
                key={tier.name}
                className="flex flex-col gap-2 rounded-2xl border border-black/[.08] bg-white p-5 dark:border-white/[.145] dark:bg-zinc-900"
              >
                <span className="flex h-7 w-7 items-center justify-center rounded-full bg-blue-600 text-xs font-semibold text-white dark:bg-blue-500">
                  {tier.step}
                </span>
                <span className="font-semibold text-black dark:text-zinc-50">
                  {tier.name}
                </span>
                <span className="text-xs text-zinc-500 dark:text-zinc-500">
                  {tier.range}
                </span>
                <p className="text-sm text-zinc-600 dark:text-zinc-400">
                  {tier.desc}
                </p>
                {tier.step === "1" && (
                  <Link
                    href="/belajar"
                    className="mt-1 text-sm font-medium text-blue-600 dark:text-blue-400"
                  >
                    Coba sekarang →
                  </Link>
                )}
              </div>
            ))}
          </div>
        </section>

        {/* Cara main */}
        <section className="flex flex-col gap-8">
          <div className="flex flex-col gap-2 text-center">
            <h2 className="text-2xl font-semibold tracking-tight text-black dark:text-zinc-50">
              Bukan cuma latihan soal biasa
            </h2>
            <p className="text-zinc-600 dark:text-zinc-400">
              Ada mekanik game di dalamnya, biar latihan matematika terasa
              seru dan bikin ketagihan (yang positif).
            </p>
          </div>

          <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-4">
            {features.map((feature) => (
              <div
                key={feature.title}
                className="flex flex-col gap-2 rounded-2xl border border-black/[.08] bg-white p-5 dark:border-white/[.145] dark:bg-zinc-900"
              >
                <span className="text-2xl">{feature.icon}</span>
                <span className="font-semibold text-black dark:text-zinc-50">
                  {feature.title}
                </span>
                <p className="text-sm text-zinc-600 dark:text-zinc-400">
                  {feature.desc}
                </p>
              </div>
            ))}
          </div>
        </section>
      </main>

      <footer className="px-6 py-8 text-center text-xs text-zinc-500 sm:px-10 dark:text-zinc-600">
        MathQuest — dibangun buat belajar matematika sambil main.
      </footer>
    </div>
  );
}
