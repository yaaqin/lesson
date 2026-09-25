import Link from "next/link";

// Isi halaman undangan publik (tanpa login) dari link share: kartu share +
// ajakan gabung.
export function ShareInvite({
  headline,
  text,
  imageSrc,
  imageAlt,
}: {
  headline: string;
  text: string;
  imageSrc: string;
  imageAlt: string;
}) {
  return (
    <div className="flex min-h-screen flex-col items-center justify-center gap-8 bg-zinc-50 px-4 py-10 font-sans dark:bg-black">
      <div className="flex flex-col items-center gap-2 text-center">
        <span className="text-sm font-semibold tracking-wide text-blue-600 uppercase dark:text-blue-400">
          Undangan MathQuest
        </span>
        <h1 className="max-w-xl text-2xl font-semibold tracking-tight text-black sm:text-3xl dark:text-zinc-50">
          {headline}
        </h1>
        <p className="max-w-md text-sm text-zinc-600 dark:text-zinc-400">{text}</p>
      </div>

      {/* eslint-disable-next-line @next/next/no-img-element -- PNG dinamis dari route /image */}
      <img
        src={imageSrc}
        alt={imageAlt}
        width={1200}
        height={630}
        className="h-auto w-full max-w-2xl rounded-2xl border border-black/[.08] shadow-lg dark:border-white/[.145]"
      />

      <div className="flex w-full max-w-sm flex-col gap-2">
        <Link
          href="/login"
          className="rounded-full bg-foreground px-6 py-3 text-center text-sm font-medium text-background transition-colors hover:bg-[#383838] dark:hover:bg-[#ccc]"
        >
          Gabung gratis &amp; mulai latihan
        </Link>
        <p className="text-center text-xs text-zinc-500">
          Latihan matematika jadi game: SD, SMP, SMK/SMA sampai Kampus, lengkap sama streak harian, ranking, dan
          multiplayer.
        </p>
      </div>
    </div>
  );
}
