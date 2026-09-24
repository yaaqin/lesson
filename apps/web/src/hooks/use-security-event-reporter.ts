import { useCallback, useEffect, useRef } from "react";
import { privateApi } from "@/lib/http";

// Anti-cheating: sinyal mencurigakan selama attempt jalan. Dikirim batch ke
// services/api (POST /app/attempts/{id}/security-events), lalu diakumulasi jadi
// risk score pas submit -- lihat services/api internal/curriculumsvc/security.go.
export type SecurityEventType = "tab_hidden" | "window_blur" | "fullscreen_exit" | "copy_blocked";

export type SecurityEvent = {
  eventType: SecurityEventType;
  questionIndex: number | null;
  startedAt: string;
  durationMs: number | null;
};

const FLUSH_EVERY_EVENTS = 5;
const FLUSH_INTERVAL_MS = 15_000;

// useSecurityEventReporter: buffer event di memori, kirim batch tiap 5 event,
// tiap 15 detik, atau pas halaman manggil flush() (tiap ganti soal) -- bukan
// per event biar gak nge-spam API. Sisa buffer gak dikirim sendiri pas submit,
// tapi ikut body submit lewat drainForSubmit() biar pasti kehitung di risk score.
export function useSecurityEventReporter(attemptId: string | null) {
  const bufferRef = useRef<SecurityEvent[]>([]);
  const inflightRef = useRef<Promise<void>>(Promise.resolve());

  useEffect(() => {
    bufferRef.current = [];
  }, [attemptId]);

  const flush = useCallback(() => {
    if (!attemptId || bufferRef.current.length === 0) return;
    const batch = bufferRef.current;
    bufferRef.current = [];
    inflightRef.current = inflightRef.current
      .then(() => privateApi.post(`/app/attempts/${attemptId}/security-events`, { events: batch }))
      .then(
        () => undefined,
        () => {
          // gagal kirim -> balikin ke buffer biar ikut kebawa di batch berikut / body submit
          bufferRef.current = [...batch, ...bufferRef.current];
        },
      );
  }, [attemptId]);

  const record = useCallback(
    (event: SecurityEvent) => {
      bufferRef.current.push(event);
      if (bufferRef.current.length >= FLUSH_EVERY_EVENTS) flush();
    },
    [flush],
  );

  // Tunggu batch yang lagi jalan kelar dulu (biar gak balapan sama submit yang
  // nutup attempt), lalu ambil sisa buffer buat dikirim bareng submit.
  const drainForSubmit = useCallback(async () => {
    await inflightRef.current;
    const rest = bufferRef.current;
    bufferRef.current = [];
    return rest;
  }, []);

  useEffect(() => {
    if (!attemptId) return;
    const id = window.setInterval(flush, FLUSH_INTERVAL_MS);
    return () => window.clearInterval(id);
  }, [attemptId, flush]);

  return { record, flush, drainForSubmit };
}
