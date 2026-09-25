import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { privateApi } from "@/lib/http";
import type { UserAvatar } from "@/hooks/use-curriculum";

// Adventure: mode jalan terus lintas jenjang (SD -> Kampus), 40 checkpoint x
// 100 soal. Jawaban dicek per soal di server -- opsi gak bawa penanda benar.
// Kontrak lengkap: services/api internal/curriculumsvc/adventure.go.

export type AdventureTierCode = "sd" | "smp" | "smk" | "kampus";

export type AdventureQuestion = {
  id: string;
  tierCode: AdventureTierCode;
  type: "multiple_choice" | "essay_numeric";
  prompt: string;
  options?: number[];
  timeLimitSeconds: number;
};

export type AdventureHistoryItem = {
  attemptId: string;
  checkpointNo: number;
  durationSeconds: number;
  correctCount: number;
  failsRemaining: number;
  failedAttempts: number;
  // Skor waktu: waktu disediain, waktu setelah dipotong bonus sisa jatah, dan
  // persentasenya (makin kecil makin cepat).
  timeAllowedSeconds: number;
  adjustedSeconds: number;
  timePercent: number;
  completedAt: string;
  rolledBack: boolean;
};

export type AdventureState = {
  currentCheckpoint: number;
  totalCheckpoints: number;
  questionsPerCheckpoint: number;
  maxFails: number;
  bonusSecondsPerFail: number;
  completed: boolean;
  history: AdventureHistoryItem[];
};

export type AdventureStartResult = {
  attemptId: string;
  checkpointNo: number;
  totalCheckpoints: number;
  questionOffset: number;
  maxFails: number;
  questions: AdventureQuestion[];
};

export type AdventureAnswerResult = {
  correct: boolean;
  timedOut: boolean;
  failsUsed: number;
  failsRemaining: number;
  correctCount: number;
  status: "in_progress" | "passed" | "failed";
  nextIndex: number;
  durationSeconds?: number;
  nextCheckpoint?: number;
  adventureCompleted?: boolean;
  timeAllowedSeconds?: number;
  adjustedSeconds?: number;
  timePercent?: number;
};

export type AdventureLeaderboardEntry = {
  rank: number;
  nickname: string;
  avatar: UserAvatar;
  checkpointsCleared: number;
  questionsCleared: number;
  completed: boolean;
  timeAllowedSeconds: number;
  adjustedSeconds: number;
  timePercent: number;
  lastClearedAt: string;
  isMe: boolean;
};

export type AdventureLeaderboard = {
  entries: AdventureLeaderboardEntry[];
  // Keisi kalau user sendiri ada di ranking tapi di luar top 100.
  me: AdventureLeaderboardEntry | null;
  totalRanked: number;
  totalCheckpoints: number;
  bonusSecondsPerFail: number;
};

export function useAdventureQuery() {
  return useQuery({
    queryKey: ["adventure"],
    queryFn: async () => (await privateApi.get<AdventureState>("/app/adventure")).data,
  });
}

export function useAdventureLeaderboardQuery() {
  return useQuery({
    queryKey: ["adventure", "leaderboard"],
    queryFn: async () => (await privateApi.get<AdventureLeaderboard>("/app/adventure/leaderboard")).data,
  });
}

export function useStartAdventureMutation() {
  return useMutation({
    mutationFn: async () => (await privateApi.post<AdventureStartResult>("/app/adventure/start")).data,
  });
}

export function useAnswerAdventureMutation() {
  return useMutation({
    mutationFn: async ({
      attemptId,
      questionIndex,
      selectedValue,
    }: {
      attemptId: string;
      questionIndex: number;
      selectedValue: number | null;
    }) =>
      (
        await privateApi.post<AdventureAnswerResult>(`/app/adventure/attempts/${attemptId}/answer`, {
          questionIndex,
          selectedValue,
        })
      ).data,
  });
}

export function useRollbackAdventureMutation() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (checkpoint: number) =>
      (await privateApi.post<AdventureState>("/app/adventure/rollback", { checkpoint })).data,
    onSuccess: (state) => {
      queryClient.setQueryData(["adventure"], state);
    },
  });
}
