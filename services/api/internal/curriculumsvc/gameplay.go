package curriculumsvc

import (
	"context"
	"encoding/json"
	"errors"
	"math"
	"math/rand/v2"
	"time"

	"github.com/jackc/pgx/v5"
)

type bankQuestion struct {
	ID                 string
	Type               string
	Prompt             string
	Options            []QuestionOption
	CorrectAnswerValue *float64
	Puzzle             *PuzzlePayload
	PuzzleSolution     map[string]float64
}

// StartChallenge = "gacha fetch": pastiin nyawa cukup, ambil seluruh bank soal
// published milik challenge ini, acak urutan soal & opsi, simpan snapshot-nya
// sebagai attempt baru (in_progress). Snapshot inilah yang jadi acuan scoring
// saat submit, bukan bank soal live — jadi konsisten walau bank berubah di
// tengah sesi (FSD.md 4.4).
func (s *Service) StartChallenge(ctx context.Context, userID, challengeID string) (*StartResult, error) {
	livesRemaining, err := s.ensureDailyLivesReset(ctx, userID)
	if err != nil {
		return nil, err
	}
	if livesRemaining <= 0 {
		return nil, ErrNoLives
	}

	var (
		name                  string
		isExam                bool
		questionCountRequired int
		passThresholdPercent  int
		timeLimitSeconds      int
		batchID               *string
	)
	err = s.db.QueryRow(ctx, `
		SELECT name, is_exam, question_count_required, pass_threshold_percent, time_limit_seconds, batch_id
		FROM challenges WHERE id = $1
	`, challengeID).Scan(&name, &isExam, &questionCountRequired, &passThresholdPercent, &timeLimitSeconds, &batchID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	if err := s.ensureChallengeUnlocked(ctx, userID, batchID, isExam); err != nil {
		return nil, err
	}

	bank, err := s.loadPublishedBank(ctx, challengeID)
	if err != nil {
		return nil, err
	}
	if len(bank) < questionCountRequired {
		return nil, ErrNotEnoughBank
	}

	rand.Shuffle(len(bank), func(i, j int) { bank[i], bank[j] = bank[j], bank[i] })
	picked := bank[:questionCountRequired]

	snapshot := make([]snapshotQuestion, len(picked))
	questions := make([]SessionQuestion, len(picked))
	for i, q := range picked {
		switch q.Type {
		case QuestionTypeEssayNumeric:
			snapshot[i] = snapshotQuestion{ID: q.ID, Type: q.Type, Prompt: q.Prompt, CorrectAnswerValue: q.CorrectAnswerValue}
			questions[i] = SessionQuestion{ID: q.ID, Type: q.Type, Prompt: q.Prompt, CorrectAnswerValue: q.CorrectAnswerValue}

		case QuestionTypeGridPuzzle:
			snapshot[i] = snapshotQuestion{ID: q.ID, Type: q.Type, Prompt: q.Prompt, Puzzle: q.Puzzle, PuzzleSolution: q.PuzzleSolution}
			questions[i] = SessionQuestion{ID: q.ID, Type: q.Type, Prompt: q.Prompt, Puzzle: q.Puzzle}

		default:
			opts := append([]QuestionOption{}, q.Options...)
			rand.Shuffle(len(opts), func(a, b int) { opts[a], opts[b] = opts[b], opts[a] })

			snapshot[i] = snapshotQuestion{ID: q.ID, Type: q.Type, Prompt: q.Prompt, Options: opts}
			questions[i] = SessionQuestion{ID: q.ID, Type: q.Type, Prompt: q.Prompt, Options: opts}
		}
	}

	snapshotJSON, err := json.Marshal(snapshot)
	if err != nil {
		return nil, err
	}

	var attemptID string
	err = s.db.QueryRow(ctx, `
		INSERT INTO challenge_attempts (user_id, challenge_id, status, questions_snapshot)
		VALUES ($1, $2, 'in_progress', $3)
		RETURNING id
	`, userID, challengeID, snapshotJSON).Scan(&attemptID)
	if err != nil {
		return nil, err
	}

	return &StartResult{
		AttemptID:            attemptID,
		ChallengeName:        name,
		IsExam:               isExam,
		TimeLimitSeconds:     timeLimitSeconds,
		PassThresholdPercent: passThresholdPercent,
		Questions:            questions,
	}, nil
}

// loadPublishedBank ambil semua soal published, MC/essay/grid_puzzle. LEFT
// JOIN (bukan INNER) buat question_options & puzzle_questions karena essay
// gak punya baris question_options, grid_puzzle gak punya baris keduanya --
// INNER JOIN bakal diam-diam ngilangin soal-soal itu dari bank.
func (s *Service) loadPublishedBank(ctx context.Context, challengeID string) ([]bankQuestion, error) {
	rows, err := s.db.Query(ctx, `
		SELECT
			q.id, q.question_type, q.prompt, q.correct_answer_value,
			qo.option_value, qo.is_correct,
			pq.kind, pq.payload
		FROM questions q
		LEFT JOIN question_options qo ON qo.question_id = q.id
		LEFT JOIN puzzle_questions pq ON pq.question_id = q.id
		WHERE q.challenge_id = $1 AND q.status = 'published'
		ORDER BY q.id
	`, challengeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	byID := map[string]*bankQuestion{}
	order := []string{}
	for rows.Next() {
		var (
			qID, qType, prompt string
			correctAnswer      *float64
			optValue           *float64
			isCorrect          *bool
			puzzleKind         *string
			puzzlePayloadRaw   []byte
		)
		if err := rows.Scan(&qID, &qType, &prompt, &correctAnswer, &optValue, &isCorrect, &puzzleKind, &puzzlePayloadRaw); err != nil {
			return nil, err
		}
		q, ok := byID[qID]
		if !ok {
			q = &bankQuestion{ID: qID, Type: qType, Prompt: prompt, CorrectAnswerValue: correctAnswer}
			if puzzleKind != nil {
				puzzle, solution, err := derivePuzzle(*puzzleKind, puzzlePayloadRaw)
				if err != nil {
					return nil, err
				}
				q.Puzzle = puzzle
				q.PuzzleSolution = solution
			}
			byID[qID] = q
			order = append(order, qID)
		}
		if optValue != nil && isCorrect != nil {
			q.Options = append(q.Options, QuestionOption{Value: *optValue, IsCorrect: *isCorrect})
		}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	bank := make([]bankQuestion, 0, len(order))
	for _, id := range order {
		bank = append(bank, *byID[id])
	}
	return bank, nil
}

// SubmitAttempt hitung skor dari snapshot (bukan bank live), tentuin lulus/gagal,
// lalu update nyawa (berkurang kalau gagal) & streak harian (jalan terus baik
// lulus maupun gagal, asal sesi diselesaikan — FSD.md 3.7). securityEvents =
// sisa buffer anti-cheating di klien yang belum sempat ke-flush, disimpen dulu
// sebelum risk score dihitung (lihat security.go).
func (s *Service) SubmitAttempt(ctx context.Context, userID, attemptID string, answers []SubmitAnswer, securityEvents []SecurityEvent) (*SubmitResult, error) {
	if err := validateSecurityEvents(securityEvents); err != nil {
		return nil, err
	}

	var (
		challengeID string
		status      string
		snapshotRaw []byte
	)
	err := s.db.QueryRow(ctx, `
		SELECT challenge_id, status, questions_snapshot
		FROM challenge_attempts
		WHERE id = $1 AND user_id = $2
	`, attemptID, userID).Scan(&challengeID, &status, &snapshotRaw)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrAttemptNotFound
		}
		return nil, err
	}
	if status != "in_progress" {
		return nil, ErrAttemptFinished
	}

	var snapshot []snapshotQuestion
	if err := json.Unmarshal(snapshotRaw, &snapshot); err != nil {
		return nil, err
	}

	if err := s.insertSecurityEvents(ctx, attemptID, securityEvents); err != nil {
		return nil, err
	}
	riskScore, riskLevel, err := s.scoreAttemptRisk(ctx, attemptID)
	if err != nil {
		return nil, err
	}

	var passThresholdPercent int
	if err := s.db.QueryRow(ctx,
		`SELECT pass_threshold_percent FROM challenges WHERE id = $1`, challengeID,
	).Scan(&passThresholdPercent); err != nil {
		return nil, err
	}

	selectedByQuestion := make(map[string]float64, len(answers))
	selectedGridByQuestion := make(map[string]map[string]float64, len(answers))
	for _, a := range answers {
		if a.SelectedValue != nil {
			selectedByQuestion[a.QuestionID] = *a.SelectedValue
		}
		if a.SelectedGrid != nil {
			selectedGridByQuestion[a.QuestionID] = a.SelectedGrid
		}
	}

	correctCount := 0
	for _, q := range snapshot {
		if q.Type == QuestionTypeGridPuzzle {
			grid, answered := selectedGridByQuestion[q.ID]
			if !answered || len(q.PuzzleSolution) == 0 {
				continue
			}
			allMatch := true
			for key, want := range q.PuzzleSolution {
				got, ok := grid[key]
				if !ok || !floatEquals(want, got) {
					allMatch = false
					break
				}
			}
			if allMatch {
				correctCount++
			}
			continue
		}

		selected, answered := selectedByQuestion[q.ID]
		if !answered {
			continue
		}
		if q.Type == QuestionTypeEssayNumeric {
			if q.CorrectAnswerValue != nil && floatEquals(*q.CorrectAnswerValue, selected) {
				correctCount++
			}
			continue
		}
		for _, opt := range q.Options {
			if opt.IsCorrect && floatEquals(opt.Value, selected) {
				correctCount++
				break
			}
		}
	}

	totalQuestions := len(snapshot)
	scorePercent := 0
	if totalQuestions > 0 {
		scorePercent = int(math.Round(float64(correctCount) / float64(totalQuestions) * 100))
	}
	passed := scorePercent >= passThresholdPercent

	newStatus := "failed"
	if passed {
		newStatus = "passed"
	}

	answersJSON, err := json.Marshal(answers)
	if err != nil {
		return nil, err
	}

	if _, err := s.db.Exec(ctx, `
		UPDATE challenge_attempts
		SET status = $1, score_percent = $2, answers_submitted = $3, completed_at = now(),
			risk_score = $4, risk_level = $5, correct_count = $6
		WHERE id = $7
	`, newStatus, scorePercent, answersJSON, riskScore, riskLevel, correctCount, attemptID); err != nil {
		return nil, err
	}

	livesRemaining, err := s.applyLivesOnResult(ctx, userID, passed)
	if err != nil {
		return nil, err
	}

	currentStreak, err := s.applyDailyStreak(ctx, userID)
	if err != nil {
		return nil, err
	}

	return &SubmitResult{
		ScorePercent:   scorePercent,
		Passed:         passed,
		CorrectCount:   correctCount,
		TotalQuestions: totalQuestions,
		LivesRemaining: livesRemaining,
		CurrentStreak:  currentStreak,
		RiskLevel:      riskLevel,
	}, nil
}

func floatEquals(a, b float64) bool {
	const epsilon = 1e-9
	diff := a - b
	if diff < 0 {
		diff = -diff
	}
	return diff < epsilon
}

func truncateToUTCDate(t time.Time) time.Time {
	t = t.UTC()
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
}

// ensureDailyLivesReset: reset nyawa ke 3 kalau reset terakhir bukan hari ini
// (UTC), bikin baris nyawa baru kalau user belum pernah punya (FSD.md 3.6 —
// cap mingguan 15 belum diimplementasikan, disederhanakan jadi reset harian polos).
func (s *Service) ensureDailyLivesReset(ctx context.Context, userID string) (int, error) {
	today := truncateToUTCDate(time.Now())

	var (
		remaining int
		lastReset time.Time
	)
	err := s.db.QueryRow(ctx, `
		SELECT lives_remaining, last_daily_reset_at FROM lives WHERE user_id = $1
	`, userID).Scan(&remaining, &lastReset)

	if errors.Is(err, pgx.ErrNoRows) {
		if err := s.db.QueryRow(ctx, `
			INSERT INTO lives (user_id, lives_remaining, last_daily_reset_at)
			VALUES ($1, 3, now())
			RETURNING lives_remaining
		`, userID).Scan(&remaining); err != nil {
			return 0, err
		}
		return remaining, nil
	}
	if err != nil {
		return 0, err
	}

	if truncateToUTCDate(lastReset).Before(today) {
		if err := s.db.QueryRow(ctx, `
			UPDATE lives SET lives_remaining = 3, last_daily_reset_at = now()
			WHERE user_id = $1
			RETURNING lives_remaining
		`, userID).Scan(&remaining); err != nil {
			return 0, err
		}
	}

	return remaining, nil
}

func (s *Service) applyLivesOnResult(ctx context.Context, userID string, passed bool) (int, error) {
	if passed {
		var remaining int
		err := s.db.QueryRow(ctx,
			`SELECT lives_remaining FROM lives WHERE user_id = $1`, userID,
		).Scan(&remaining)
		return remaining, err
	}

	var remaining int
	err := s.db.QueryRow(ctx, `
		UPDATE lives SET lives_remaining = GREATEST(lives_remaining - 1, 0)
		WHERE user_id = $1
		RETURNING lives_remaining
	`, userID).Scan(&remaining)
	return remaining, err
}

// applyDailyStreak: nyala sekali per hari kalender (UTC), gak peduli lulus atau
// gagal — cukup nyelesain satu sesi (FSD.md 3.7). Ini yang beda dari versi dummy
// sebelumnya di userApp, yang nambah streak tiap attempt selesai (bisa >1x/hari).
func (s *Service) applyDailyStreak(ctx context.Context, userID string) (int, error) {
	var (
		currentStreak, longestStreak int
		lastActiveDate               *time.Time
	)
	if err := s.db.QueryRow(ctx, `
		SELECT current_streak, longest_streak, last_active_date FROM users WHERE id = $1
	`, userID).Scan(&currentStreak, &longestStreak, &lastActiveDate); err != nil {
		return 0, err
	}

	today := truncateToUTCDate(time.Now())

	switch {
	case lastActiveDate == nil:
		currentStreak = 1
	case truncateToUTCDate(*lastActiveDate).Equal(today):
		// udah main hari ini, streak gak berubah
	case truncateToUTCDate(*lastActiveDate).Equal(today.AddDate(0, 0, -1)):
		currentStreak++
	default:
		currentStreak = 1
	}

	if currentStreak > longestStreak {
		longestStreak = currentStreak
	}

	if _, err := s.db.Exec(ctx, `
		UPDATE users
		SET current_streak = $1, longest_streak = $2, last_active_date = $3, updated_at = now()
		WHERE id = $4
	`, currentStreak, longestStreak, today, userID); err != nil {
		return 0, err
	}

	return currentStreak, nil
}

func (s *Service) GetMe(ctx context.Context, userID string) (*MeInfo, error) {
	var info MeInfo
	err := s.db.QueryRow(ctx, `
		SELECT id, email, display_name, role, current_streak, longest_streak,
			username, avatar_type, avatar_key, google_avatar_url, theme_preference::text, is_premium, room_quota
		FROM users WHERE id = $1
	`, userID).Scan(&info.ID, &info.Email, &info.DisplayName, &info.Role, &info.CurrentStreak, &info.LongestStreak,
		&info.Nickname, &info.Avatar.Type, &info.Avatar.Key, &info.Avatar.GoogleURL, &info.ThemePreference, &info.IsPremium,
		&info.RoomQuota)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	livesRemaining, err := s.ensureDailyLivesReset(ctx, userID)
	if err != nil {
		return nil, err
	}
	info.LivesRemaining = livesRemaining

	return &info, nil
}
