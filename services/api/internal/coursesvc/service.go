package coursesvc

import (
	"context"
	"encoding/json"
	"errors"
	"math"
	"math/rand/v2"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Course Coding Dasar (migrations/0016). Aturan kunci:
//   - Module pertama selalu kebuka, module berikutnya kebuka kalau final test
//     module sebelumnya udah lulus.
//   - Level pertama di module yang kebuka selalu kebuka, level berikutnya
//     kebuka kalau level sebelumnya lulus (100% bener).
//   - Final test kebuka kalau semua level di module itu lulus.
// Gak ada nyawa: latihan & final test boleh diulang sepuasnya, soal diacak
// ulang dari bank tiap mulai.

type Service struct {
	db *pgxpool.Pool
}

func NewService(db *pgxpool.Pool) *Service {
	return &Service{db: db}
}

// overview: ModuleItem + id internal yang dibutuhin buat cek kunci.
type moduleState struct {
	ModuleItem
	id       string
	levelIDs []string
}

// structure: module + level + ukuran bank final test, belum ada progress
// user (kunci/lulus diisi applyLocks). byID: module id -> indeks.
func (s *Service) structure(ctx context.Context) ([]moduleState, map[string]int, error) {
	rows, err := s.db.Query(ctx, `
		SELECT id, code, title, description, exam_question_count, exam_time_limit_seconds, exam_pass_percent
		FROM course_modules ORDER BY order_index
	`)
	if err != nil {
		return nil, nil, err
	}
	modules := []moduleState{}
	byID := map[string]int{}
	for rows.Next() {
		var m moduleState
		if err := rows.Scan(&m.id, &m.Code, &m.Title, &m.Description,
			&m.Exam.QuestionCount, &m.Exam.TimeLimitSeconds, &m.Exam.PassPercent); err != nil {
			rows.Close()
			return nil, nil, err
		}
		m.Levels = []LevelItem{}
		byID[m.id] = len(modules)
		modules = append(modules, m)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return nil, nil, err
	}

	rows, err = s.db.Query(ctx, `
		SELECT id, module_id, code, title, summary, question_count
		FROM course_levels ORDER BY order_index
	`)
	if err != nil {
		return nil, nil, err
	}
	for rows.Next() {
		var l LevelItem
		var moduleID string
		if err := rows.Scan(&l.ID, &moduleID, &l.Code, &l.Title, &l.Summary, &l.QuestionCount); err != nil {
			rows.Close()
			return nil, nil, err
		}
		if i, ok := byID[moduleID]; ok {
			modules[i].Levels = append(modules[i].Levels, l)
			modules[i].levelIDs = append(modules[i].levelIDs, l.ID)
		}
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return nil, nil, err
	}

	rows, err = s.db.Query(ctx, `
		SELECT module_id, count(*) FROM course_questions WHERE in_exam GROUP BY module_id
	`)
	if err != nil {
		return nil, nil, err
	}
	for rows.Next() {
		var moduleID string
		var n int
		if err := rows.Scan(&moduleID, &n); err != nil {
			rows.Close()
			return nil, nil, err
		}
		if i, ok := byID[moduleID]; ok {
			modules[i].Exam.BankSize = n
		}
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return nil, nil, err
	}

	return modules, byID, nil
}

func (s *Service) overview(ctx context.Context, userID string) ([]moduleState, error) {
	modules, byID, err := s.structure(ctx)
	if err != nil {
		return nil, err
	}

	passedLevels := map[string]bool{}
	rows, err := s.db.Query(ctx, `
		SELECT DISTINCT level_id FROM course_attempts
		WHERE user_id = $1 AND passed AND level_id IS NOT NULL
	`, userID)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			rows.Close()
			return nil, err
		}
		passedLevels[id] = true
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return nil, err
	}

	rows, err = s.db.Query(ctx, `
		SELECT module_id, max(score_percent), bool_or(passed), min(duration_ms) FILTER (WHERE passed)
		FROM course_attempts
		WHERE user_id = $1 AND level_id IS NULL AND completed_at IS NOT NULL
		GROUP BY module_id
	`, userID)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var moduleID string
		var best, bestTime *int
		var passed *bool
		if err := rows.Scan(&moduleID, &best, &passed, &bestTime); err != nil {
			rows.Close()
			return nil, err
		}
		if i, ok := byID[moduleID]; ok {
			modules[i].Exam.BestScore = best
			modules[i].Exam.BestTimeMs = bestTime
			modules[i].Exam.Passed = passed != nil && *passed
		}
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return nil, err
	}

	applyLocks(modules, passedLevels)
	return modules, nil
}

