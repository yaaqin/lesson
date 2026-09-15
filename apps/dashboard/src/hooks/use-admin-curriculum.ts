import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { privateApi } from "@/lib/http";

export type AdminChallenge = {
  id: string;
  name: string;
  isExam: boolean;
  timeLimitSeconds: number;
  questionCountRequired: number;
  passThresholdPercent: number;
  optionCount: number;
  questionBankSize: number;
};

export type AdminBatch = {
  id: string;
  name: string;
  challenges: AdminChallenge[];
};

export type AdminTier = {
  id: string;
  code: string;
  name: string;
  batches: AdminBatch[];
};

export type AdminCategory = {
  id: string;
  code: string;
  name: string;
  challenges: AdminChallenge[];
};

export function useAdminCurriculumQuery() {
  return useQuery({
    queryKey: ["admin", "curriculum"],
    queryFn: async () => (await privateApi.get<AdminTier[]>("/admin/curriculum")).data,
  });
}

export function useAdminCategoriesQuery(tierCode: string) {
  return useQuery({
    queryKey: ["admin", "categories", tierCode],
    queryFn: async () =>
      (await privateApi.get<AdminCategory[]>(`/admin/tiers/${tierCode}/categories`)).data,
    enabled: !!tierCode,
  });
}

export function useUpdateChallengeTimingMutation() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async ({ challengeId, timeLimitSeconds }: { challengeId: string; timeLimitSeconds: number }) =>
      privateApi.patch(`/admin/challenges/${challengeId}/timing`, { timeLimitSeconds }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["admin", "curriculum"] });
    },
  });
}
