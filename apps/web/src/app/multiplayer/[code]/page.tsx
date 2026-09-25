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
import { PathShareButton } from "@/components/share-button";
import { FORMAT_LABEL, JOIN_ERROR_TEXT, MODE_INFO, TIER_BADGE, TIER_LABEL, normalizeRoomCode } from "@/lib/multiplayer";
import { clearPostLoginPath, rememberPostLoginPath } from "@/lib/post-login";
import { buildMatchShareContent, matchSharePath } from "@/lib/share";

const ACTION_ERROR_TEXT: Record<string, string> = {
  not_enough_players: "Butuh minimal 2 pemain buat mulai.",
  room_full: "Room udah penuh.",
  game_started: "Game udah mulai.",
};

const ENDED_TEXT: Record<string, string> = {
  ...JOIN_ERROR_TEXT,
  closed: "Room udah ditutup.",
  host_closed: "Host udah nutup room ini.",
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
  const { state, status, lastError, clockOffset, send, leave, closeRoom } = useRoomSocket(code);

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
  // Kill room: host nutup room, semua pemain ikut dikeluarin.
  const kill = () => {
    closeRoom();
    router.push("/multiplayer");
  };

  return (
    <div className="flex min-h-screen flex-col bg-zinc-50 font-sans dark:bg-black">
      <header className="flex items-center justify-between gap-2 px-4 py-4 sm:px-10">
        <span className="font-mono text-sm font-semibold tracking-[0.2em] text-zinc-500">{state.room.code}</span>
        <span className="text-sm font-semibold text-black dark:text-zinc-50">
          {MODE_INFO[state.room.mode].icon} {MODE_INFO[state.room.mode].label}
        </span>
        <ExitButton state={state} onExit={exit} onKill={kill} />
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
        {state.phase === "countdown" && (
          <div className="flex flex-1 flex-col items-center justify-center py-20">
            <BigCountdown endsAt={state.phaseEndsAt} clockOffset={clockOffset} label="Siap-siap…" />
          </div>
        )}
        {state.phase === "question" && (
          <QuestionView key={state.question?.index} state={state} clockOffset={clockOffset} send={send} />
        )}
        {state.phase === "reveal" && <WinnerFlash state={state} />}
        {state.phase === "cooldown" && <CooldownView state={state} clockOffset={clockOffset} />}
        {state.phase === "awaiting_results" && <AwaitingResults state={state} send={send} />}
        {state.phase === "results_countdown" && (
          <div className="flex flex-1 flex-col items-center justify-center gap-4 py-20">
            <span className="text-5xl">🥁</span>
            <BigCountdown endsAt={state.phaseEndsAt} clockOffset={clockOffset} label="Hasil diumumin dalam" />
          </div>
        )}
        {state.phase === "finished" && <Results state={state} onExit={exit} />}
      </main>
    </div>
  );
}

