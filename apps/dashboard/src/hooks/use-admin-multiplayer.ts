import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { privateApi } from "@/lib/http";

// Jeda antar soal multiplayer (apps/web). Berlaku buat room yang dibikin
// setelah disimpan -- room yang lagi jalan gak ikut berubah.
export type MultiplayerConfig = {
  classicCooldownSeconds: number;
  raceCooldownSeconds: number;
  raceWinnerRevealSeconds: number;
  resultsCountdownSeconds: number;
};

export function useMultiplayerConfigQuery() {
  return useQuery({
    queryKey: ["admin", "multiplayer", "config"],
    queryFn: async () => (await privateApi.get<MultiplayerConfig>("/admin/multiplayer/config")).data,
  });
}

export function useUpdateMultiplayerConfigMutation() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (config: MultiplayerConfig) => privateApi.put("/admin/multiplayer/config", config),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["admin", "multiplayer", "config"] });
    },
  });
}
