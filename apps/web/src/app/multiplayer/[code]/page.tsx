"use client";

import Link from "next/link";
import { useParams, useRouter } from "next/navigation";
import { useEffect, useState } from "react";
import { useAuthStore } from "@/store/auth-store";
import { useRequireNickname } from "@/hooks/use-profile";
import { useCountdown, useRoomSocket, type ClientMessage, type RoomState } from "@/hooks/use-multiplayer";
import { NumericKeypad } from "@/components/numeric-keypad";
import { UserAvatar } from "@/components/user-avatar";
import { PremiumBadge } from "@/components/premium-badge";
import { FORMAT_LABEL, JOIN_ERROR_TEXT, MODE_INFO, TIER_BADGE, TIER_LABEL, normalizeRoomCode } from "@/lib/multiplayer";
import { clearPostLoginPath, rememberPostLoginPath } from "@/lib/post-login";

const ACTION_ERROR_TEXT: Record<string, string> = {
  not_enough_players: "Butuh minimal 2 pemain buat mulai.",
  room_full: "Room udah penuh.",
  game_started: "Game udah mulai.",
};

const ENDED_TEXT: Record<string, string> = {
  ...JOIN_ERROR_TEXT,
  closed: "Room udah ditutup.",
  replaced: "Kamu buka room ini di tab/device lain.",
};

export default function RoomPage() {
  const router = useRouter();
  const params = useParams<{ code: string }>();
  const code = normalizeRoomCode(params.code);
  const session = useAuthStore((s) => s.session);
  const hasHydrated = useAuthStore((s) => s.hasHydrated);
  const me = useRequireNickname().data;

  useEffect(() => {
    if (!hasHydrated) return;
    if (session === null) {
      rememberPostLoginPath(`/multiplayer/${code}`);
      router.replace("/login");
    }
  }, [hasHydrated, session, router, code]);

  if (!hasHydrated || !session || !me?.nickname) {
    return <Centered text="Memuat…" />;
  }
  return <Room code={code} />;
}

function Room({ code }: { code: string }) {
  const router = useRouter();
  const { state, status, lastError, clockOffset, send, leave } = useRoomSocket(code);

  useEffect(() => clearPostLoginPath(), []);
  // Error aksi (mis. mulai kurang pemain) tampil 3 detik lalu ilang sendiri.
  const [dismissedAt, setDismissedAt] = useState(0);
  const notice = lastError && lastError.at !== dismissedAt ? ACTION_ERROR_TEXT[lastError.code] : undefined;

  useEffect(() => {
    if (!lastError) return;
    const id = window.setTimeout(() => setDismissedAt(lastError.at), 3000);
    return () => window.clearTimeout(id);
  }, [lastError]);

  if (status.kind === "ended") {
    if (status.reason === "left") return <Centered text="Keluar dari room…" redirect="/multiplayer" />;
    return (
      <Centered text={ENDED_TEXT[status.reason] ?? "Koneksi ke room berakhir."}>
        <Link href="/multiplayer" className="rounded-full bg-foreground px-6 py-3 text-sm font-semibold text-background">
          Kembali
        </Link>
      </Centered>
    );
  }
  if (!state) return <Centered text="Nyambung ke room…" />;

  const exit = () => {
    leave();
    router.push("/multiplayer");
  };

  return (
    <div className="flex min-h-screen flex-col bg-zinc-50 font-sans dark:bg-black">
      <header className="flex items-center justify-between gap-2 px-4 py-4 sm:px-10">
        <span className="font-mono text-sm font-semibold tracking-[0.2em] text-zinc-500">{state.room.code}</span>
        <span className="text-sm font-semibold text-black dark:text-zinc-50">
          {MODE_INFO[state.room.mode].icon} {MODE_INFO[state.room.mode].label}
        </span>
        <ExitButton state={state} onExit={exit} />
      </header>

      {status.kind === "reconnecting" && (
        <div className="bg-amber-100 px-4 py-2 text-center text-xs font-medium text-amber-800 dark:bg-amber-500/10 dark:text-amber-300">
          Koneksi putus, nyambung lagi…
        </div>
      )}
      {notice && (
        <div className="bg-red-100 px-4 py-2 text-center text-xs font-medium text-red-700 dark:bg-red-500/10 dark:text-red-300">
          {notice}
        </div>
      )}

      <main className="mx-auto flex w-full max-w-2xl flex-1 flex-col gap-6 px-4 pb-10 sm:px-6">
        {state.phase === "lobby" && <Lobby state={state} send={send} />}
        {state.phase === "countdown" && <CountdownView state={state} clockOffset={clockOffset} />}
        {(state.phase === "question" || state.phase === "reveal") && (
          <QuestionView key={state.question?.index} state={state} clockOffset={clockOffset} send={send} />
        )}
        {state.phase === "finished" && <Results state={state} onExit={exit} />}
      </main>
    </div>
  );
}

