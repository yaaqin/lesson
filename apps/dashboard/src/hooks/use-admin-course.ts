import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { privateApi } from "@/lib/http";

// Cara nampilin pembahasan Course Coding Dasar (apps/web) setelah submit,
// dipisah buat latihan level & final test. Kontrak: services/api
// internal/coursesvc/config.go.
export type ReviewMode = "full" | "timed" | "result_only";
export type ReviewSetting = { mode: ReviewMode; seconds: number };
export type CourseConfig = { level: ReviewSetting; exam: ReviewSetting };

export function useCourseConfigQuery() {
  return useQuery({
    queryKey: ["admin", "course", "config"],
    queryFn: async () => (await privateApi.get<CourseConfig>("/admin/course/config")).data,
  });
}

export function useUpdateCourseConfigMutation() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (config: CourseConfig) => privateApi.put("/admin/course/config", config),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["admin", "course", "config"] });
    },
  });
}
