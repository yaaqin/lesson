"use client";

import { useUserCourseQuery, type CoursePosition, type UserCourseModule } from "@/hooks/use-admin-course";

// Potongan UI monitoring Course Coding Dasar, dipakai di /admin/course/progress
// & detail user.

export function formatDateTime(value?: string) {
  if (!value) return "—";
  return new Date(value).toLocaleString("id-ID", { dateStyle: "medium", timeStyle: "short" });
}

export function formatDuration(ms?: number) {
  if (ms === undefined) return "";
  const seconds = Math.round(ms / 1000);
  return seconds < 60 ? `${seconds} dtk` : `${Math.floor(seconds / 60)} mnt ${seconds % 60} dtk`;
}

export function ProgressBar({ value, max }: { value: number; max: number }) {
  const percent = max > 0 ? Math.round((value / max) * 100) : 0;
  return (
    <div className="h-2 w-full overflow-hidden rounded-full bg-zinc-100 dark:bg-zinc-800">
      <div className="h-full rounded-full bg-blue-600 dark:bg-blue-500" style={{ width: `${percent}%` }} />
    </div>
  );
}

// PositionText: "udah sampe mana" dalam satu baris.
export function PositionText({ position }: { position: CoursePosition }) {
  if (position.finished) {
    return <span className="text-sm font-medium text-green-600 dark:text-green-400">✓ Selesai semua module</span>;
  }
  if (!position.currentModule) return <span className="text-sm text-zinc-400 dark:text-zinc-600">—</span>;
  return (
    <span className="flex flex-col">
      <span className="text-sm text-zinc-700 dark:text-zinc-300">
        {position.currentIsExam ? "📝 " : ""}
        {position.currentStep}
      </span>
      <span className="text-xs text-zinc-500 dark:text-zinc-500">{position.currentModule}</span>
    </span>
  );
}

// UserCourseCard: progress course 1 murid -- ringkasan, per module/level,
// dan attempt terakhir.
export function UserCourseCard({ userId }: { userId: string }) {
  const query = useUserCourseQuery(userId);
  const course = query.data;

  return (
    <div className="flex flex-col gap-4 rounded-2xl border border-black/[.08] bg-white p-5 dark:border-white/[.145] dark:bg-zinc-900">
      <span className="text-sm font-semibold text-zinc-500 dark:text-zinc-500">Progress Course Coding</span>

      {query.isLoading && <p className="text-sm text-zinc-500 dark:text-zinc-500">Memuat progress…</p>}
      {query.isError && <p className="text-sm text-red-500">Gagal memuat progress course.</p>}

      {course && course.attempts === 0 && (
        <p className="text-sm text-zinc-500 dark:text-zinc-500">Belum pernah mulai course.</p>
      )}

      {course && course.attempts > 0 && (
        <>
          <div className="grid grid-cols-1 gap-4 sm:grid-cols-3">
            <div className="flex flex-col gap-1.5">
              <span className="text-xs text-zinc-500 dark:text-zinc-500">Level lulus</span>
              <span className="text-xl font-bold text-black tabular-nums dark:text-zinc-50">
                {course.levelsPassed}/{course.totalLevels}
              </span>
              <ProgressBar value={course.levelsPassed} max={course.totalLevels} />
            </div>
            <div className="flex flex-col gap-1.5">
              <span className="text-xs text-zinc-500 dark:text-zinc-500">Module selesai</span>
              <span className="text-xl font-bold text-black tabular-nums dark:text-zinc-50">
                {course.modulesCompleted}/{course.totalModules}
              </span>
            </div>
            <div className="flex flex-col gap-1.5">
              <span className="text-xs text-zinc-500 dark:text-zinc-500">Lagi di</span>
              <PositionText position={course} />
            </div>
          </div>
          <span className="text-xs text-zinc-400 dark:text-zinc-600">
            {course.attempts} attempt · aktif terakhir {formatDateTime(course.lastActivityAt)}
          </span>

          <div className="flex flex-col gap-3">
            {course.modules
              .filter((m) => !m.comingSoon)
              .map((m) => (
                <ModuleProgress key={m.code} module={m} />
              ))}
          </div>

          <div className="flex flex-col gap-2">
            <span className="text-xs font-semibold tracking-wide text-zinc-500 uppercase dark:text-zinc-500">
              Attempt terakhir
            </span>
            <ul className="flex flex-col divide-y divide-black/[.06] text-sm dark:divide-white/[.08]">
              {course.recent.map((a) => (
                <li key={a.id} className="flex items-center justify-between gap-3 py-2">
                  <span className="flex min-w-0 flex-col">
                    <span className="truncate text-zinc-700 dark:text-zinc-300">
                      {a.isExam ? `📝 Final test ${a.moduleTitle}` : a.levelTitle}
                    </span>
                    <span className="text-xs text-zinc-400 dark:text-zinc-600">
                      {formatDateTime(a.startedAt)}
                      {a.isExam && a.durationMs !== undefined && ` · ${formatDuration(a.durationMs)}`}
                    </span>
                  </span>
                  {a.passed === undefined ? (
                    <span className="shrink-0 text-xs text-zinc-400 dark:text-zinc-600">belum dikumpulin</span>
                  ) : (
                    <span
                      className={`shrink-0 text-xs font-medium tabular-nums ${
                        a.passed ? "text-green-600 dark:text-green-400" : "text-red-500"
                      }`}
                    >
                      {a.correctCount}/{a.total} · {a.scorePercent}% {a.passed ? "✓" : "✗"}
                    </span>
                  )}
                </li>
              ))}
            </ul>
          </div>
        </>
      )}
    </div>
  );
}