function ExitButton({ state, onExit }: { state: RoomState; onExit: () => void }) {
  const [confirming, setConfirming] = useState(false);
  const inGame = state.phase !== "lobby" && state.phase !== "finished";
  const hostInLobby = state.you.isHost && state.phase === "lobby";

  if (!confirming) {
    return (
      <button
        type="button"
        onClick={() => (inGame || hostInLobby ? setConfirming(true) : onExit())}
        className="rounded-full border border-black/[.08] px-3 py-1 text-xs font-medium text-zinc-600 dark:border-white/[.145] dark:text-zinc-400"
      >
        {hostInLobby ? "Tutup room" : "Keluar"}
      </button>
    );
  }
  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/40 px-6">
      <div className="flex w-full max-w-sm flex-col gap-4 rounded-2xl bg-white p-6 dark:bg-zinc-900">
        <p className="font-semibold text-black dark:text-zinc-50">{hostInLobby ? "Tutup room?" : "Keluar dari game?"}</p>
        <p className="text-sm text-zinc-500 dark:text-zinc-400">
          {hostInLobby
            ? "Semua pemain di lobby bakal dikeluarin."
            : "Kamu gak bisa jawab soal lagi, skor yang udah didapet tetap kehitung."}
        </p>
        <div className="grid grid-cols-2 gap-2">
          <button
            type="button"
            onClick={() => setConfirming(false)}
            className="rounded-full border border-black/[.08] py-2.5 text-sm font-medium dark:border-white/[.145] dark:text-zinc-300"
          >
            Batal
          </button>
          <button type="button" onClick={onExit} className="rounded-full bg-red-600 py-2.5 text-sm font-semibold text-white">
            {hostInLobby ? "Tutup" : "Keluar"}
          </button>
        </div>
      </div>
    </div>
  );
}

// ---------- Lobby ----------

