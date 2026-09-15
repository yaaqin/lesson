import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { privateApi } from "@/lib/http";

export type PuzzleKind = "addition_grid" | "cryptarithm";

export type AdminPuzzleQuestion = {
  id: string;
  kind: PuzzleKind;
  prompt: string;
  status: "draft" | "published" | "archived";
  // addition_grid
  size?: number;
  solutionGrid?: number[][];
  givenMask?: boolean[][];
  // cryptarithm -- words termasuk kata hasil di elemen terakhir
  words?: string[];
  solution?: Record<string, number>;
};

export type PuzzleUpsertInput = {
  kind: PuzzleKind;
  prompt?: string;
  status: string;
  size?: number;
  solutionGrid?: number[][];
  givenMask?: boolean[][];
  words?: string[]; // addend, TANPA kata hasil
  result?: string;
};

const puzzleQueryKey = (challengeId: string) => ["admin", "puzzle-questions", challengeId];

export function useAdminPuzzleQuestionsQuery(challengeId: string) {
  return useQuery({
    queryKey: puzzleQueryKey(challengeId),
    queryFn: async () =>
      (await privateApi.get<AdminPuzzleQuestion[]>(`/admin/challenges/${challengeId}/puzzle-questions`)).data,
    enabled: !!challengeId,
  });
}

export function useCreatePuzzleQuestionMutation(challengeId: string) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (input: PuzzleUpsertInput) =>
      (
        await privateApi.post<AdminPuzzleQuestion>(
          `/admin/challenges/${challengeId}/puzzle-questions`,
          input,
        )
      ).data,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: puzzleQueryKey(challengeId) });
    },
  });
}

export function useUpdatePuzzleQuestionMutation(challengeId: string) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async ({ questionId, input }: { questionId: string; input: PuzzleUpsertInput }) =>
      (await privateApi.put<AdminPuzzleQuestion>(`/admin/puzzle-questions/${questionId}`, input)).data,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: puzzleQueryKey(challengeId) });
    },
  });
}

export function useDeletePuzzleQuestionMutation(challengeId: string) {
  const queryClient = useQueryClient();

  return useMutation({
    // reuse endpoint DELETE /admin/questions/{id} yang sama dipakai MC/essay --
    // puzzle_questions cascade otomatis (FK ON DELETE CASCADE), gak perlu endpoint sendiri.
    mutationFn: async (questionId: string) => privateApi.delete(`/admin/questions/${questionId}`),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: puzzleQueryKey(challengeId) });
    },
  });
}