// applyLocks: isi Locked/Passed/Completed dari level yang udah lulus (Exam.Passed
// udah diisi duluan) -- aturan kunci lihat komentar di atas Service.
func applyLocks(modules []moduleState, passedLevels map[string]bool) {
	prevModuleDone := true
	for i := range modules {
		m := &modules[i]
		m.Locked = !prevModuleDone
		m.ComingSoon = len(m.Levels) == 0
		allPassed := true
		prevLevelPassed := true
		for j := range m.Levels {
			l := &m.Levels[j]
			l.Passed = passedLevels[l.ID]
			l.Locked = m.Locked || !prevLevelPassed
			prevLevelPassed = l.Passed
			allPassed = allPassed && l.Passed
		}
		m.Exam.Locked = m.Locked || m.ComingSoon || !allPassed || m.Exam.BankSize == 0
		m.Completed = m.Exam.Passed
		prevModuleDone = m.Completed
	}
}

func (s *Service) ListModules(ctx context.Context, userID string) ([]ModuleItem, error) {
	states, err := s.overview(ctx, userID)
	if err != nil {
		return nil, err
	}
	out := make([]ModuleItem, len(states))
	for i, m := range states {
		out[i] = m.ModuleItem
	}
	return out, nil
}

func findLevel(modules []moduleState, levelID string) (*moduleState, int) {
	for i := range modules {
		for j := range modules[i].Levels {
			if modules[i].Levels[j].ID == levelID {
				return &modules[i], j
			}
		}
	}
	return nil, -1
}

func findModule(modules []moduleState, code string) *moduleState {
	for i := range modules {
		if modules[i].Code == code {
			return &modules[i]
		}
	}
	return nil
}

func (s *Service) GetLevel(ctx context.Context, userID, levelID string) (*LevelDetail, error) {
	modules, err := s.overview(ctx, userID)
	if err != nil {
		return nil, err
	}
	m, j := findLevel(modules, levelID)
	if m == nil {
		return nil, ErrNotFound
	}
	l := m.Levels[j]
	if l.Locked {
		return nil, ErrLocked
	}
	d := &LevelDetail{
		ID: l.ID, Code: l.Code, Title: l.Title, Summary: l.Summary,
		ModuleCode: m.Code, ModuleTitle: m.Title, LevelNumber: j + 1,
		QuestionCount: l.QuestionCount, Passed: l.Passed,
	}
	if err := s.db.QueryRow(ctx, `SELECT material FROM course_levels WHERE id = $1`, levelID).Scan(&d.Material); err != nil {
		return nil, err
	}
	return d, nil
}

func (s *Service) loadQuestions(ctx context.Context, where string, arg string) ([]snapshotQuestion, error) {
	rows, err := s.db.Query(ctx, `
		SELECT id, kind, prompt, coalesce(code, ''), explanation, payload
		FROM course_questions WHERE `+where+`
		ORDER BY order_index
	`, arg)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	bank := []snapshotQuestion{}
	for rows.Next() {
		var q snapshotQuestion
		var raw []byte
		if err := rows.Scan(&q.ID, &q.Kind, &q.Prompt, &q.Code, &q.Explanation, &raw); err != nil {
			return nil, err
		}
		if err := json.Unmarshal(raw, &q.Payload); err != nil {
			return nil, err
		}
		bank = append(bank, q)
	}
	return bank, rows.Err()
}

// pick: ambil n soal acak dari bank, opsi PG diacak juga (kunci ikut
// dipindah).
func pick(bank []snapshotQuestion, n int) []snapshotQuestion {
	rand.Shuffle(len(bank), func(i, j int) { bank[i], bank[j] = bank[j], bank[i] })
	if n > len(bank) {
		n = len(bank)
	}
	picked := bank[:n]
	for i := range picked {
		p := &picked[i].Payload
		if picked[i].Kind != KindMultipleChoice || len(p.Options) == 0 {
			continue
		}
		correct := p.Options[p.Correct]
		opts := append([]string(nil), p.Options...)
		rand.Shuffle(len(opts), func(a, b int) { opts[a], opts[b] = opts[b], opts[a] })
		p.Options = opts
		for k, o := range opts {
			if o == correct {
				p.Correct = k
			}
		}
	}
	return picked
}