function Lobby({ state, send }: { state: RoomState; send: (msg: ClientMessage) => boolean }) {
  const [copied, setCopied] = useState(false);
  const { room, you } = state;
  const canStart = state.playingCount >= room.minPlayers;

  const share = async () => {
    const url = `${window.location.origin}/multiplayer/${room.code}`;
    const text = `Ayo main MathQuest bareng! Kode room: ${room.code}`;
    if (navigator.share) {
      try {
        await navigator.share({ title: "MathQuest Multiplayer", text, url });
        return;
      } catch {
        // dibatalin user -> gak usah fallback
        return;
      }
    }
    await navigator.clipboard?.writeText(`${text}\n${url}`);
    setCopied(true);
    window.setTimeout(() => setCopied(false), 1500);
  };

  return (
    <>
      <div className="flex flex-col items-center gap-3 rounded-2xl border border-black/[.08] bg-white p-6 text-center dark:border-white/[.145] dark:bg-zinc-900">
        <span className="text-xs font-semibold tracking-wide text-zinc-500 uppercase">Kode room</span>
        <span className="font-mono text-4xl font-bold tracking-[0.3em] text-black dark:text-zinc-50">{room.code}</span>
        <button
          type="button"
          onClick={share}
          className="rounded-full border border-violet-300 bg-violet-50 px-5 py-2 text-sm font-semibold text-violet-700 dark:border-violet-500/30 dark:bg-violet-500/10 dark:text-violet-300"
        >
          {copied ? "Link disalin ✓" : "Ajak temen 🔗"}
        </button>
      </div>

      <div className="grid grid-cols-2 gap-3 text-sm sm:grid-cols-4">
        <Info label="Soal" value={`${room.questionCount}`} />
        <Info label="Bentuk" value={FORMAT_LABEL[room.format]} />
        <Info label="Waktu" value={`${room.secondsPerQuestion} dtk/soal`} />
        <Info label="Pemain" value={`${state.playingCount}/${room.maxPlayers}`} />
      </div>

      <div className="flex flex-wrap gap-1.5">
        {room.sources.map((s) => (
          <span key={s} className="rounded-full bg-zinc-100 px-3 py-1 text-xs text-zinc-600 dark:bg-zinc-800 dark:text-zinc-400">
            {s}
          </span>
        ))}
      </div>

      <div className="flex flex-col gap-2">
        {state.players.map((p) => (
          <div
            key={p.userId}
            className={`flex items-center gap-3 rounded-2xl border border-black/[.08] bg-white px-4 py-3 dark:border-white/[.145] dark:bg-zinc-900 ${
              p.connected ? "" : "opacity-50"
            }`}
          >
            <UserAvatar avatar={p.avatar} size="sm" />
            <span className="flex min-w-0 flex-1 flex-wrap items-center gap-1.5">
              <span className="truncate text-sm font-medium text-black dark:text-zinc-50">@{p.nickname}</span>
              {p.isPremium && <PremiumBadge size="xs" />}
              {p.isHost && <span className="text-xs text-zinc-500">· host{p.playing ? "" : " (nonton)"}</span>}
              {p.userId === you.userId && <span className="text-xs text-zinc-400">(kamu)</span>}
              {!p.connected && <span className="text-xs text-zinc-400">· offline</span>}
            </span>
            {you.isHost && !p.isHost && (
              <button
                type="button"
                onClick={() => send({ type: "kick", userId: p.userId })}
                className="text-xs font-medium text-red-500"
              >
                Keluarin
              </button>
            )}
          </div>
        ))}
      </div>

      {you.isHost ? (
        <div className="flex flex-col gap-3">
          <div className="grid grid-cols-2 gap-2">
            {[true, false].map((plays) => (
              <button
                key={String(plays)}
                type="button"
                onClick={() => send({ type: "set_host_plays", plays })}
                className={`rounded-full border-2 py-2 text-sm font-medium ${
                  you.playing === plays
                    ? "border-violet-500 bg-violet-50 text-violet-700 dark:bg-violet-500/10 dark:text-violet-300"
                    : "border-black/[.08] text-zinc-600 dark:border-white/[.145] dark:text-zinc-400"
                }`}
              >
                {plays ? "🎮 Ikut main" : "👀 Cuma nonton"}
              </button>
            ))}
          </div>
          <button
            type="button"
            disabled={!canStart}
            onClick={() => send({ type: "start" })}
            className="rounded-full bg-violet-600 py-3 text-sm font-semibold text-white disabled:opacity-40"
          >
            {canStart ? "Mulai game" : `Nunggu pemain (min. ${room.minPlayers})…`}
          </button>
        </div>
      ) : (
        <p className="text-center text-sm text-zinc-500 dark:text-zinc-400">Nunggu host mulai game…</p>
      )}
    </>
  );
}

function Info({ label, value }: { label: string; value: string }) {
  return (
    <div className="flex flex-col gap-0.5 rounded-xl border border-black/[.08] bg-white px-3 py-2 dark:border-white/[.145] dark:bg-zinc-900">
      <span className="text-xs text-zinc-500">{label}</span>
      <span className="font-semibold text-black dark:text-zinc-50">{value}</span>
    </div>
  );
}

// ---------- Main ----------

function CountdownView({ state, clockOffset }: { state: RoomState; clockOffset: number }) {
  const left = useCountdown(state.phaseEndsAt, clockOffset);
  return (
    <div className="flex flex-1 flex-col items-center justify-center gap-4 py-20">
      <span className="text-sm text-zinc-500">Siap-siap…</span>
      <span className="text-8xl font-bold text-violet-600 dark:text-violet-400">{Math.max(left, 1)}</span>
    </div>
  );
}

