package coursesvc

import (
	"context"
	"time"
)

// Monitoring progress Course Coding Dasar buat dashboard admin
// (/admin/course/progress & detail per user). Posisi user dihitung pakai
// aturan kunci yang sama kayak apps/web (structure + applyLocks).

// --- Ringkasan + daftar user ---

type AdminCourseProgress struct {
	// Murid yang udah pernah mulai course (minimal 1 attempt).
	StartedUsers   int                   `json:"startedUsers"`
	CompletedUsers int                   `json:"completedUsers"` // lulus final test semua module
	Modules        []AdminModuleFunnel   `json:"modules"`
	Users          AdminCourseUserResult `json:"users"`
}

// AdminModuleFunnel: berapa murid yang udah lulus tiap level & final test.
type AdminModuleFunnel struct {
	Code            string             `json:"code"`
	Title           string             `json:"title"`
	ComingSoon      bool               `json:"comingSoon"`
	Levels          []AdminLevelFunnel `json:"levels"`
	ExamPassedUsers int                `json:"examPassedUsers"`
}

type AdminLevelFunnel struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	PassedUsers int    `json:"passedUsers"`
}

type AdminCourseUserResult struct {
	Items    []AdminCourseUserRow `json:"items"`
	Total    int                  `json:"total"`
	Page     int                  `json:"page"`
	PageSize int                  `json:"pageSize"`
}

type AdminCourseUserRow struct {
	UserID      string `json:"userId"`
	DisplayName string `json:"displayName"`
	Email       string `json:"email"`
	CoursePosition
	Attempts       int       `json:"attempts"`
	LastActivityAt time.Time `json:"lastActivityAt"`
}

// CoursePosition: ringkasan "udah sampe mana".
type CoursePosition struct {
	LevelsPassed     int  `json:"levelsPassed"`
	TotalLevels      int  `json:"totalLevels"`
	ModulesCompleted int  `json:"modulesCompleted"`
	TotalModules     int  `json:"totalModules"` // module yang udah ada levelnya
	Finished         bool `json:"finished"`     // semua module selesai
	// Langkah berikutnya yang lagi dikerjain: level yang belum lulus, atau
	// final test kalau semua level module itu udah lulus. Kosong kalau Finished.
	CurrentModule string `json:"currentModule,omitempty"`
	CurrentStep   string `json:"currentStep,omitempty"`
	CurrentIsExam bool   `json:"currentIsExam,omitempty"`
}

func position(modules []moduleState) CoursePosition {
	var p CoursePosition
	for _, m := range modules {
		if m.ComingSoon {
			continue
		}
		p.TotalModules++
		p.TotalLevels += len(m.Levels)
		if m.Completed {
			p.ModulesCompleted++
		}
		for _, l := range m.Levels {
			if l.Passed {
				p.LevelsPassed++
			}
		}
		if p.CurrentModule != "" || m.Completed || m.Locked {
			continue
		}
		p.CurrentModule = m.Title
		p.CurrentStep, p.CurrentIsExam = "Final test", true
		for _, l := range m.Levels {
			if !l.Passed {
				p.CurrentStep, p.CurrentIsExam = l.Title, false
				break
			}
		}
	}
	p.Finished = p.TotalModules > 0 && p.ModulesCompleted == p.TotalModules
	return p
}

// cloneModules: salinan struktur biar applyLocks per user gak saling timpa.
func cloneModules(src []moduleState) []moduleState {
	out := make([]moduleState, len(src))
	for i, m := range src {
		out[i] = m
		out[i].Levels = append([]LevelItem(nil), m.Levels...)
	}
	return out
}

// userProgress: level yang lulus + final test yang lulus, buat banyak user
// sekaligus (1 query masing-masing).
func (s *Service) userProgress(ctx context.Context, userIDs []string) (map[string]map[string]bool, map[string]map[string]bool, error) {
	levels := map[string]map[string]bool{}
	exams := map[string]map[string]bool{}
	rows, err := s.db.Query(ctx, `
		SELECT DISTINCT user_id, level_id, module_id FROM course_attempts
		WHERE user_id = ANY($1::uuid[]) AND passed
	`, userIDs)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var userID, moduleID string
		var levelID *string
		if err := rows.Scan(&userID, &levelID, &moduleID); err != nil {
			return nil, nil, err
		}
		if levelID != nil {
			if levels[userID] == nil {
				levels[userID] = map[string]bool{}
			}
			levels[userID][*levelID] = true
		} else {
			if exams[userID] == nil {
				exams[userID] = map[string]bool{}
			}
			exams[userID][moduleID] = true
		}
	}
	return levels, exams, rows.Err()
}

func positionFor(base []moduleState, byID map[string]int, passedLevels, passedExams map[string]bool) []moduleState {
	modules := cloneModules(base)
	for moduleID := range passedExams {
		if i, ok := byID[moduleID]; ok {
			modules[i].Exam.Passed = true
		}
	}
	applyLocks(modules, passedLevels)
	return modules
}

