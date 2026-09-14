import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { privateApi } from "@/lib/http";

export type AdminQuestionOption = { value: number; isCorrect: boolean };

export type AdminQuestion = {
  id: string;
  prompt: string;
  status: "draft" | "published" | "archived";
  options: AdminQuestionOption[];
};

export type QuestionInput = {
  prompt: string;
  status: string;
  options: AdminQuestionOption[];
};

export function useAdminQuestionsQuery(challengeId: string) {
  return useQuery({
    queryKey: ["admin", "questions", challengeId],
    queryFn: async () =>
      (await privateApi.get<AdminQuestion[]>(`/admin/challenges/${challengeId}/questions`)).data,
    enabled: !!challengeId,
  });
}

export function useCreateQuestionMutation(challengeId: string) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (input: QuestionInput) =>
      privateApi.post(`/admin/challenges/${challengeId}/questions`, input),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["admin", "questions", challengeId] });
      queryClient.invalidateQueries({ queryKey: ["admin", "curriculum"] });
    },
  });
}

export function useUpdateQuestionMutation(challengeId: string) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async ({ questionId, input }: { questionId: string; input: QuestionInput }) =>
      privateApi.put(`/admin/questions/${questionId}`, input),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["admin", "questions", challengeId] });
    },
  });
}

export function useDeleteQuestionMutation(challengeId: string) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (questionId: string) => privateApi.delete(`/admin/questions/${questionId}`),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["admin", "questions", challengeId] });
      queryClient.invalidateQueries({ queryKey: ["admin", "curriculum"] });
    },
  });
}