function QuestionView({
  state,
  clockOffset,
  send,
}: {
  state: RoomState;
  clockOffset: number;
  send: (msg: ClientMessage) => boolean;
}) {
  const q = state.question!;
  const left = useCountdown(state.phaseEndsAt, clockOffset);
  const [essayValue, setEssayValue] = useState("");
  // pendingValue: jawaban yang udah dikirim tapi snapshot dari server belum
  // nyampe -- biar tombol langsung kekunci & gak bisa dobel kirim.
  const [pendingValue, setPendingValue] = useState<number | null>(null);

  const isReveal = state.phase === "reveal";
  const spectator = !state.you.playing;
  const answered = state.myAnswer?.value ?? pendingValue;
  const locked = spectator || isReveal || answered !== null;
  const myCorrect = state.myAnswer?.correct;

  const answer = (value: number) => {
    if (locked) return;
    if (send({ type: "answer", index: q.index, value })) setPendingValue(value);
  };

  const submitEssay = () => {
    if (essayValue === "" || essayValue === "-") return;
    const num = Number(essayValue);
    if (Number.isFinite(num)) answer(num);
  };

  const optionClass = (value: number) => {
    const picked = answered === value;
    if (isReveal && state.reveal && value === state.reveal.correctValue)
      return "border-green-500 bg-green-50 text-green-700 dark:bg-green-500/10 dark:text-green-400";
    if (picked && myCorrect === false) return "border-red-500 bg-red-50 text-red-700 dark:bg-red-500/10 dark:text-red-400";
    if (picked) return "border-violet-500 bg-violet-50 text-violet-700 dark:bg-violet-500/10 dark:text-violet-300";
    return "border-black/[.08] bg-white text-black hover:border-violet-400 dark:border-white/[.145] dark:bg-zinc-900 dark:text-zinc-50";
  };

  return (
    <>
      <div className="flex flex-col gap-2">
        <div className="flex items-center justify-between text-sm text-zinc-500">
          <span>
            Soal {q.index + 1} / {q.total}
          </span>
          {!isReveal && (
            <span className={`font-semibold ${left <= 5 ? "text-red-500" : "text-zinc-500"}`}>⏱️ {left}s</span>
          )}
        </div>
        <div className="h-2 w-full overflow-hidden rounded-full bg-black/[.06] dark:bg-white/[.08]">
          <div className="h-full rounded-full bg-violet-600 transition-all" style={{ width: `${((q.index + 1) / q.total) * 100}%` }} />
        </div>
      </div>

      <div className="flex flex-col items-center gap-3 py-2">
        <div className="flex items-center gap-2">
          <span className={`rounded-full px-3 py-1 text-xs font-semibold ${TIER_BADGE[q.tierCode]}`}>{TIER_LABEL[q.tierCode]}</span>
          {q.type === "essay_numeric" && (
            <span className="rounded-full bg-purple-100 px-3 py-1 text-xs font-semibold text-purple-700 dark:bg-purple-500/10 dark:text-purple-400">
              ✏️ Isian
            </span>
          )}
        </div>
        <h2 className="text-center text-2xl font-semibold tracking-tight break-words text-black sm:text-4xl dark:text-zinc-50">
          {q.prompt}
        </h2>
      </div>

      {q.type === "essay_numeric" ? (
        <div className="flex flex-col items-center gap-2">
          {isReveal && state.reveal && (
            <span className="text-sm font-semibold text-green-600 dark:text-green-400">Jawaban: {state.reveal.correctValue}</span>
          )}
          {!spectator && (
            <NumericKeypad
              value={answered !== null ? String(answered) : essayValue}
              onChange={setEssayValue}
              onSubmit={submitEssay}
              disabled={locked}
              feedback={myCorrect === true ? "correct" : myCorrect === false ? "incorrect" : null}
            />
          )}
        </div>
      ) : (
        <div className="grid grid-cols-2 gap-3">
          {(q.options ?? []).map((value) => (
            <button
              key={value}
              type="button"
              disabled={locked}
              onClick={() => answer(value)}
              className={`rounded-2xl border-2 px-4 py-5 text-xl font-semibold break-all transition-colors disabled:cursor-default ${optionClass(value)}`}
            >
              {value}
            </button>
          ))}
        </div>
      )}

      <StatusLine state={state} answered={answered !== null} />
    </>
  );
}

function StatusLine({ state, answered }: { state: RoomState; answered: boolean }) {
  const isRace = state.room.mode === "race";
  const spectator = !state.you.playing;
  const reveal = state.reveal;
  const base = "text-center text-sm font-medium";

  if (state.phase === "reveal" && reveal) {
    if (isRace) {
      if (!reveal.winner) return <p className={`${base} text-zinc-500`}>Gak ada yang benar di soal ini.</p>;
      const mine = reveal.winner.userId === state.you.userId;
      return (
        <p className={`${base} ${mine ? "text-green-600 dark:text-green-400" : "text-zinc-600 dark:text-zinc-300"}`}>
          {mine ? "⚡ Kamu paling cepet! +1 poin" : `⚡ @${reveal.winner.nickname} paling cepet`}
        </p>
      );
    }
    const summary = `${reveal.correctCount} dari ${state.playingCount} pemain benar`;
    if (spectator) return <p className={`${base} text-zinc-500`}>{summary}</p>;
    const my = state.myAnswer;
    return (
      <div className="flex flex-col items-center gap-1">
        {!my ? (
          <p className={`${base} text-red-500`}>Waktu habis!</p>
        ) : my.correct ? (
          <p className={`${base} text-green-600 dark:text-green-400`}>Benar! +{my.points} poin</p>
        ) : (
          <p className={`${base} text-red-500`}>Salah</p>
        )}
        <p className="text-xs text-zinc-500">{summary}</p>
      </div>
    );
  }

  const progress = `${state.answeredCount}/${state.playingCount} udah jawab`;
  if (spectator) return <p className={`${base} text-zinc-500`}>👀 Kamu nonton · {progress}</p>;
  if (isRace && state.myAnswer?.correct === false)
    return <p className={`${base} text-red-500`}>Salah! Kamu kekunci di soal ini, tunggu yang lain…</p>;
  if (answered)
    return <p className={`${base} text-zinc-500`}>{isRace ? "Ngecek jawaban…" : `Jawaban terkirim, tunggu waktu habis · ${progress}`}</p>;
  return <p className={`${base} text-zinc-400`}>{progress}</p>;
}