function ExitButton({ state, onExit, onKill }: { state: RoomState; onExit: () => void; onKill: () => void }) {
  const [confirming, setConfirming] = useState(false);
  const inGame = state.phase !== "lobby" && state.phase !== "finished";
  const isHost = state.you.isHost;
  const hostInLobby = isHost && state.phase === "lobby";

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

  const cancel = (
    <button
      type="button"
      onClick={() => setConfirming(false)}
      className="rounded-full border border-black/[.08] py-2.5 text-sm font-medium dark:border-white/[.145] dark:text-zinc-300"
    >
      Batal
    </button>
  );

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/40 px-6">
      <div className="flex w-full max-w-sm flex-col gap-4 rounded-2xl bg-white p-6 dark:bg-zinc-900">
        {hostInLobby ? (
          <>
            <p className="font-semibold text-black dark:text-zinc-50">Tutup room?</p>
            <p className="text-sm text-zinc-500 dark:text-zinc-400">Semua pemain di lobby bakal dikeluarin.</p>
            <div className="grid grid-cols-2 gap-2">
              {cancel}
              <button type="button" onClick={onKill} className="rounded-full bg-red-600 py-2.5 text-sm font-semibold text-white">
                Tutup
              </button>
            </div>
          </>
        ) : isHost ? (
          <>
            <p className="font-semibold text-black dark:text-zinc-50">Keluar dari game?</p>
            <p className="text-sm text-zinc-500 dark:text-zinc-400">
              Kamu yang bikin room ini. Mau bubarin room buat semua pemain, atau keluar sendiri aja?
            </p>
            <div className="flex flex-col gap-2">
              <button type="button" onClick={onKill} className="rounded-full bg-red-600 py-2.5 text-sm font-semibold text-white">
                Tutup room (semua keluar)
              </button>
              <button
                type="button"
                onClick={onExit}
                className="rounded-full border border-red-300 py-2.5 text-sm font-medium text-red-600 dark:border-red-500/40 dark:text-red-400"
              >
                Keluar aja, game tetap lanjut
              </button>
              {cancel}
            </div>
          </>
        ) : (
          <>
            <p className="font-semibold text-black dark:text-zinc-50">Keluar dari game?</p>
            <p className="text-sm text-zinc-500 dark:text-zinc-400">
              Kamu gak bisa jawab soal lagi, skor yang udah didapet tetap kehitung.
            </p>
            <div className="grid grid-cols-2 gap-2">
              {cancel}
              <button type="button" onClick={onExit} className="rounded-full bg-red-600 py-2.5 text-sm font-semibold text-white">
                Keluar
              </button>
            </div>
          </>
        )}
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

// BigCountdown: angka hitung mundur gede (countdown awal, cooldown antar
// soal, countdown hasil) -- tiap angka ganti "nongol" biar kerasa.
function BigCountdown({ endsAt, clockOffset, label }: { endsAt: number; clockOffset: number; label: string }) {
  const left = Math.max(useCountdown(endsAt, clockOffset), 1);
  return (
    <div className="flex flex-col items-center gap-2">
      <span className="text-sm font-semibold text-zinc-500">{label}</span>
      <span key={left} className="animate-pop text-9xl leading-none font-bold text-violet-600 tabular-nums dark:text-violet-400">
        {left}
      </span>
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

  const spectator = !state.you.playing;
  const answered = state.myAnswer?.value ?? pendingValue;
  const locked = spectator || answered !== null;
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
    if (picked && myCorrect === false) return "border-red-500 bg-red-50 text-red-700 dark:bg-red-500/10 dark:text-red-400";
    if (picked) return "border-violet-500 bg-violet-50 text-violet-700 dark:bg-violet-500/10 dark:text-violet-300";
    return "border-black/[.08] bg-white text-black hover:border-violet-400 dark:border-white/[.145] dark:bg-zinc-900 dark:text-zinc-50";
  };

  return (
    <>
      <QuestionHeader state={state} right={<span className={`font-semibold ${left <= 5 ? "text-red-500" : "text-zinc-500"}`}>⏱️ {left}s</span>} />

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
        !spectator && (
          <div className="flex flex-col items-center gap-2">
            <NumericKeypad
              value={answered !== null ? String(answered) : essayValue}
              onChange={setEssayValue}
              onSubmit={submitEssay}
              disabled={locked}
              feedback={myCorrect === true ? "correct" : myCorrect === false ? "incorrect" : null}
            />
          </div>
        )
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

function QuestionHeader({ state, right }: { state: RoomState; right?: React.ReactNode }) {
  const q = state.question!;
  return (
    <div className="flex flex-col gap-2">
      <div className="flex items-center justify-between text-sm text-zinc-500">
        <span>
          Soal {q.index + 1} / {q.total}
        </span>
        {right}
      </div>
      <div className="h-2 w-full overflow-hidden rounded-full bg-black/[.06] dark:bg-white/[.08]">
        <div className="h-full rounded-full bg-violet-600 transition-all" style={{ width: `${((q.index + 1) / q.total) * 100}%` }} />
      </div>
    </div>
  );
}

function StatusLine({ state, answered }: { state: RoomState; answered: boolean }) {
  const isRace = state.room.mode === "race";
  const spectator = !state.you.playing;
  const base = "text-center text-sm font-medium";

  const progress = `${state.answeredCount}/${state.playingCount} udah jawab`;
  if (spectator) return <p className={`${base} text-zinc-500`}>👀 Kamu nonton · {progress}</p>;
  if (isRace && state.myAnswer?.correct === false)
    return <p className={`${base} text-red-500`}>Salah! Kamu kekunci di soal ini, tunggu yang lain…</p>;
  if (answered)
    return (
      <p className={`${base} text-zinc-500`}>
        {isRace ? "Ngecek jawaban…" : `Jawaban terkirim, nunggu yang lain · ${progress}`}
      </p>
    );
  return <p className={`${base} text-zinc-400`}>{progress}</p>;
}

// WinnerFlash: adu cepat + showFastest -- pamer yang paling cepet jawab
// benar sebentar sebelum cooldown.
function WinnerFlash({ state }: { state: RoomState }) {
  const winner = state.reveal?.winner;
  if (!winner) return null;
  const mine = winner.userId === state.you.userId;
  return (
    <div className="flex flex-1 flex-col items-center justify-center gap-4 py-16 text-center">
      <span className="animate-pop text-6xl">⚡</span>
      <UserAvatar avatar={winner.avatar} size="lg" />
      <span className="animate-pop text-3xl font-bold tracking-tight text-black dark:text-zinc-50">
        {mine ? "Kamu paling cepet!" : `@${winner.nickname}`}
      </span>
      <span className={`text-sm font-medium ${mine ? "text-green-600 dark:text-green-400" : "text-zinc-500"}`}>
        {mine ? "+1 poin" : "paling cepet jawab benar"}
      </span>
    </div>
  );
}

// CooldownView: jeda antar soal -- angka hitung mundur gede + kunci jawaban
// soal barusan + hasil kamu.
function CooldownView({ state, clockOffset }: { state: RoomState; clockOffset: number }) {
  const q = state.question!;
  const isLast = q.index + 1 >= q.total;
  return (
    <>
      <QuestionHeader state={state} />
      <div className="flex flex-1 flex-col items-center justify-center gap-8 py-6 text-center">
        <BigCountdown
          endsAt={state.phaseEndsAt}
          clockOffset={clockOffset}
          label={isLast ? "Soal selesai! Hasil dalam" : `Soal ${q.index + 2} dalam`}
        />
        <div className="flex w-full flex-col items-center gap-2 rounded-2xl border border-black/[.08] bg-white p-5 dark:border-white/[.145] dark:bg-zinc-900">
          <span className="text-sm break-words text-zinc-500">{q.prompt}</span>
          {state.reveal && (
            <span className="text-2xl font-bold text-green-600 dark:text-green-400">Jawaban: {state.reveal.correctValue}</span>
          )}
          <CooldownResult state={state} />
        </div>
      </div>
    </>
  );
}

function CooldownResult({ state }: { state: RoomState }) {
  const reveal = state.reveal;
  if (!reveal) return null;
  const spectator = !state.you.playing;
  const my = state.myAnswer;
  const base = "text-center text-sm font-medium";

  if (state.room.mode === "race") {
    if (my?.correct) return <p className={`${base} text-green-600 dark:text-green-400`}>⚡ Kamu paling cepet! +1 poin</p>;
    if (reveal.winner) return <p className={`${base} text-zinc-600 dark:text-zinc-300`}>⚡ @{reveal.winner.nickname} paling cepet</p>;
    if (reveal.hasWinner) return <p className={`${base} text-zinc-600 dark:text-zinc-300`}>🤫 Ada yang jawab benar duluan…</p>;
    return <p className={`${base} text-zinc-500`}>Gak ada yang benar di soal ini.</p>;
  }

  const summary = `${reveal.correctCount} dari ${state.playingCount} pemain benar`;
  if (spectator) return <p className={`${base} text-zinc-500`}>{summary}</p>;
  return (
    <div className="flex flex-col items-center gap-1">
      {!my ? (
        <p className={`${base} text-red-500`}>Waktu habis!</p>
      ) : my.correct ? (
        <p className={`${base} text-green-600 dark:text-green-400`}>Benar! +{my.points} poin</p>
      ) : (
        <p className={`${base} text-red-500`}>Salah (jawabanmu {my.value})</p>
      )}
      <p className="text-xs text-zinc-500">{summary}</p>
    </div>
  );
}

// AwaitingResults: adu cepat yang hasilnya dirahasiain -- nunggu host buka.
function AwaitingResults({ state, send }: { state: RoomState; send: (msg: ClientMessage) => boolean }) {
  return (
    <div className="flex flex-1 flex-col items-center justify-center gap-4 py-16 text-center">
      <span className="text-6xl">🏁</span>
      <h2 className="text-2xl font-semibold tracking-tight text-black dark:text-zinc-50">Semua soal selesai!</h2>
      {state.you.isHost ? (
        <>
          <p className="text-sm text-zinc-500 dark:text-zinc-400">Hasilnya masih rahasia. Buka kalau semua udah siap.</p>
          <button
            type="button"
            onClick={() => send({ type: "reveal_results" })}
            className="rounded-full bg-violet-600 px-8 py-3 text-sm font-semibold text-white transition-colors hover:bg-violet-700"
          >
            Tampilkan hasil 🏆
          </button>
        </>
      ) : (
        <p className="animate-pulse text-sm text-zinc-500 dark:text-zinc-400">🤫 Nunggu host nampilin hasil…</p>
      )}
    </div>
  );
}

// ---------- Hasil ----------

function Results({ state, onExit }: { state: RoomState; onExit: () => void }) {
  const results = state.results ?? [];
  const isRace = state.room.mode === "race";
  const me = results.find((r) => r.player.userId === state.you.userId);
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
        {me && state.matchId && (
          <PathShareButton
            path={matchSharePath(me.player.nickname, state.matchId)}
            text={
              buildMatchShareContent(
                state.room.mode,
                results.map((r) => ({ rank: r.rank, nickname: r.player.nickname })),
                me.player.nickname,
              ).text
            }
            fileName={`mathquest-multiplayer-${me.player.nickname}.png`}
            label={me.rank === 1 ? "Pamerin kemenanganmu" : "Bagikan hasil"}
            className="flex items-center justify-center gap-2 rounded-full bg-foreground py-3 text-sm font-semibold text-background"
          />
        )}
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
