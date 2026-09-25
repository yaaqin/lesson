package curriculumsvc

import (
	"context"
	"encoding/json"
	"errors"
	"math/rand/v2"
	"time"

	"github.com/jackc/pgx/v5"
)

// Adventure: mode "jalan terus" lintas jenjang. Soal diambil dari bank soal
// kurikulum (MC + essay, published) jenjang SD/SMP/SMK/Kampus -- puzzle tier
// "umum" gak ikut. 40 checkpoint x 100 soal = 4000 soal.
//
//   - CP 1-10 (1000 soal pertama): urut per blok jenjang sesuai komposisi --
//     soal 1-50 SD, 51-200 SMP, 201-550 SMK, 551-1000 Kampus, masing-masing
//     diambil urut dari awal bank jenjangnya (batch -> challenge -> soal).
//   - CP 11+: gacha, tiap CP komposisinya sama (5 SD, 15 SMP, 35 SMK, 45
//     Kampus) diacak campur. Soal yang udah kepake di progres yang masih
//     berlaku (blok urut + CP gacha yang lulus) gak diambil lagi selama bank
//     masih cukup.
//
// Tiap CP punya jatah salah AdventureMaxFails (salah atau waktu habis). Salah
// ke-(jatah+1) = CP gagal, ulang dari soal pertama CP itu. Jawaban dicek per
// soal di server, jadi kunci jawaban gak pernah dikirim ke klien.
const (
	AdventureTotalCheckpoints       = 40
	AdventureQuestionsPerCheckpoint = 100
	AdventureMaxFails               = 5
	adventureOrderedCheckpoints     = 10

	// Toleransi latency jaringan + jeda animasi feedback di klien sebelum
	// jawaban dianggap lewat batas waktu.
	adventureAnswerGrace = 5 * time.Second
	// Waktu per soal yang diambil dari challenge ujian = total waktu ujian
	// dibagi jumlah soalnya, dengan batas bawah ini.
	adventureMinSecondsPerQuestion = 10
)

// adventureComposition: jatah soal per jenjang tiap 100 soal (urutan = urutan
// blok di CP 1-10).
var adventureComposition = []struct {
	TierCode string
	Count    int
}{
	{"sd", 5},
	{"smp", 15},
	{"smk", 35},
	{"kampus", 45},
}

var (
	ErrAdventureCompleted     = errors.New("adventure udah tamat")
	ErrAdventureIndexMismatch = errors.New("index soal gak sesuai posisi attempt")
	ErrAdventureInvalidTarget = errors.New("checkpoint tujuan gak valid")
)

const (
	AdventureStatusInProgress = "in_progress"
	AdventureStatusPassed     = "passed"
	AdventureStatusFailed     = "failed"
)

type adventureBankQuestion struct {
	ID                 string
	TierCode           string
	Type               string
	Prompt             string
	Options            []QuestionOption
	CorrectAnswerValue *float64
	TimeLimitSeconds   int
}

// adventureSnapshotQuestion: format kolom questions_snapshot (ada kuncinya).
type adventureSnapshotQuestion struct {
	ID                 string           `json:"id"`
	TierCode           string           `json:"tierCode"`
	Type               string           `json:"type"`
	Prompt             string           `json:"prompt"`
	Options            []QuestionOption `json:"options,omitempty"`
	CorrectAnswerValue *float64         `json:"correctAnswerValue,omitempty"`
	TimeLimitSeconds   int              `json:"timeLimitSeconds"`
}

// AdventureQuestion: yang dikirim ke klien -- opsi cuma nilainya, tanpa
// penanda benar, dan essay tanpa correctAnswerValue.
type AdventureQuestion struct {
	ID               string    `json:"id"`
	TierCode         string    `json:"tierCode"`
	Type             string    `json:"type"`
	Prompt           string    `json:"prompt"`
	Options          []float64 `json:"options,omitempty"`
	TimeLimitSeconds int       `json:"timeLimitSeconds"`
}