// AdminProgress: funnel per level/final test + daftar murid yang udah mulai
// course, urut aktivitas terakhir. search nyocokin nama/email.
func (s *Service) AdminProgress(ctx context.Context, page, pageSize int, search string) (*AdminCourseProgress, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	base, byID, err := s.structure(ctx)
	if err != nil {
		return nil, err
	}

	res := &AdminCourseProgress{Modules: []AdminModuleFunnel{}}
	levelIdx := map[string][2]int{}
	for i, m := range base {
		f := AdminModuleFunnel{Code: m.Code, Title: m.Title, ComingSoon: len(m.Levels) == 0, Levels: []AdminLevelFunnel{}}
		for j, l := range m.Levels {
			f.Levels = append(f.Levels, AdminLevelFunnel{ID: l.ID, Title: l.Title})
			levelIdx[l.ID] = [2]int{i, j}
		}
		res.Modules = append(res.Modules, f)
	}

	rows, err := s.db.Query(ctx, `
		SELECT a.level_id, a.module_id, count(DISTINCT a.user_id)
		FROM course_attempts a JOIN users u ON u.id = a.user_id
		WHERE a.passed AND u.role = 'student'
		GROUP BY a.level_id, a.module_id
	`)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var levelID *string
		var moduleID string
		var n int
		if err := rows.Scan(&levelID, &moduleID, &n); err != nil {
			rows.Close()
			return nil, err
		}
		if levelID != nil {
			if ij, ok := levelIdx[*levelID]; ok {
				res.Modules[ij[0]].Levels[ij[1]].PassedUsers = n
			}
		} else if i, ok := byID[moduleID]; ok {
			res.Modules[i].ExamPassedUsers = n
		}
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return nil, err
	}

	// Selesai semua = lulus final test semua module yang udah ada levelnya.
	activeModules := 0
	for _, m := range base {
		if len(m.Levels) > 0 {
			activeModules++
		}
	}
	if err := s.db.QueryRow(ctx, `
		SELECT
			(SELECT count(DISTINCT a.user_id) FROM course_attempts a JOIN users u ON u.id = a.user_id
				WHERE u.role = 'student'),
			(SELECT count(*) FROM (
				SELECT a.user_id FROM course_attempts a
				JOIN users u ON u.id = a.user_id
				JOIN course_modules m ON m.id = a.module_id
				WHERE u.role = 'student' AND a.passed AND a.level_id IS NULL
					AND EXISTS (SELECT 1 FROM course_levels l WHERE l.module_id = m.id)
				GROUP BY a.user_id
				HAVING count(DISTINCT a.module_id) >= $1 AND $1 > 0
			) done)
	`, activeModules).Scan(&res.StartedUsers, &res.CompletedUsers); err != nil {
		return nil, err
	}

	likeSearch := "%" + search + "%"
	if err := s.db.QueryRow(ctx, `
		SELECT count(*) FROM users u
		WHERE u.role = 'student' AND ($1 = '' OR u.display_name ILIKE $2 OR u.email ILIKE $2)
			AND EXISTS (SELECT 1 FROM course_attempts a WHERE a.user_id = u.id)
	`, search, likeSearch).Scan(&res.Users.Total); err != nil {
		return nil, err
	}

	rows, err = s.db.Query(ctx, `
		SELECT u.id, u.display_name, u.email, count(a.id),
			max(COALESCE(a.completed_at, a.started_at)) AS last_at
		FROM users u JOIN course_attempts a ON a.user_id = u.id
		WHERE u.role = 'student' AND ($1 = '' OR u.display_name ILIKE $2 OR u.email ILIKE $2)
		GROUP BY u.id
		ORDER BY last_at DESC
		LIMIT $3 OFFSET $4
	`, search, likeSearch, pageSize, (page-1)*pageSize)
	if err != nil {
		return nil, err
	}
	items := []AdminCourseUserRow{}
	ids := []string{}
	for rows.Next() {
		var it AdminCourseUserRow
		if err := rows.Scan(&it.UserID, &it.DisplayName, &it.Email, &it.Attempts, &it.LastActivityAt); err != nil {
			rows.Close()
			return nil, err
		}
		items = append(items, it)
		ids = append(ids, it.UserID)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return nil, err
	}

	if len(ids) > 0 {
		levels, exams, err := s.userProgress(ctx, ids)
		if err != nil {
			return nil, err
		}
		for i := range items {
			id := items[i].UserID
			items[i].CoursePosition = position(positionFor(base, byID, levels[id], exams[id]))
		}
	}
	res.Users = AdminCourseUserResult{Items: items, Total: res.Users.Total, Page: page, PageSize: pageSize}
	return res, nil
}

// --- Detail per user ---

type AdminUserCourse struct {
	CoursePosition
	Attempts       int                   `json:"attempts"`
	LastActivityAt *time.Time            `json:"lastActivityAt,omitempty"`
	Modules        []AdminModuleProgress `json:"modules"`
	Recent         []AdminCourseAttempt  `json:"recent"`
}

