import { useMutation, useQueries, useQuery, useQueryClient } from "@tanstack/react-query";
import { publicApi, privateApi } from "@/lib/http";
import type { SecurityEvent } from "@/hooks/use-security-event-reporter";

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
  // Kunci progres dari server: "batch" = ujian batch sebelumnya belum lulus,
  // "exam" = latihan di batch ini yang lulus belum cukup buat buka ujian.
  locked: boolean;
  lockReason?: "batch" | "exam";
  examRequiredPassed?: number;
  examPassedCount?: number;
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
  // Hasil scoring anti-cheating di server (lihat use-security-event-reporter).
  riskLevel: "normal" | "low_confidence" | "review";
};

export type MeInfo = {
  id: string;
  email: string;
  displayName: string;
  role: string;
  currentStreak: number;
  longestStreak: number;
  livesRemaining: number;
  // null = belum bikin nickname -> diarahin ke /onboarding (useRequireNickname).
  nickname: string | null;
  avatar: UserAvatar;
  themePreference: "light" | "dark" | "system";
};

export type UserAvatar = {
  type: "character" | "google";
  key: string;
  googleUrl: string | null;
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

// useChallengesForBatchesQuery: tier bisa punya lebih dari 1 batch (mis. SMP
// "Batch 1 — Semester 1" lanjut "Batch 2 — Soal Campuran") -- gabungin semua
// challenge-nya jadi satu list berurutan (urutan batch tetap dipertahankan)
// biar user ngerasain progresi yang menyambung, bukan cuma batch pertama.
export function useChallengesForBatchesQuery(batches: Batch[] | undefined) {
  const list = batches ?? [];
  const results = useQueries({
    queries: list.map((batch) => ({
      queryKey: ["challenges", batch.id],
      queryFn: async () =>
        (await privateApi.get<ChallengeListItem[]>(`/app/batches/${batch.id}/challenges`)).data,
      enabled: !!batch.id,
    })),
  });

  const isLoading = list.length === 0 ? false : results.some((r) => r.isLoading);
  const challenges = results.flatMap((r, i) =>
    (r.data ?? []).map((challenge) => ({ ...challenge, batchName: list[i].name })),
  );

  return { data: challenges, isLoading };
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

// Jenjang yang punya ranking (sinkron sama TierLeaderboardCodes di
// services/api curriculumsvc/tier_leaderboard.go).
export const TIER_LEADERBOARD_CODES = ["sd", "smp", "smk", "kampus"];

export type TierLeaderboardEntry = {
  rank: number;
  nickname: string;
  avatar: UserAvatar;
  points: number;
  challengesCompleted: number;
  examsPassed: number;
  lastScoredAt: string;
  isMe: boolean;
};

export type TierLeaderboard = {
  tierCode: string;
  tierName: string;
  entries: TierLeaderboardEntry[];
  // Keisi kalau user sendiri ada di ranking tapi di luar top 100.
  me: TierLeaderboardEntry | null;
  totalRanked: number;
};

export function useTierLeaderboardQuery(tierCode: string) {
  return useQuery({
    queryKey: ["tier-leaderboard", tierCode],
    queryFn: async () => (await privateApi.get<TierLeaderboard>(`/app/tiers/${tierCode}/leaderboard`)).data,
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
      securityEvents,
    }: {
      attemptId: string;
      answers: SubmitAnswer[];
      securityEvents: SecurityEvent[];
    }) =>
      (
        await privateApi.post<SubmitResult>(`/app/attempts/${attemptId}/submit`, {
          answers,
          securityEvents,
        })
      ).data,
    // nyawa/streak berubah di server begitu attempt disubmit -> invalidate biar
    // indikator di header (dan daftar "Selesai") langsung ke-refresh, gak nunggu
    // staleTime useMeQuery habis dulu.
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["me"] });
      queryClient.invalidateQueries({ queryKey: ["challenges"] });
      queryClient.invalidateQueries({ queryKey: ["challenges-by-category"] });
      queryClient.invalidateQueries({ queryKey: ["tier-leaderboard"] });
    },
  });
}