func publicQuestions(snap []snapshotQuestion) []Question {
	out := make([]Question, len(snap))
	for i, q := range snap {
		out[i] = Question{
			ID: q.ID, Kind: q.Kind, Prompt: q.Prompt, Code: q.Code,
			Placeholder: q.Payload.Placeholder, StrictJSON: q.Payload.StrictJSON,
		}
		if q.Kind == KindMultipleChoice {
			out[i].Options = q.Payload.Options
		}
	}
	return out
}

func (s *Service) createAttempt(ctx context.Context, userID, moduleID string, levelID *string, snap []snapshotQuestion) (string, error) {
	raw, err := json.Marshal(snap)
	if err != nil {
		return "", err
	}
	var id string
	err = s.db.QueryRow(ctx, `
		INSERT INTO course_attempts (user_id, module_id, level_id, questions_snapshot, total_questions)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`, userID, moduleID, levelID, raw, len(snap)).Scan(&id)
	return id, err
}

func (s *Service) StartLevel(ctx context.Context, userID, levelID string) (*StartResult, error) {
	modules, err := s.overview(ctx, userID)
	if err != nil {
		return nil, err
	}
	m, j := findLevel(modules, levelID)
	if m == nil {
		return nil, ErrNotFound
	}
	l := m.Levels[j]
	if l.Locked {
		return nil, ErrLocked
	}
	bank, err := s.loadQuestions(ctx, "level_id = $1", levelID)
	if err != nil {
		return nil, err
	}
	if len(bank) == 0 {
		return nil, ErrEmptyBank
	}
	snap := pick(bank, l.QuestionCount)
	attemptID, err := s.createAttempt(ctx, userID, m.id, &levelID, snap)
	if err != nil {
		return nil, err
	}
	return &StartResult{
		AttemptID: attemptID, Title: l.Title, ModuleCode: m.Code,
		PassPercent: 100, Questions: publicQuestions(snap),
	}, nil
}

func (s *Service) StartExam(ctx context.Context, userID, moduleCode string) (*StartResult, error) {
	modules, err := s.overview(ctx, userID)
	if err != nil {
		return nil, err
	}
	m := findModule(modules, moduleCode)
	if m == nil {
		return nil, ErrNotFound
	}
	if m.Exam.Locked {
		return nil, ErrLocked
	}
	bank, err := s.loadQuestions(ctx, "module_id = $1 AND in_exam", m.id)
	if err != nil {
		return nil, err
	}
	if len(bank) == 0 {
		return nil, ErrEmptyBank
	}
	snap := pick(bank, m.Exam.QuestionCount)
	attemptID, err := s.createAttempt(ctx, userID, m.id, nil, snap)
	if err != nil {
		return nil, err
	}
	return &StartResult{
		AttemptID: attemptID, IsExam: true, Title: "Final Test — " + m.Title, ModuleCode: m.Code,
		PassPercent: m.Exam.PassPercent, TimeLimitSeconds: m.Exam.TimeLimitSeconds,
		Questions: publicQuestions(snap),
	}, nil
}

// normalizeTypeAnswer: "  `String`; " -> "string". Awalan "tipe:" / "typeof"
// juga dibuang biar jawaban kayak "tipe: array" tetep kebaca.
func normalizeTypeAnswer(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	s = strings.TrimSuffix(s, ";")
	for _, prefix := range []string{"tipe:", "tipe", "typeof"} {
		s = strings.TrimSpace(strings.TrimPrefix(s, prefix))
	}
	return strings.Trim(s, "\"'` ")
}

func grade(q snapshotQuestion, a *Answer) QuestionResult {
	r := QuestionResult{QuestionID: q.ID, Explanation: q.Explanation}
	p := q.Payload
	switch q.Kind {
	case KindMultipleChoice:
		if p.Correct >= 0 && p.Correct < len(p.Options) {
			r.Expected = p.Options[p.Correct]
		}
		r.Correct = a != nil && a.Choice != nil && *a.Choice == p.Correct
	case KindTypeAnswer:
		if len(p.Accepted) > 0 {
			r.Expected = p.Accepted[0]
		}
		if a != nil {
			got := normalizeTypeAnswer(a.Text)
			for _, acc := range p.Accepted {
				if got == strings.ToLower(acc) {
					r.Correct = true
				}
			}
		}
	case KindWriteValue:
		r.Expected = p.Sample
		if a == nil || strings.TrimSpace(a.Text) == "" {
			r.Feedback = "belum dijawab"
			break
		}
		if p.Schema == nil {
			break
		}
		if err := CheckWriteValue(a.Text, *p.Schema, p.StrictJSON); err != nil {
			r.Feedback = err.Error()
		} else {
			r.Correct = true
		}
	}
	return r
}

