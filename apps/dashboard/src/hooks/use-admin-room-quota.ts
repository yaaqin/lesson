import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { privateApi } from "@/lib/http";

// Permintaan tambahan kesempatan bikin room multiplayer dari user (apps/web).
// Approve = TAMBAH `grant` kesempatan ke kuota user.
export type RoomQuotaRequestStatus = "pending" | "approved" | "rejected";

export type AdminRoomQuotaRequest = {
  id: string;
  message: string;
  status: RoomQuotaRequestStatus;
  grantedQuota: number | null;
  createdAt: string;
  decidedAt: string | null;
  user: {
    id: string;
    email: string;
    displayName: string;
    nickname: string | null;
    roomQuota: number;
    isPremium: boolean;
  };
};

export type AdminRoomQuotaRequestList = {
  items: AdminRoomQuotaRequest[];
  total: number;
  pendingCount: number;
  page: number;
  pageSize: number;
};

export const MAX_ROOM_QUOTA_GRANT = 100;

export function useRoomQuotaRequestsQuery({ status, page, pageSize }: { status: RoomQuotaRequestStatus | ""; page: number; pageSize: number }) {
  return useQuery({
    queryKey: ["admin", "room-quota-requests", status, page, pageSize],
    queryFn: async () =>
      (
        await privateApi.get<AdminRoomQuotaRequestList>("/admin/room-quota-requests", {
          params: { status: status || undefined, page, pageSize },
        })
      ).data,
    placeholderData: (previous) => previous,
  });
}

function useInvalidateRequests() {
  const queryClient = useQueryClient();
  return () => {
    queryClient.invalidateQueries({ queryKey: ["admin", "room-quota-requests"] });
    queryClient.invalidateQueries({ queryKey: ["admin", "users"] });
  };
}

export function useApproveRoomQuotaRequestMutation() {
  const invalidate = useInvalidateRequests();
  return useMutation({
    mutationFn: async ({ id, grant }: { id: string; grant: number }) =>
      privateApi.post(`/admin/room-quota-requests/${id}/approve`, { grant }),
    onSettled: invalidate,
  });
}

export function useRejectRoomQuotaRequestMutation() {
  const invalidate = useInvalidateRequests();
  return useMutation({
    mutationFn: async (id: string) => privateApi.post(`/admin/room-quota-requests/${id}/reject`),
    onSettled: invalidate,
  });
}
