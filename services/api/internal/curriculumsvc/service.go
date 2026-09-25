package curriculumsvc

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Service struct {
	db *pgxpool.Pool
}

func NewService(db *pgxpool.Pool) *Service {
	return &Service{db: db}
}

func (s *Service) ListTiers(ctx context.Context) ([]Tier, error) {
	rows, err := s.db.Query(ctx, `
		SELECT id, code, name, uses_batch FROM tiers ORDER BY order_index
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tiers := []Tier{}
	for rows.Next() {
		var t Tier
		if err := rows.Scan(&t.ID, &t.Code, &t.Name, &t.UsesBatch); err != nil {
			return nil, err
		}
		tiers = append(tiers, t)
	}
	return tiers, rows.Err()
}

// ListCategoriesByTierCode: dipakai tier yang uses_batch = false sebagai
// pengganti ListBatchesByTierCode.
func (s *Service) ListCategoriesByTierCode(ctx context.Context, tierCode string) ([]Category, error) {
	rows, err := s.db.Query(ctx, `
		SELECT c.id, c.code, c.name
		FROM categories c
		JOIN tiers t ON t.id = c.tier_id
		WHERE t.code = $1
		ORDER BY c.order_index
	`, tierCode)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	categories := []Category{}
	for rows.Next() {
		var c Category
		if err := rows.Scan(&c.ID, &c.Code, &c.Name); err != nil {
			return nil, err
		}
		categories = append(categories, c)
	}
	return categories, rows.Err()
}

// ListChallengesByCategory: sama kayak ListChallenges tapi buat challenge yang
// nempel ke category (bukan batch) -- lihat chk_challenges_batch_xor_category.
func (s *Service) ListChallengesByCategory(ctx context.Context, categoryID, userID string) ([]ChallengeListItem, error) {
	rows, err := s.db.Query(ctx, `
		SELECT
			c.id, c.name, c.is_exam, c.question_count_required,
			c.pass_threshold_percent, c.time_limit_seconds,
			EXISTS(
				SELECT 1 FROM challenge_attempts a
				WHERE a.challenge_id = c.id AND a.user_id = $2 AND a.status = 'passed'
			) AS completed,
			EXISTS(
				SELECT 1 FROM questions q
				WHERE q.challenge_id = c.id AND q.status = 'published' AND q.question_type = 'essay_numeric'
			) AS has_essay
		FROM challenges c
		WHERE c.category_id = $1
		ORDER BY c.order_index
	`, categoryID, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := []ChallengeListItem{}
	for rows.Next() {
		var c ChallengeListItem
		if err := rows.Scan(
			&c.ID, &c.Name, &c.IsExam, &c.QuestionCountRequired,
			&c.PassThresholdPercent, &c.TimeLimitSeconds, &c.Completed, &c.HasEssay,
		); err != nil {
			return nil, err
		}
		items = append(items, c)
	}
	return items, rows.Err()
}

func (s *Service) ListBatchesByTierCode(ctx context.Context, tierCode string) ([]Batch, error) {
	rows, err := s.db.Query(ctx, `
		SELECT b.id, b.name
		FROM batches b
		JOIN tiers t ON t.id = b.tier_id
		WHERE t.code = $1
		ORDER BY b.order_index
	`, tierCode)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	batches := []Batch{}
	for rows.Next() {
		var b Batch
		if err := rows.Scan(&b.ID, &b.Name); err != nil {
			return nil, err
		}
		batches = append(batches, b)
	}
	return batches, rows.Err()
}

func (s *Service) ListChallenges(ctx context.Context, batchID, userID string) ([]ChallengeListItem, error) {
	rows, err := s.db.Query(ctx, `
		SELECT
			c.id, c.name, c.is_exam, c.question_count_required,
			c.pass_threshold_percent, c.time_limit_seconds,
			EXISTS(
				SELECT 1 FROM challenge_attempts a
				WHERE a.challenge_id = c.id AND a.user_id = $2 AND a.status = 'passed'
			) AS completed,
			EXISTS(
				SELECT 1 FROM questions q
				WHERE q.challenge_id = c.id AND q.status = 'published' AND q.question_type = 'essay_numeric'
			) AS has_essay
		FROM challenges c
		WHERE c.batch_id = $1
		ORDER BY c.order_index
	`, batchID, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := []ChallengeListItem{}
	for rows.Next() {
		var c ChallengeListItem
		if err := rows.Scan(
			&c.ID, &c.Name, &c.IsExam, &c.QuestionCountRequired,
			&c.PassThresholdPercent, &c.TimeLimitSeconds, &c.Completed, &c.HasEssay,
		); err != nil {
			return nil, err
		}
		items = append(items, c)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	access, err := s.batchAccess(ctx, userID, batchID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return items, nil
		}
		return nil, err
	}
	applyBatchLocks(items, access)
	return items, nil
}
