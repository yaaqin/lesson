import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { publicApi, privateApi } from "@/lib/http";

export type Tier = { id: string; code: string; name: string; usesBatch: boolean };
export type Batch = { id: string; name: string };
export type Category = { id: string; code: string; name: string };

export type ChallengeListItem = {
  id: string;
  name: string;
  isExam: boolean;
  questionCountRequired: number;
  passThresholdPercent: number;
  timeLimitSeconds: number;
  completed: boolean;
  hasEssay: boolean;
};

export type QuestionType = "multiple_choice" | "essay_numeric" | "grid_puzzle";
export type QuestionOption = { value: number; isCorrect: boolean };

// PuzzlePayload: bentuk publik puzzle (grid_puzzle) -- gak pernah bawa solusi,
// cuma given + blankKeys yang harus diisi user. Lihat services/api
// internal/curriculumsvc/types.go (PuzzlePayload) buat kontrak lengkapnya.
export type PuzzlePayload = {
  kind: "addition_grid" | "cryptarithm";
  size?: number;
  words?: string[];
  given?: Record<string, number>;
  blankKeys: string[];
  rowSums?: number[];
  colSums?: number[];
};

export type SessionQuestion = {
  id: string;
  type: QuestionType;
  prompt: string;
  options?: QuestionOption[];
  correctAnswerValue?: number;
  puzzle?: PuzzlePayload;
};

export type StartResult = {
  attemptId: string;
  challengeName: string;
  isExam: boolean;
  timeLimitSeconds: number;
  passThresholdPercent: number;
  questions: SessionQuestion[];
};

export type SubmitAnswer = {
  questionId: string;
  selectedValue: number | null;
  selectedGrid?: Record<string, number>;
};

export type SubmitResult = {
  scorePercent: number;
  passed: boolean;
  correctCount: number;
  totalQuestions: number;
  livesRemaining: number;
  currentStreak: number;
};

export type MeInfo = {
  id: string;
  email: string;
  displayName: string;
  role: string;
  currentStreak: number;
  longestStreak: number;
  livesRemaining: number;
};

export function useTiersQuery() {
  return useQuery({
    queryKey: ["tiers"],
    queryFn: async () => (await publicApi.get<Tier[]>("/app/tiers")).data,
  });
}

export function useBatchesQuery(tierCode: string) {
  return useQuery({
    queryKey: ["batches", tierCode],
    queryFn: async () => (await publicApi.get<Batch[]>(`/app/tiers/${tierCode}/batches`)).data,
    enabled: !!tierCode,
  });
}

export function useChallengesQuery(batchId: string | undefined) {
  return useQuery({
    queryKey: ["challenges", batchId],
    queryFn: async () =>
      (await privateApi.get<ChallengeListItem[]>(`/app/batches/${batchId}/challenges`)).data,
    enabled: !!batchId,
  });
}

export function useCategoriesQuery(tierCode: string) {
  return useQuery({
    queryKey: ["categories", tierCode],
    queryFn: async () => (await publicApi.get<Category[]>(`/app/tiers/${tierCode}/categories`)).data,
    enabled: !!tierCode,
  });
}

export function useChallengesByCategoryQuery(categoryId: string | undefined) {
  return useQuery({
    queryKey: ["challenges-by-category", categoryId],
    queryFn: async () =>
      (await privateApi.get<ChallengeListItem[]>(`/app/categories/${categoryId}/challenges`)).data,
    enabled: !!categoryId,
  });
}

export function useMeQuery() {
  return useQuery({
    queryKey: ["me"],
    queryFn: async () => (await privateApi.get<MeInfo>("/app/me")).data,
    staleTime: 5_000,
  });
}

export function useStartChallengeMutation() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (challengeId: string) =>
      (await privateApi.post<StartResult>(`/app/challenges/${challengeId}/start`)).data,
    // start juga bisa memicu reset nyawa harian di server -> pastikan header gak nampilin
    // angka basi (lihat juga staleTime di useMeQuery).
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["me"] });
    },
  });
}

export function useSubmitAttemptMutation() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async ({
      attemptId,
      answers,
    }: {
      attemptId: string;
      answers: SubmitAnswer[];
    }) =>
      (await privateApi.post<SubmitResult>(`/app/attempts/${attemptId}/submit`, { answers }))
        .data,
    // nyawa/streak berubah di server begitu attempt disubmit -> invalidate biar
    // indikator di header (dan daftar "Selesai") langsung ke-refresh, gak nunggu
    // staleTime useMeQuery habis dulu.
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["me"] });
      queryClient.invalidateQueries({ queryKey: ["challenges"] });
      queryClient.invalidateQueries({ queryKey: ["challenges-by-category"] });
    },
  });
}
