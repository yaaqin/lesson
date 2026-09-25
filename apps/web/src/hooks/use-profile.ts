import { useEffect } from "react";
import { useRouter } from "next/navigation";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { privateApi } from "@/lib/http";
import { useMeQuery, type MeInfo } from "@/hooks/use-curriculum";
import { useAuthStore } from "@/store/auth-store";

export type NicknameCheck = { nickname: string; valid: boolean; available: boolean };

export type ProfileUpdate = {
  nickname?: string;
  avatarType?: "character" | "google";
  avatarKey?: string;
};

export function useNicknameCheckQuery(nickname: string, enabled: boolean) {
  return useQuery({
    queryKey: ["nickname-check", nickname],
    queryFn: async () =>
      (await privateApi.get<NicknameCheck>("/app/me/nickname-check", { params: { nickname } })).data,
    enabled,
    staleTime: 10_000,
  });
}

export function useUpdateProfileMutation() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (update: ProfileUpdate) =>
      (await privateApi.patch<MeInfo>("/app/me/profile", update)).data,
    onSuccess: (me) => {
      queryClient.setQueryData(["me"], me);
      queryClient.invalidateQueries({ queryKey: ["nickname-check"] });
    },
  });
}

// useRequireNickname: dipanggil di halaman yang butuh login. User yang belum
// punya nickname (mis. baru pertama login Google) dilempar ke /onboarding.
export function useRequireNickname() {
  const router = useRouter();
  const session = useAuthStore((s) => s.session);
  const meQuery = useMeQuery();
  const me = meQuery.data;

  useEffect(() => {
    if (session && me && me.nickname === null) router.replace("/onboarding");
  }, [session, me, router]);

  return meQuery;
}
