import { useCallback, useEffect, useRef, useState } from "react";
import { useMutation, useQuery } from "@tanstack/react-query";
import { privateApi } from "@/lib/http";
import type { UserAvatar } from "@/hooks/use-curriculum";

// Multiplayer: room dibikin user premium, pemain gabung pakai kode 6 huruf.
// Semua aksi game lewat WebSocket; server ngirim snapshot state utuh ("state")
// tiap ada perubahan, jadi klien cuma render ulang -- reconnect juga cukup
// nunggu snapshot pertama. Kontrak lengkap: services/api internal/multiplayer.

const API_BASE_URL = process.env.NEXT_PUBLIC_API_BASE_URL ?? "http://localhost:9801/api/v1";

export type MultiplayerMode = "classic" | "race";
export type MultiplayerFormat = "mc" | "essay" | "mixed";
export type MultiplayerTierCode = "sd" | "smp" | "smk" | "kampus";

export type MultiplayerSourceBatch = {
  id: string;
  name: string;
  questionCount: number;
  mcCount: number;
};

export type MultiplayerSourceTier = {
  tierCode: MultiplayerTierCode;
  tierName: string;
  batches: MultiplayerSourceBatch[];
};

export type MultiplayerOptions = {
  tiers: MultiplayerSourceTier[];
  classicQuestionCounts: number[];
  raceQuestionCounts: number[];
  secondsPerQuestionOptions: number[];
  maxPlayers: number;
};

export type CreateRoomInput = {
  mode: MultiplayerMode;
  questionCount: number;
  format: MultiplayerFormat;
  secondsPerQuestion: number;
  batchIds: string[];
  hostPlays: boolean;
};

export type PlayerProfile = {
  userId: string;
  nickname: string;
  avatar: UserAvatar;
  isPremium: boolean;
};

export type RoomPhase = "lobby" | "countdown" | "question" | "reveal" | "finished" | "closed";

export type RoomState = {
  type: "state";
  serverNow: number;
  room: {
    code: string;
    mode: MultiplayerMode;
    questionCount: number;
    format: MultiplayerFormat;
    secondsPerQuestion: number;
    sources: string[];
    hostId: string;
    maxPlayers: number;
    minPlayers: number;
  };
  you: { userId: string; isHost: boolean; playing: boolean };
  phase: RoomPhase;
  phaseEndsAt: number;
  players: (PlayerProfile & { isHost: boolean; playing: boolean; connected: boolean; left: boolean })[];
  question?: {
    index: number;
    total: number;
    tierCode: MultiplayerTierCode;
    type: "multiple_choice" | "essay_numeric";
    prompt: string;
    options?: number[];
  };
  answeredCount: number;
  playingCount: number;
  myAnswer?: { value: number; correct?: boolean; points?: number };
  reveal?: {
    correctValue: number;
    correctCount: number;
    answeredCount: number;
    winner?: PlayerProfile;
  };
  results?: {
    rank: number;
    player: PlayerProfile;
    score: number;
    correctCount: number;
    avgCorrectMs: number;
    left: boolean;
  }[];
};

export function useMultiplayerOptionsQuery() {
  return useQuery({
    queryKey: ["multiplayer", "options"],
    queryFn: async () => (await privateApi.get<MultiplayerOptions>("/app/multiplayer/options")).data,
  });
}

export function useCreateRoomMutation() {
  return useMutation({
    mutationFn: async (input: CreateRoomInput) =>
      (await privateApi.post<{ code: string }>("/app/multiplayer/rooms", input)).data,
  });
}

export function errorCode(err: unknown) {
  return (err as { response?: { data?: { error?: string } } })?.response?.data?.error;
}

// Kode close custom dari server yang artinya "jangan reconnect".
const CLOSE_REPLACED = 4001;
const CLOSE_KICKED = 4003;
const CLOSE_ROOM_NOT_FOUND = 4004;

// Error join yang final (gak ada gunanya dicoba lagi).
const FATAL_JOIN_ERRORS = new Set(["room_not_found", "room_full", "game_started", "kicked", "nickname_required"]);

export type ConnectionStatus =
  | { kind: "connecting" }
  | { kind: "open" }
  | { kind: "reconnecting" }
  | { kind: "ended"; reason: "closed" | "kicked" | "replaced" | "left" | string };

export type ClientMessage =
  | { type: "answer"; index: number; value: number }
  | { type: "start" }
  | { type: "kick"; userId: string }
  | { type: "set_host_plays"; plays: boolean }
  | { type: "close" }
  | { type: "leave" };

