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

// --- Monitoring progress (kontrak: services/api internal/coursesvc/admin_progress.go) ---

export type CoursePosition = {
  levelsPassed: number;
  totalLevels: number;
  modulesCompleted: number;
  totalModules: number;
  finished: boolean;
  currentModule?: string;
  currentStep?: string;
  currentIsExam?: boolean;
};

export type CourseLevelFunnel = { id: string; title: string; passedUsers: number };
export type CourseModuleFunnel = {
  code: string;
  title: string;
  comingSoon: boolean;
  levels: CourseLevelFunnel[];
  examPassedUsers: number;
};

export type CourseUserRow = CoursePosition & {
  userId: string;
  displayName: string;
  email: string;
  attempts: number;
  lastActivityAt: string;
};

export type CourseProgress = {
  startedUsers: number;
  completedUsers: number;
  modules: CourseModuleFunnel[];
  users: { items: CourseUserRow[]; total: number; page: number; pageSize: number };
};

export function useCourseProgressQuery({ page, pageSize, q }: { page: number; pageSize: number; q: string }) {
  return useQuery({
    queryKey: ["admin", "course", "progress", page, pageSize, q],
    queryFn: async () =>
      (await privateApi.get<CourseProgress>("/admin/course/progress", { params: { page, pageSize, q: q || undefined } })).data,
    placeholderData: (previous) => previous,
  });
}

export type UserCourseLevel = {
  id: string;
  title: string;
  locked: boolean;
  passed: boolean;
  attempts: number;
  bestScore?: number;
  lastAttemptAt?: string;
};

export type UserCourseModule = {
  code: string;
  title: string;
  locked: boolean;
  completed: boolean;
  comingSoon: boolean;
  levels: UserCourseLevel[];
  exam: {
    locked: boolean;
    passed: boolean;
    questionCount: number;
    bankSize: number;
    timeLimitSeconds: number;
    passPercent: number;
    bestScore?: number;
    bestTimeMs?: number;
    attempts: number;
    lastAttemptAt?: string;
  };
};

export type UserCourseAttempt = {
  id: string;
  moduleTitle: string;
  levelTitle?: string;
  isExam: boolean;
  total: number;
  correctCount?: number;
  scorePercent?: number;
  passed?: boolean;
  durationMs?: number;
  startedAt: string;
  completedAt?: string;
};

export type UserCourse = CoursePosition & {
  attempts: number;
  lastActivityAt?: string;
  modules: UserCourseModule[];
  recent: UserCourseAttempt[];
};

export function useUserCourseQuery(userId: string) {
  return useQuery({
    queryKey: ["admin", "users", userId, "course"],
    queryFn: async () => (await privateApi.get<UserCourse>(`/admin/users/${userId}/course`)).data,
    enabled: Boolean(userId),
  });
}