type AdventureHistoryItem struct {
	AttemptID       string    `json:"attemptId"`
	CheckpointNo    int       `json:"checkpointNo"`
	DurationSeconds int       `json:"durationSeconds"`
	CorrectCount    int       `json:"correctCount"`
	FailsRemaining  int       `json:"failsRemaining"`
	FailedAttempts  int       `json:"failedAttempts"`
	CompletedAt     time.Time `json:"completedAt"`
	RolledBack      bool      `json:"rolledBack"`
}

type AdventureState struct {
	CurrentCheckpoint      int                    `json:"currentCheckpoint"`
	TotalCheckpoints       int                    `json:"totalCheckpoints"`
	QuestionsPerCheckpoint int                    `json:"questionsPerCheckpoint"`
	MaxFails               int                    `json:"maxFails"`
	Completed              bool                   `json:"completed"`
	History                []AdventureHistoryItem `json:"history"`
}

type AdventureStartResult struct {
	AttemptID        string              `json:"attemptId"`
	CheckpointNo     int                 `json:"checkpointNo"`
	TotalCheckpoints int                 `json:"totalCheckpoints"`
	QuestionOffset   int                 `json:"questionOffset"`
	MaxFails         int                 `json:"maxFails"`
	Questions        []AdventureQuestion `json:"questions"`
}

type AdventureAnswerResult struct {
	Correct        bool   `json:"correct"`
	TimedOut       bool   `json:"timedOut"`
	FailsUsed      int    `json:"failsUsed"`
	FailsRemaining int    `json:"failsRemaining"`
	CorrectCount   int    `json:"correctCount"`
	Status         string `json:"status"`
	NextIndex      int    `json:"nextIndex"`
	// Keisi kalau Status passed/failed.
	DurationSeconds    int  `json:"durationSeconds,omitempty"`
	NextCheckpoint     int  `json:"nextCheckpoint,omitempty"`
	AdventureCompleted bool `json:"adventureCompleted,omitempty"`
}

func (s *Service) ensureAdventureProgress(ctx context.Context, q pgx.Tx, userID string, forUpdate bool) (current int, completed bool, err error) {
	if _, err := q.Exec(ctx, `
		INSERT INTO adventure_progress (user_id) VALUES ($1) ON CONFLICT (user_id) DO NOTHING
	`, userID); err != nil {
		return 0, false, err
	}
	sql := `SELECT current_checkpoint, completed_at IS NOT NULL FROM adventure_progress WHERE user_id = $1`
	if forUpdate {
		sql += ` FOR UPDATE`
	}
	err = q.QueryRow(ctx, sql, userID).Scan(&current, &completed)
	return current, completed, err
}