// ---------- Hasil ----------

function Results({ state, onExit }: { state: RoomState; onExit: () => void }) {
  const results = state.results ?? [];
  const isRace = state.room.mode === "race";
  const medal = (rank: number) => (rank === 1 ? "🥇" : rank === 2 ? "🥈" : rank === 3 ? "🥉" : `#${rank}`);

  return (
    <>
      <div className="flex flex-col items-center gap-1 pt-4 text-center">
        <span className="text-4xl">🏆</span>
        <h1 className="text-2xl font-semibold tracking-tight text-black dark:text-zinc-50">Hasil Akhir</h1>
        <p className="text-sm text-zinc-500">
          {state.room.questionCount} soal · {MODE_INFO[state.room.mode].label}
        </p>
      </div>

      <div className="flex flex-col gap-2">
        {results.map((r) => {
          const mine = r.player.userId === state.you.userId;
          return (
            <div
              key={r.player.userId}
              className={`flex items-center gap-3 rounded-2xl border-2 px-4 py-3 ${
                mine
                  ? "border-violet-500 bg-violet-50 dark:bg-violet-500/10"
                  : "border-black/[.08] bg-white dark:border-white/[.145] dark:bg-zinc-900"
              }`}
            >
              <span className="w-8 text-center text-lg font-semibold">{medal(r.rank)}</span>
              <UserAvatar avatar={r.player.avatar} size="sm" />
              <span className="flex min-w-0 flex-1 flex-col">
                <span className="flex flex-wrap items-center gap-1.5">
                  <span className="truncate text-sm font-semibold text-black dark:text-zinc-50">@{r.player.nickname}</span>
                  {r.player.isPremium && <PremiumBadge size="xs" />}
                  {r.left && <span className="text-xs text-zinc-400">(keluar)</span>}
                </span>
                <span className="text-xs text-zinc-500">
                  {r.correctCount} benar
                  {r.correctCount > 0 && ` · rata-rata ${(r.avgCorrectMs / 1000).toFixed(1)} dtk`}
                </span>
              </span>
              <span className="text-right">
                <span className="block text-lg font-bold text-black dark:text-zinc-50">{r.score.toLocaleString("id-ID")}</span>
                <span className="text-xs text-zinc-500">poin</span>
              </span>
            </div>
          );
        })}
      </div>

      {isRace && <p className="text-center text-xs text-zinc-400">Adu cepat: 1 poin tiap soal yang kamu jawab benar paling duluan.</p>}

      <div className="flex flex-col gap-2">
        {state.you.isHost && (
          <Link href="/multiplayer/buat" className="rounded-full bg-violet-600 py-3 text-center text-sm font-semibold text-white">
            Bikin room baru
          </Link>
        )}
        <button
          type="button"
          onClick={onExit}
          className="rounded-full border border-black/[.08] py-3 text-sm font-medium text-zinc-700 dark:border-white/[.145] dark:text-zinc-300"
        >
          Selesai
        </button>
      </div>
    </>
  );
}

function Centered({ text, redirect, children }: { text: string; redirect?: string; children?: React.ReactNode }) {
  const router = useRouter();
  useEffect(() => {
    if (redirect) router.replace(redirect);
  }, [redirect, router]);
  return (
    <div className="flex min-h-screen flex-col items-center justify-center gap-4 bg-zinc-50 px-6 text-center dark:bg-black">
      <p className="text-zinc-500 dark:text-zinc-500">{text}</p>
      {children}
    </div>
  );
}
