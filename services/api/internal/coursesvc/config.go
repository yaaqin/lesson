package coursesvc

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
)

// Mode pembahasan setelah submit (migrations/0017).
const (
	ReviewFull       = "full"
	ReviewTimed      = "timed"
	ReviewResultOnly = "result_only"
)

const (
	minReviewSeconds = 1
	maxReviewSeconds = 120
)

var ErrInvalidConfig = errors.New("pengaturan course gak valid")

// ReviewSetting: cara nampilin pembahasan. Seconds cuma kepake buat timed.
type ReviewSetting struct {
	Mode    string `json:"mode"`
	Seconds int    `json:"seconds"`
}

type Config struct {
	Level ReviewSetting `json:"level"`
	Exam  ReviewSetting `json:"exam"`
}

var DefaultConfig = Config{
	Level: ReviewSetting{Mode: ReviewTimed, Seconds: 5},
	Exam:  ReviewSetting{Mode: ReviewResultOnly, Seconds: 5},
}

func (r ReviewSetting) valid() bool {
	switch r.Mode {
	case ReviewFull, ReviewTimed, ReviewResultOnly:
	default:
		return false
	}
	return r.Seconds >= minReviewSeconds && r.Seconds <= maxReviewSeconds
}

func (s *Service) GetConfig(ctx context.Context) (*Config, error) {
	var c Config
	err := s.db.QueryRow(ctx, `
		SELECT level_review_mode, level_review_seconds, exam_review_mode, exam_review_seconds
		FROM course_config WHERE id
	`).Scan(&c.Level.Mode, &c.Level.Seconds, &c.Exam.Mode, &c.Exam.Seconds)
	if errors.Is(err, pgx.ErrNoRows) {
		c = DefaultConfig
		return &c, nil
	}
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (s *Service) AdminUpdateConfig(ctx context.Context, c Config) error {
	if !c.Level.valid() || !c.Exam.valid() {
		return ErrInvalidConfig
	}
	_, err := s.db.Exec(ctx, `
		INSERT INTO course_config (id, level_review_mode, level_review_seconds, exam_review_mode, exam_review_seconds, updated_at)
		VALUES (true, $1, $2, $3, $4, now())
		ON CONFLICT (id) DO UPDATE SET
			level_review_mode = EXCLUDED.level_review_mode,
			level_review_seconds = EXCLUDED.level_review_seconds,
			exam_review_mode = EXCLUDED.exam_review_mode,
			exam_review_seconds = EXCLUDED.exam_review_seconds,
			updated_at = now()
	`, c.Level.Mode, c.Level.Seconds, c.Exam.Mode, c.Exam.Seconds)
	return err
}

// applyReview: mode result_only -> kunci, penjelasan, & feedback dibuang dari
// respons biar gak bisa diintip lewat network tab. Mode timed tetep ngirim
// semuanya (klien yang nyembunyiin setelah N detik).
func applyReview(res *SubmitResult, review ReviewSetting) {
	res.Review = review
	if review.Mode != ReviewResultOnly {
		return
	}
	for i := range res.Results {
		res.Results[i].Expected = ""
		res.Results[i].Explanation = ""
		res.Results[i].Feedback = ""
	}
}