func (s *Service) GetAdventure(ctx context.Context, userID string) (*AdventureState, error) {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	current, completed, err := s.ensureAdventureProgress(ctx, tx, userID, false)
	if err != nil {
		return nil, err
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	// failed_attempts = berapa kali CP itu gagal sebelum attempt lulus ini
	// (sejak attempt lulus sebelumnya di CP yang sama, kalau ada).
	rows, err := s.db.Query(ctx, `
		SELECT
			a.id, a.checkpoint_no, a.correct_count, a.max_fails - a.fails_used,
			a.started_at, a.completed_at, a.rolled_back_at IS NOT NULL,
			(
				SELECT count(*) FROM adventure_checkpoint_attempts f
				WHERE f.user_id = a.user_id AND f.checkpoint_no = a.checkpoint_no
					AND f.status = 'failed' AND f.completed_at < a.completed_at
					AND NOT EXISTS (
						SELECT 1 FROM adventure_checkpoint_attempts p
						WHERE p.user_id = a.user_id AND p.checkpoint_no = a.checkpoint_no
							AND p.status = 'passed' AND p.completed_at > f.completed_at
							AND p.completed_at < a.completed_at
					)
			)
		FROM adventure_checkpoint_attempts a
		WHERE a.user_id = $1 AND a.status = 'passed'
		ORDER BY a.completed_at DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	history := []AdventureHistoryItem{}
	for rows.Next() {
		var (
			h         AdventureHistoryItem
			startedAt time.Time
		)
		if err := rows.Scan(&h.AttemptID, &h.CheckpointNo, &h.CorrectCount, &h.FailsRemaining,
			&startedAt, &h.CompletedAt, &h.RolledBack, &h.FailedAttempts); err != nil {
			return nil, err
		}
		h.DurationSeconds = int(h.CompletedAt.Sub(startedAt).Seconds())
		history = append(history, h)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return &AdventureState{
		CurrentCheckpoint:      current,
		TotalCheckpoints:       AdventureTotalCheckpoints,
		QuestionsPerCheckpoint: AdventureQuestionsPerCheckpoint,
		MaxFails:               AdventureMaxFails,
		Completed:              completed,
		History:                history,
	}, nil
}

// StartAdventureCheckpoint mulai (atau mulai ulang) CP user sekarang. Attempt
// in_progress yang masih ada (user keluar di tengah CP) di-abandon -- lanjutnya
// selalu dari soal pertama CP.
func (s *Service) StartAdventureCheckpoint(ctx context.Context, userID string) (*AdventureStartResult, error) {
	bank, err := s.loadAdventureBank(ctx)
	if err != nil {
		return nil, err
	}

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	checkpoint, completed, err := s.ensureAdventureProgress(ctx, tx, userID, true)
	if err != nil {
		return nil, err
	}
	if completed || checkpoint > AdventureTotalCheckpoints {
		return nil, ErrAdventureCompleted
	}

	if _, err := tx.Exec(ctx, `
		UPDATE adventure_checkpoint_attempts SET status = 'abandoned', completed_at = now()
		WHERE user_id = $1 AND status = 'in_progress'
	`, userID); err != nil {
		return nil, err
	}

	var picked []adventureBankQuestion
	if checkpoint <= adventureOrderedCheckpoints {
		picked, err = pickOrderedCheckpoint(bank, checkpoint)
	} else {
		var used map[string]bool
		used, err = s.adventureUsedQuestionIDs(ctx, tx, userID, bank)
		if err == nil {
			picked, err = pickGachaCheckpoint(bank, used)
		}
	}
	if err != nil {
		return nil, err
	}

	snapshot := make([]adventureSnapshotQuestion, len(picked))
	questions := make([]AdventureQuestion, len(picked))
	for i, q := range picked {
		opts := append([]QuestionOption{}, q.Options...)
		rand.Shuffle(len(opts), func(a, b int) { opts[a], opts[b] = opts[b], opts[a] })
		values := make([]float64, len(opts))
		for j, o := range opts {
			values[j] = o.Value
		}
		snapshot[i] = adventureSnapshotQuestion{
			ID: q.ID, TierCode: q.TierCode, Type: q.Type, Prompt: q.Prompt,
			Options: opts, CorrectAnswerValue: q.CorrectAnswerValue, TimeLimitSeconds: q.TimeLimitSeconds,
		}
		questions[i] = AdventureQuestion{
			ID: q.ID, TierCode: q.TierCode, Type: q.Type, Prompt: q.Prompt,
			Options: values, TimeLimitSeconds: q.TimeLimitSeconds,
		}
	}

	snapshotJSON, err := json.Marshal(snapshot)
	if err != nil {
		return nil, err
	}

	var attemptID string
	if err := tx.QueryRow(ctx, `
		INSERT INTO adventure_checkpoint_attempts (user_id, checkpoint_no, questions_snapshot, max_fails)
		VALUES ($1, $2, $3, $4)
		RETURNING id
	`, userID, checkpoint, snapshotJSON, AdventureMaxFails).Scan(&attemptID); err != nil {
		return nil, err
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return &AdventureStartResult{
		AttemptID:        attemptID,
		CheckpointNo:     checkpoint,
		TotalCheckpoints: AdventureTotalCheckpoints,
		QuestionOffset:   (checkpoint - 1) * AdventureQuestionsPerCheckpoint,
		MaxFails:         AdventureMaxFails,
		Questions:        questions,
	}, nil
}

// AnswerAdventureQuestion: jawab 1 soal. selectedValue nil = waktu habis di
// klien. questionIndex wajib sama dengan posisi attempt (nolak jawaban dobel /
// lompat soal).
func (s *Service) AnswerAdventureQuestion(ctx context.Context, userID, attemptID string, questionIndex int, selectedValue *float64) (*AdventureAnswerResult, error) {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	var (
		checkpoint, currentIndex, correctCount, failsUsed, maxFails int
		status                                                      string
		snapshotRaw                                                 []byte
		questionStartedAt, startedAt                                time.Time
	)
	err = tx.QueryRow(ctx, `
		SELECT checkpoint_no, status, questions_snapshot, current_index, correct_count,
			fails_used, max_fails, question_started_at, started_at
		FROM adventure_checkpoint_attempts
		WHERE id = $1 AND user_id = $2
		FOR UPDATE
	`, attemptID, userID).Scan(&checkpoint, &status, &snapshotRaw, &currentIndex, &correctCount,
		&failsUsed, &maxFails, &questionStartedAt, &startedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrAttemptNotFound
		}
		return nil, err
	}
	if status != AdventureStatusInProgress {
		return nil, ErrAttemptFinished
	}
	if questionIndex != currentIndex {
		return nil, ErrAdventureIndexMismatch
	}

	var snapshot []adventureSnapshotQuestion
	if err := json.Unmarshal(snapshotRaw, &snapshot); err != nil {
		return nil, err
	}
	if currentIndex >= len(snapshot) {
		return nil, ErrAttemptFinished
	}
	q := snapshot[currentIndex]

	now := time.Now()
	limit := time.Duration(q.TimeLimitSeconds)*time.Second + adventureAnswerGrace
	timedOut := selectedValue == nil || now.Sub(questionStartedAt) > limit
	correct := !timedOut && adventureAnswerCorrect(q, *selectedValue)

	if correct {
		correctCount++
	} else {
		failsUsed++
	}

	res := &AdventureAnswerResult{
		Correct:      correct,
		TimedOut:     timedOut,
		FailsUsed:    failsUsed,
		CorrectCount: correctCount,
		NextIndex:    currentIndex + 1,
	}
	res.FailsRemaining = max(maxFails-failsUsed, 0)

	switch {
	case failsUsed > maxFails:
		res.Status = AdventureStatusFailed
	case currentIndex+1 >= len(snapshot):
		res.Status = AdventureStatusPassed
	default:
		res.Status = AdventureStatusInProgress
	}

	if res.Status == AdventureStatusInProgress {
		if _, err := tx.Exec(ctx, `
			UPDATE adventure_checkpoint_attempts
			SET current_index = $1, correct_count = $2, fails_used = $3, question_started_at = now()
			WHERE id = $4
		`, currentIndex+1, correctCount, failsUsed, attemptID); err != nil {
			return nil, err
		}
	} else {
		if _, err := tx.Exec(ctx, `
			UPDATE adventure_checkpoint_attempts
			SET status = $1, current_index = $2, correct_count = $3, fails_used = $4, completed_at = now()
			WHERE id = $5
		`, res.Status, currentIndex+1, correctCount, failsUsed, attemptID); err != nil {
			return nil, err
		}
		res.DurationSeconds = int(now.Sub(startedAt).Seconds())
	}

	if res.Status == AdventureStatusPassed {
		// Guard current_checkpoint = checkpoint: kalau user sempat rollback di
		// tab lain, attempt ini udah ke-abandon duluan, jadi gak nyampe sini.
		completedNow := checkpoint >= AdventureTotalCheckpoints
		if _, err := tx.Exec(ctx, `
			UPDATE adventure_progress
			SET current_checkpoint = $1,
				completed_at = CASE WHEN $2 THEN now() ELSE NULL END,
				updated_at = now()
			WHERE user_id = $3 AND current_checkpoint = $4
		`, checkpoint+1, completedNow, userID, checkpoint); err != nil {
			return nil, err
		}
		res.NextCheckpoint = checkpoint + 1
		res.AdventureCompleted = completedNow
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return res, nil
}

// RollbackAdventure: balik ke CP sebelumnya. Progres di CP >= target dipotong
// (attempt lulusnya ditandai rolled_back, tetap muncul di history), user harus
// ngerjain ulang dari CP target.
func (s *Service) RollbackAdventure(ctx context.Context, userID string, targetCheckpoint int) (*AdventureState, error) {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	current, _, err := s.ensureAdventureProgress(ctx, tx, userID, true)
	if err != nil {
		return nil, err
	}
	if targetCheckpoint < 1 || targetCheckpoint >= current {
		return nil, ErrAdventureInvalidTarget
	}

	if _, err := tx.Exec(ctx, `
		UPDATE adventure_checkpoint_attempts SET rolled_back_at = now()
		WHERE user_id = $1 AND status = 'passed' AND checkpoint_no >= $2 AND rolled_back_at IS NULL
	`, userID, targetCheckpoint); err != nil {
		return nil, err
	}
	if _, err := tx.Exec(ctx, `
		UPDATE adventure_checkpoint_attempts SET status = 'abandoned', completed_at = now()
		WHERE user_id = $1 AND status = 'in_progress'
	`, userID); err != nil {
		return nil, err
	}
	if _, err := tx.Exec(ctx, `
		UPDATE adventure_progress
		SET current_checkpoint = $1, completed_at = NULL, updated_at = now()
		WHERE user_id = $2
	`, targetCheckpoint, userID); err != nil {
		return nil, err
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return s.GetAdventure(ctx, userID)
}

func adventureAnswerCorrect(q adventureSnapshotQuestion, selected float64) bool {
	if q.Type == QuestionTypeEssayNumeric {
		return q.CorrectAnswerValue != nil && floatEquals(*q.CorrectAnswerValue, selected)
	}
	for _, opt := range q.Options {
		if opt.IsCorrect && floatEquals(opt.Value, selected) {
			return true
		}
	}
	return false
}

// loadAdventureBank: bank soal per jenjang, urut batch -> challenge -> soal.
// Waktu per soal ikut challenge asalnya; challenge ujian (waktunya total
// sesi) dibagi rata ke jumlah soal ujiannya.
func (s *Service) loadAdventureBank(ctx context.Context) (map[string][]adventureBankQuestion, error) {
	tierCodes := make([]string, len(adventureComposition))
	for i, c := range adventureComposition {
		tierCodes[i] = c.TierCode
	}

	rows, err := s.db.Query(ctx, `
		SELECT
			t.code, q.id, q.question_type, q.prompt, q.correct_answer_value,
			c.is_exam, c.time_limit_seconds, c.question_count_required,
			qo.option_value, qo.is_correct
		FROM questions q
		JOIN challenges c ON c.id = q.challenge_id
		JOIN batches b ON b.id = c.batch_id
		JOIN tiers t ON t.id = b.tier_id
		LEFT JOIN question_options qo ON qo.question_id = q.id
		WHERE t.code = ANY($1) AND q.status = 'published'
			AND q.question_type IN ('multiple_choice', 'essay_numeric')
		ORDER BY t.order_index, b.order_index, c.order_index, q.created_at, q.id
	`, tierCodes)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	bank := map[string][]adventureBankQuestion{}
	indexByID := map[string]int{}
	for rows.Next() {
		var (
			tierCode, qID, qType, prompt     string
			correctAnswer                    *float64
			isExam                           bool
			timeLimit, questionCountRequired int
			optValue                         *float64
			isCorrect                        *bool
		)
		if err := rows.Scan(&tierCode, &qID, &qType, &prompt, &correctAnswer,
			&isExam, &timeLimit, &questionCountRequired, &optValue, &isCorrect); err != nil {
			return nil, err
		}
		idx, ok := indexByID[qID]
		if !ok {
			perQuestion := timeLimit
			if isExam && questionCountRequired > 0 {
				perQuestion = (timeLimit + questionCountRequired - 1) / questionCountRequired
			}
			bank[tierCode] = append(bank[tierCode], adventureBankQuestion{
				ID: qID, TierCode: tierCode, Type: qType, Prompt: prompt,
				CorrectAnswerValue: correctAnswer,
				TimeLimitSeconds:   max(perQuestion, adventureMinSecondsPerQuestion),
			})
			idx = len(bank[tierCode]) - 1
			indexByID[qID] = idx
		}
		if optValue != nil && isCorrect != nil {
			q := &bank[tierCode][idx]
			q.Options = append(q.Options, QuestionOption{Value: *optValue, IsCorrect: *isCorrect})
		}
	}
	return bank, rows.Err()
}

// orderedBlock: soal blok urut CP 1-10 buat 1 jenjang (share x 10 soal dari
// awal bank jenjang itu).
func orderedBlock(bank map[string][]adventureBankQuestion, tierCode string, count int) ([]adventureBankQuestion, error) {
	n := count * adventureOrderedCheckpoints
	if len(bank[tierCode]) < n {
		return nil, ErrNotEnoughBank
	}
	return bank[tierCode][:n], nil
}

func pickOrderedCheckpoint(bank map[string][]adventureBankQuestion, checkpoint int) ([]adventureBankQuestion, error) {
	var all []adventureBankQuestion
	for _, c := range adventureComposition {
		block, err := orderedBlock(bank, c.TierCode, c.Count)
		if err != nil {
			return nil, err
		}
		all = append(all, block...)
	}
	start := (checkpoint - 1) * AdventureQuestionsPerCheckpoint
	return all[start : start+AdventureQuestionsPerCheckpoint], nil
}

// adventureUsedQuestionIDs: soal di blok urut + soal CP gacha yang lulus dan
// belum di-rollback -- dihindari biar soal gacha gak ngulang.
func (s *Service) adventureUsedQuestionIDs(ctx context.Context, tx pgx.Tx, userID string, bank map[string][]adventureBankQuestion) (map[string]bool, error) {
	used := map[string]bool{}
	for _, c := range adventureComposition {
		block, err := orderedBlock(bank, c.TierCode, c.Count)
		if err != nil {
			return nil, err
		}
		for _, q := range block {
			used[q.ID] = true
		}
	}

	rows, err := tx.Query(ctx, `
		SELECT elem->>'id'
		FROM adventure_checkpoint_attempts a, jsonb_array_elements(a.questions_snapshot) elem
		WHERE a.user_id = $1 AND a.status = 'passed' AND a.rolled_back_at IS NULL
			AND a.checkpoint_no > $2
	`, userID, adventureOrderedCheckpoints)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		used[id] = true
	}
	return used, rows.Err()
}

// pickGachaCheckpoint: ambil acak sesuai komposisi, prioritas soal yang belum
// kepake; kalau sisa bank jenjang itu kurang, sisanya diambil dari soal yang
// udah pernah kepake.
func pickGachaCheckpoint(bank map[string][]adventureBankQuestion, used map[string]bool) ([]adventureBankQuestion, error) {
	var picked []adventureBankQuestion
	for _, c := range adventureComposition {
		pool := bank[c.TierCode]
		if len(pool) < c.Count {
			return nil, ErrNotEnoughBank
		}
		var fresh, reused []adventureBankQuestion
		for _, q := range pool {
			if used[q.ID] {
				reused = append(reused, q)
			} else {
				fresh = append(fresh, q)
			}
		}
		rand.Shuffle(len(fresh), func(i, j int) { fresh[i], fresh[j] = fresh[j], fresh[i] })
		rand.Shuffle(len(reused), func(i, j int) { reused[i], reused[j] = reused[j], reused[i] })
		picked = append(picked, append(fresh, reused...)[:c.Count]...)
	}
	rand.Shuffle(len(picked), func(i, j int) { picked[i], picked[j] = picked[j], picked[i] })
	return picked, nil
}
