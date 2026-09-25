import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { privateApi } from "@/lib/http";

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
  completedAt: string;
  rolledBack: boolean;
};

export type AdventureState = {
  currentCheckpoint: number;
  totalCheckpoints: number;
  questionsPerCheckpoint: number;
  maxFails: number;
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
};

export function useAdventureQuery() {
  return useQuery({
    queryKey: ["adventure"],
    queryFn: async () => (await privateApi.get<AdventureState>("/app/adventure")).data,
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