type AdminModuleProgress struct {
	Code       string               `json:"code"`
	Title      string               `json:"title"`
	Locked     bool                 `json:"locked"`
	Completed  bool                 `json:"completed"`
	ComingSoon bool                 `json:"comingSoon"`
	Levels     []AdminLevelProgress `json:"levels"`
	Exam       AdminExamProgress    `json:"exam"`
}

type AdminLevelProgress struct {
	ID            string     `json:"id"`
	Title         string     `json:"title"`
	Locked        bool       `json:"locked"`
	Passed        bool       `json:"passed"`
	Attempts      int        `json:"attempts"` // yang udah disubmit
	BestScore     *int       `json:"bestScore,omitempty"`
	LastAttemptAt *time.Time `json:"lastAttemptAt,omitempty"`
}

type AdminExamProgress struct {
	ExamInfo
	Attempts      int        `json:"attempts"`
	LastAttemptAt *time.Time `json:"lastAttemptAt,omitempty"`
}

type AdminCourseAttempt struct {
	ID           string     `json:"id"`
	ModuleTitle  string     `json:"moduleTitle"`
	LevelTitle   string     `json:"levelTitle,omitempty"` // kosong = final test
	IsExam       bool       `json:"isExam"`
	Total        int        `json:"total"`
	CorrectCount *int       `json:"correctCount,omitempty"`
	ScorePercent *int       `json:"scorePercent,omitempty"`
	Passed       *bool      `json:"passed,omitempty"` // nil = belum disubmit
	DurationMs   *int       `json:"durationMs,omitempty"`
	StartedAt    time.Time  `json:"startedAt"`
	CompletedAt  *time.Time `json:"completedAt,omitempty"`
}

const recentAttemptLimit = 15

func (s *Service) AdminUserCourse(ctx context.Context, userID string) (*AdminUserCourse, error) {
	var exists bool
	if err := s.db.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM users WHERE id = $1)`, userID).Scan(&exists); err != nil {
		return nil, err
	}
	if !exists {
		return nil, ErrNotFound
	}

	modules, err := s.overview(ctx, userID)
	if err != nil {
		return nil, err
	}

	type stat struct {
		n    int
		best *int
		last *time.Time
	}
	levelStats := map[string]stat{}
	examStats := map[string]stat{}
	rows, err := s.db.Query(ctx, `
		SELECT level_id, module_id, count(*), max(score_percent), max(completed_at)
		FROM course_attempts
		WHERE user_id = $1 AND completed_at IS NOT NULL
		GROUP BY level_id, module_id
	`, userID)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var levelID *string
		var moduleID string
		var st stat
		if err := rows.Scan(&levelID, &moduleID, &st.n, &st.best, &st.last); err != nil {
			rows.Close()
			return nil, err
		}
		if levelID != nil {
			levelStats[*levelID] = st
		} else {
			examStats[moduleID] = st
		}
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return nil, err
	}

	out := &AdminUserCourse{
		CoursePosition: position(modules),
		Modules:        []AdminModuleProgress{},
		Recent:         []AdminCourseAttempt{},
	}
	for _, m := range modules {
		mp := AdminModuleProgress{
			Code: m.Code, Title: m.Title, Locked: m.Locked, Completed: m.Completed, ComingSoon: m.ComingSoon,
			Levels: []AdminLevelProgress{},
			Exam:   AdminExamProgress{ExamInfo: m.Exam},
		}
		for _, l := range m.Levels {
			st := levelStats[l.ID]
			mp.Levels = append(mp.Levels, AdminLevelProgress{
				ID: l.ID, Title: l.Title, Locked: l.Locked, Passed: l.Passed,
				Attempts: st.n, BestScore: st.best, LastAttemptAt: st.last,
			})
		}
		st := examStats[m.id]
		mp.Exam.Attempts, mp.Exam.LastAttemptAt = st.n, st.last
		out.Modules = append(out.Modules, mp)
	}

	if err := s.db.QueryRow(ctx, `
		SELECT count(*), max(COALESCE(completed_at, started_at)) FROM course_attempts WHERE user_id = $1
	`, userID).Scan(&out.Attempts, &out.LastActivityAt); err != nil {
		return nil, err
	}

	rows, err = s.db.Query(ctx, `
		SELECT a.id, m.title, COALESCE(l.title, ''), a.level_id IS NULL, a.total_questions,
			a.correct_count, a.score_percent, a.passed, a.duration_ms, a.started_at, a.completed_at
		FROM course_attempts a
		JOIN course_modules m ON m.id = a.module_id
		LEFT JOIN course_levels l ON l.id = a.level_id
		WHERE a.user_id = $1
		ORDER BY a.started_at DESC
		LIMIT $2
	`, userID, recentAttemptLimit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var a AdminCourseAttempt
		if err := rows.Scan(&a.ID, &a.ModuleTitle, &a.LevelTitle, &a.IsExam, &a.Total,
			&a.CorrectCount, &a.ScorePercent, &a.Passed, &a.DurationMs, &a.StartedAt, &a.CompletedAt); err != nil {
			return nil, err
		}
		out.Recent = append(out.Recent, a)
	}
	return out, rows.Err()
}
