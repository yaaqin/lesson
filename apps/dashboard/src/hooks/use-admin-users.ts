import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { privateApi } from "@/lib/http";

export type AdminUserListItem = {
  id: string;
  email: string;
  displayName: string;
  currentStreak: number;
  longestStreak: number;
  livesRemaining: number;
  isPremium: boolean;
  roomQuota: number;
  createdAt: string;
};

export type AdminUserListResult = {
  items: AdminUserListItem[];
  total: number;
  page: number;
  pageSize: number;
};

export type AdminUserDetail = {
  id: string;
  email: string;
  displayName: string;
  role: "student" | "admin" | "superadmin";
  currentStreak: number;
  longestStreak: number;
  lastActiveDate?: string;
  livesRemaining: number;
  livesLastResetAt: string;
  totalAttempts: number;
  passedAttempts: number;
  // Premium: tag spesial -- bikin room multiplayer tanpa batas.
  isPremium: boolean;
  // Sisa kesempatan bikin room (user non-premium).
  roomQuota: number;
  createdAt: string;
};

export function useAdminUsersQuery({ page, pageSize, q }: { page: number; pageSize: number; q: string }) {
  return useQuery({
    queryKey: ["admin", "users", page, pageSize, q],
    queryFn: async () =>
      (
        await privateApi.get<AdminUserListResult>("/admin/users", {
          params: { page, pageSize, q: q || undefined },
        })
      ).data,
    placeholderData: (previous) => previous,
  });
}

export function useAdminUserDetailQuery(userId: string) {
  return useQuery({
    queryKey: ["admin", "users", "detail", userId],
    queryFn: async () => (await privateApi.get<AdminUserDetail>(`/admin/users/${userId}`)).data,
    enabled: !!userId,
  });
}

export function useResetUserLivesMutation(userId: string) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async () => privateApi.post(`/admin/users/${userId}/lives/reset`),
    onSuccess: () => {
      // prefix ["admin","users"] nyakup query list & detail sekaligus (queryKey detail
      // dimulai ["admin","users","detail",userId]).
      queryClient.invalidateQueries({ queryKey: ["admin", "users"] });
    },
  });
}

export function useSetUserPremiumMutation(userId: string) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (isPremium: boolean) => privateApi.put(`/admin/users/${userId}/premium`, { isPremium }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["admin", "users"] });
    },
  });
}

export function useSetUserRoomQuotaMutation(userId: string) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (roomQuota: number) => privateApi.put(`/admin/users/${userId}/room-quota`, { roomQuota }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["admin", "users"] });
    },
  });
}