func (s *Service) Submit(ctx context.Context, userID, attemptID string, answers []Answer) (*SubmitResult, error) {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	var (
		moduleID    string
		levelID     *string
		rawSnap     []byte
		startedAt   time.Time
		completedAt *time.Time
		passPercent int
		timeLimit   int
	)
	err = tx.QueryRow(ctx, `
		SELECT a.module_id, a.level_id, a.questions_snapshot, a.started_at, a.completed_at,
			m.exam_pass_percent, m.exam_time_limit_seconds
		FROM course_attempts a JOIN course_modules m ON m.id = a.module_id
		WHERE a.id = $1 AND a.user_id = $2
		FOR UPDATE OF a
	`, attemptID, userID).Scan(&moduleID, &levelID, &rawSnap, &startedAt, &completedAt, &passPercent, &timeLimit)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrAttemptNotFound
	}
	if err != nil {
		return nil, err
	}
	if completedAt != nil {
		return nil, ErrAttemptFinished
	}

	var snap []snapshotQuestion
	if err := json.Unmarshal(rawSnap, &snap); err != nil {
		return nil, err
	}
	byID := map[string]*Answer{}
	for i := range answers {
		byID[answers[i].QuestionID] = &answers[i]
	}

	isExam := levelID == nil
	if !isExam {
		passPercent = 100
	}
	res := &SubmitResult{Total: len(snap), PassPercent: passPercent, Results: make([]QuestionResult, len(snap))}
	for i, q := range snap {
		res.Results[i] = grade(q, byID[q.ID])
		if res.Results[i].Correct {
			res.CorrectCount++
		}
	}
	if res.Total > 0 {
		res.ScorePercent = int(math.Floor(float64(res.CorrectCount) * 100 / float64(res.Total)))
	}
	res.DurationMs = int(time.Since(startedAt).Milliseconds())
	res.Passed = res.ScorePercent >= passPercent
	if isExam && res.DurationMs > (timeLimit+examGraceSeconds)*1000 {
		res.TimedOut = true
		res.Passed = false
	}

	rawAnswers, err := json.Marshal(answers)
	if err != nil {
		return nil, err
	}
	if _, err := tx.Exec(ctx, `
		UPDATE course_attempts
		SET answers = $2, correct_count = $3, score_percent = $4, passed = $5,
			duration_ms = $6, completed_at = now()
		WHERE id = $1
	`, attemptID, rawAnswers, res.CorrectCount, res.ScorePercent, res.Passed, res.DurationMs); err != nil {
		return nil, err
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	cfg, err := s.GetConfig(ctx)
	if err != nil {
		cfg = &DefaultConfig
	}
	if isExam {
		applyReview(res, cfg.Exam)
	} else {
		applyReview(res, cfg.Level)
	}

	if res.Passed {
		s.fillNavigation(ctx, userID, moduleID, levelID, res)
	}
	return res, nil
}

// fillNavigation: abis lulus, kasih tau klien ke mana lanjutnya. Gagal baca
// overview gak bikin submit gagal -- cuma navigasinya yang kosong.
func (s *Service) fillNavigation(ctx context.Context, userID, moduleID string, levelID *string, res *SubmitResult) {
	modules, err := s.overview(ctx, userID)
	if err != nil {
		return
	}
	for i, m := range modules {
		if m.id != moduleID {
			continue
		}
		if levelID == nil {
			if i+1 < len(modules) && !modules[i+1].Locked {
				res.NextModuleUnlocked = modules[i+1].Code
			}
			return
		}
		for j, l := range m.Levels {
			if l.ID != *levelID {
				continue
			}
			if j+1 < len(m.Levels) {
				res.NextLevelID = m.Levels[j+1].ID
			} else {
				res.ExamUnlocked = !m.Exam.Locked
			}
			return
		}
	}
}