function StatusIcon({ passed, locked }: { passed: boolean; locked: boolean }) {
  if (passed) return <span className="text-green-600 dark:text-green-400">✓</span>;
  if (locked) return <span className="opacity-50">🔒</span>;
  return <span className="text-blue-600 dark:text-blue-400">●</span>;
}

function ModuleProgress({ module: m }: { module: UserCourseModule }) {
  const passed = m.levels.filter((l) => l.passed).length;
  return (
    <div className="flex flex-col gap-2 rounded-xl border border-black/[.06] p-4 dark:border-white/[.08]">
      <div className="flex items-center justify-between gap-2">
        <span className="font-medium text-black dark:text-zinc-50">{m.title}</span>
        <span
          className={`rounded-full px-2.5 py-0.5 text-xs font-medium ${
            m.completed
              ? "bg-green-100 text-green-700 dark:bg-green-500/10 dark:text-green-400"
              : m.locked
                ? "bg-zinc-100 text-zinc-500 dark:bg-zinc-800 dark:text-zinc-500"
                : "bg-blue-100 text-blue-700 dark:bg-blue-500/10 dark:text-blue-400"
          }`}
        >
          {m.completed ? "Selesai" : m.locked ? "Kekunci" : `${passed}/${m.levels.length} level`}
        </span>
      </div>
      <ul className="flex flex-col gap-1 text-sm">
        {m.levels.map((l, i) => (
          <li key={l.id} className="flex items-center justify-between gap-3">
            <span className="flex min-w-0 items-center gap-2">
              <StatusIcon passed={l.passed} locked={l.locked} />
              <span className={`truncate ${l.locked ? "text-zinc-400 dark:text-zinc-600" : "text-zinc-700 dark:text-zinc-300"}`}>
                {i + 1}. {l.title}
              </span>
            </span>
            <span className="shrink-0 text-xs text-zinc-500 tabular-nums dark:text-zinc-500">
              {l.attempts > 0 ? `${l.attempts}× · terbaik ${l.bestScore ?? 0}%` : "—"}
            </span>
          </li>
        ))}
        <li className="flex items-center justify-between gap-3 border-t border-black/[.06] pt-1 dark:border-white/[.08]">
          <span className="flex items-center gap-2">
            <StatusIcon passed={m.exam.passed} locked={m.exam.locked} />
            <span className={m.exam.locked ? "text-zinc-400 dark:text-zinc-600" : "text-zinc-700 dark:text-zinc-300"}>
              📝 Final test
            </span>
          </span>
          <span className="shrink-0 text-xs text-zinc-500 tabular-nums dark:text-zinc-500">
            {m.exam.attempts > 0
              ? `${m.exam.attempts}× · terbaik ${m.exam.bestScore ?? 0}%${
                  m.exam.bestTimeMs !== undefined ? ` · ${formatDuration(m.exam.bestTimeMs)}` : ""
                }`
              : "—"}
          </span>
        </li>
      </ul>
    </div>
  );
}
