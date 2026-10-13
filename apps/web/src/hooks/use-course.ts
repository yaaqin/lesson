import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { privateApi } from "@/lib/http";

// Course Coding Dasar -- kontrak lengkap lihat services/api
// internal/coursesvc/types.go. Gak ada nyawa: latihan & final test boleh
// diulang terus, soal diacak ulang tiap mulai.

export type MaterialBlock =
  | { type: "heading"; text: string }
  | { type: "paragraph"; text: string }
  | { type: "code"; title?: string; code: string; output?: string }
  | { type: "example"; title?: string; text?: string; code: string; output?: string }
  | { type: "tip"; text: string }
  | { type: "list"; items: string[] }
  | { type: "table"; headers: string[]; rows: string[][] };

export type CourseLevelItem = {
  id: string;
  code: string;
  title: string;
  summary: string;
  questionCount: number;
  locked: boolean;
  passed: boolean;
};

export type CourseExamInfo = {
  locked: boolean;
  passed: boolean;
  questionCount: number;
  bankSize: number;
  timeLimitSeconds: number;
  passPercent: number;
  bestScore?: number;
  bestTimeMs?: number;
};

export type CourseModule = {
  code: string;
  title: string;
  description: string;
  locked: boolean;
  completed: boolean;
  comingSoon: boolean;
  levels: CourseLevelItem[];
  exam: CourseExamInfo;
};

export type CourseLevelDetail = {
  id: string;
  code: string;
  title: string;
  summary: string;
  moduleCode: string;
  moduleTitle: string;
  levelNumber: number;
  questionCount: number;
  passed: boolean;
  material: MaterialBlock[];
};

export type CourseQuestionKind = "multiple_choice" | "type_answer" | "write_value";

export type CourseQuestion = {
  id: string;
  kind: CourseQuestionKind;
  prompt: string;
  code?: string;
  options?: string[];
  placeholder?: string;
  strictJson?: boolean;
};

export type CourseStartResult = {
  attemptId: string;
  isExam: boolean;
  title: string;
  moduleCode: string;
  passPercent: number;
  timeLimitSeconds?: number;
  questions: CourseQuestion[];
};

export type CourseAnswer = { questionId: string; choice?: number; text?: string };

// expected / explanation / feedback kosong kalau admin nyetel mode
// result_only (server sengaja gak ngirim).
export type CourseQuestionResult = {
  questionId: string;
  correct: boolean;
  feedback?: string;
  expected?: string;
  explanation?: string;
};

// Cara nampilin pembahasan, diatur admin dari dashboard (/admin/course).
export type CourseReviewSetting = { mode: "full" | "timed" | "result_only"; seconds: number };

export type CourseSubmitResult = {
  passed: boolean;
  scorePercent: number;
  correctCount: number;
  total: number;
  passPercent: number;
  durationMs: number;
  timedOut?: boolean;
  results: CourseQuestionResult[];
  review: CourseReviewSetting;
  nextLevelId?: string;
  examUnlocked?: boolean;
  nextModuleUnlocked?: string;
};

export function useCourseModulesQuery() {
  return useQuery({
    queryKey: ["course-modules"],
    queryFn: async () => (await privateApi.get<CourseModule[]>("/app/course/modules")).data,
  });
}

export function useCourseLevelQuery(levelId: string) {
  return useQuery({
    queryKey: ["course-level", levelId],
    queryFn: async () => (await privateApi.get<CourseLevelDetail>(`/app/course/levels/${levelId}`)).data,
    enabled: !!levelId,
    retry: false,
  });
}

export function useStartCourseLevelMutation() {
  return useMutation({
    mutationFn: async (levelId: string) =>
      (await privateApi.post<CourseStartResult>(`/app/course/levels/${levelId}/start`)).data,
  });
}

export function useStartCourseExamMutation() {
  return useMutation({
    mutationFn: async (moduleCode: string) =>
      (await privateApi.post<CourseStartResult>(`/app/course/modules/${moduleCode}/exam/start`)).data,
  });
}

export function useSubmitCourseAttemptMutation() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: async ({ attemptId, answers }: { attemptId: string; answers: CourseAnswer[] }) =>
      (await privateApi.post<CourseSubmitResult>(`/app/course/attempts/${attemptId}/submit`, { answers })).data,
    // lulus bisa buka level / final test / module berikutnya
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["course-modules"] });
      queryClient.invalidateQueries({ queryKey: ["course-level"] });
    },
  });
}

export function formatDuration(ms: number): string {
  const total = Math.round(ms / 1000);
  const m = Math.floor(total / 60);
  const s = total % 60;
  return `${m}:${String(s).padStart(2, "0")}`;
}
