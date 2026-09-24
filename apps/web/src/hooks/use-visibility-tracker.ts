import { useCallback, useEffect, useRef } from "react";
import type { SecurityEvent } from "@/hooks/use-security-event-reporter";

// Kepergian super singkat (mis. prompt izin browser yang nyomot fokus sekejap)
// dianggep noise, gak dicatat.
const MIN_AWAY_MS = 500;

type Episode = { startedAt: number; hidden: boolean; questionIndex: number | null };

// useVisibilityTracker: deteksi user ninggalin halaman soal -- ganti tab /
// minimize (Page Visibility API) atau window kehilangan fokus (blur), termasuk
// pas buka sidebar extension kayak Gemini di Chrome. Pindah tab mancing blur DAN
// hidden sekaligus, jadi keduanya digabung jadi satu "episode pergi": dicatat
// sebagai tab_hidden kalau sempat hidden, window_blur kalau cuma blur. Episode
// ditutup (dan durasinya dihitung) begitu halaman keliatan & fokus lagi.
export function useVisibilityTracker({
  enabled,
  getQuestionIndex,
  onAwayEnd,
}: {
  enabled: boolean;
  getQuestionIndex: () => number | null;
  onAwayEnd: (event: SecurityEvent) => void;
}) {
  const episodeRef = useRef<Episode | null>(null);
  const getQuestionIndexRef = useRef(getQuestionIndex);
  const onAwayEndRef = useRef(onAwayEnd);

  useEffect(() => {
    getQuestionIndexRef.current = getQuestionIndex;
    onAwayEndRef.current = onAwayEnd;
  }, [getQuestionIndex, onAwayEnd]);

  const closeEpisode = useCallback(() => {
    const episode = episodeRef.current;
    if (!episode) return;
    episodeRef.current = null;
    const durationMs = Date.now() - episode.startedAt;
    if (durationMs < MIN_AWAY_MS) return;
    onAwayEndRef.current({
      eventType: episode.hidden ? "tab_hidden" : "window_blur",
      questionIndex: episode.questionIndex,
      startedAt: new Date(episode.startedAt).toISOString(),
      durationMs,
    });
  }, []);

  useEffect(() => {
    if (!enabled) {
      episodeRef.current = null;
      return;
    }

    const markAway = (hidden: boolean) => {
      if (episodeRef.current) {
        episodeRef.current.hidden ||= hidden;
        return;
      }
      episodeRef.current = { startedAt: Date.now(), hidden, questionIndex: getQuestionIndexRef.current() };
    };

    const maybeBack = () => {
      if (document.visibilityState === "visible" && document.hasFocus()) closeEpisode();
    };

    const onVisibilityChange = () => {
      if (document.visibilityState === "hidden") markAway(true);
      else maybeBack();
    };
    const onBlur = () => markAway(false);

    document.addEventListener("visibilitychange", onVisibilityChange);
    window.addEventListener("blur", onBlur);
    window.addEventListener("focus", maybeBack);
    return () => {
      document.removeEventListener("visibilitychange", onVisibilityChange);
      window.removeEventListener("blur", onBlur);
      window.removeEventListener("focus", maybeBack);
    };
  }, [enabled, closeEpisode]);

  // closeOpenEpisode: dipanggil pas submit -- kalau attempt kelar (mis. waktu
  // ujian habis) selagi user masih di luar halaman, episodenya tetap kehitung.
  return { closeOpenEpisode: closeEpisode };
}