// useRoomSocket: minta ticket (HTTP, lewat auth + auto refresh token), buka
// WebSocket, dan otomatis nyambung lagi kalau putus -- penting di iOS yang
// mutus koneksi tiap PWA ditinggal sebentar. Balik ke app langsung reconnect.
export function useRoomSocket(code: string) {
  const [state, setState] = useState<RoomState | null>(null);
  const [status, setStatus] = useState<ConnectionStatus>({ kind: "connecting" });
  const [lastError, setLastError] = useState<{ code: string; at: number } | null>(null);
  // Selisih jam server - jam device, biar timer akurat walau jam HP ngaco.
  const [clockOffset, setClockOffset] = useState(0);

  const wsRef = useRef<WebSocket | null>(null);
  const endedRef = useRef(false);
  const attemptRef = useRef(0);
  const retryTimerRef = useRef<number | null>(null);
  const connectRef = useRef<() => void>(() => {});
  const genRef = useRef(0);
  const connectingRef = useRef(false);

  const end = useCallback((reason: string) => {
    endedRef.current = true;
    setStatus({ kind: "ended", reason });
  }, []);

  const scheduleRetry = useCallback(() => {
    if (endedRef.current || retryTimerRef.current !== null) return;
    setStatus({ kind: "reconnecting" });
    const delay = Math.min(1000 * 2 ** attemptRef.current, 8000);
    attemptRef.current += 1;
    retryTimerRef.current = window.setTimeout(() => {
      retryTimerRef.current = null;
      connectRef.current();
    }, delay);
  }, []);

  const connect = useCallback(async () => {
    if (endedRef.current || connectingRef.current) return;
    const current = wsRef.current;
    if (current && (current.readyState === WebSocket.OPEN || current.readyState === WebSocket.CONNECTING)) return;

    // gen berubah tiap mount/unmount -- hasil connect dari mount lama (mis.
    // StrictMode dev yang mount 2x) dibuang biar gak ada 2 socket rebutan.
    const gen = genRef.current;
    connectingRef.current = true;
    let ticket: string;
    try {
      ticket = (await privateApi.post<{ ticket: string }>(`/app/multiplayer/rooms/${code}/join`)).data.ticket;
    } catch (err) {
      if (gen !== genRef.current) return;
      connectingRef.current = false;
      const errCode = errorCode(err);
      if (errCode && FATAL_JOIN_ERRORS.has(errCode)) {
        end(errCode);
      } else {
        scheduleRetry();
      }
      return;
    }
    if (gen !== genRef.current) return;
    connectingRef.current = false;
    if (endedRef.current) return;

    const ws = new WebSocket(`${API_BASE_URL.replace(/^http/, "ws")}/app/multiplayer/ws?ticket=${ticket}`);
    wsRef.current = ws;

    ws.onopen = () => {
      attemptRef.current = 0;
      setStatus({ kind: "open" });
    };
    ws.onmessage = (event) => {
      if (wsRef.current !== ws) return;
      let msg: { type: string; [key: string]: unknown };
      try {
        msg = JSON.parse(event.data);
      } catch {
        return;
      }
      if (msg.type === "state") {
        const next = msg as unknown as RoomState;
        setClockOffset(next.serverNow - Date.now());
        setState(next);
      } else if (msg.type === "kicked") {
        end("kicked");
      } else if (msg.type === "closed") {
        end("closed");
      } else if (msg.type === "error") {
        setLastError({ code: String(msg.code), at: Date.now() });
      }
    };
    ws.onclose = (event) => {
      if (wsRef.current !== ws) return;
      wsRef.current = null;
      if (event.code === CLOSE_KICKED) return end("kicked");
      if (event.code === CLOSE_REPLACED) return end("replaced");
      if (event.code === CLOSE_ROOM_NOT_FOUND) return end("room_not_found");
      scheduleRetry();
    };
  }, [code, end, scheduleRetry]);

  useEffect(() => {
    connectRef.current = () => void connect();
  }, [connect]);

  useEffect(() => {
    genRef.current += 1;
    endedRef.current = false;
    connectingRef.current = false;
    connectRef.current();

    const onVisible = () => {
      if (document.visibilityState !== "visible" || endedRef.current) return;
      if (retryTimerRef.current !== null) {
        window.clearTimeout(retryTimerRef.current);
        retryTimerRef.current = null;
      }
      attemptRef.current = 0;
      connectRef.current();
    };
    document.addEventListener("visibilitychange", onVisible);
    window.addEventListener("online", onVisible);

    return () => {
      genRef.current += 1;
      endedRef.current = true;
      document.removeEventListener("visibilitychange", onVisible);
      window.removeEventListener("online", onVisible);
      if (retryTimerRef.current !== null) window.clearTimeout(retryTimerRef.current);
      const ws = wsRef.current;
      wsRef.current = null;
      ws?.close(1000, "unmount");
    };
  }, [code]);

  const send = useCallback((msg: ClientMessage) => {
    const ws = wsRef.current;
    if (!ws || ws.readyState !== WebSocket.OPEN) return false;
    ws.send(JSON.stringify(msg));
    return true;
  }, []);

  const leave = useCallback(() => {
    send({ type: "leave" });
    end("left");
  }, [send, end]);

  return { state, status, lastError, clockOffset, send, leave };
}

// useCountdown: sisa detik sampai `endsAt` (epoch ms jam server).
export function useCountdown(endsAt: number, clockOffset: number) {
  const [now, setNow] = useState(() => Date.now());
  useEffect(() => {
    if (!endsAt) return;
    const id = window.setInterval(() => setNow(Date.now()), 200);
    return () => window.clearInterval(id);
  }, [endsAt]);
  if (!endsAt) return 0;
  return Math.max(0, Math.ceil((endsAt - (now + clockOffset)) / 1000));
}
