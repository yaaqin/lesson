package curriculumsvc

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
)

// Aturan kunci per jenjang (tier yang pakai batch):
//   - Batch pertama selalu kebuka. Batch berikutnya kebuka kalau semua ujian
//     di batch sebelumnya (urut order_index) udah lulus.
//   - Ujian di sebuah batch kebuka kalau user udah lulus minimal
//     examUnlockPercent% challenge latihan (non-ujian) di batch itu, dibulatin
//     ke atas (10 latihan -> 8, 5 -> 4, 12 -> 10).
//   - Challenge latihan di batch yang kebuka bebas dikerjain tanpa urutan.
//
// Tier tanpa batch (kategori, mis. "umum") gak kena aturan ini.
const examUnlockPercent = 80

const (
	LockReasonBatch = "batch"
	LockReasonExam  = "exam"
)

var (
	ErrBatchLocked = errors.New("batch masih kekunci -- lulus ujian batch sebelumnya dulu")
	ErrExamLocked  = errors.New("ujian masih kekunci -- lulus lebih banyak latihan di batch ini dulu")
)

type batchAccess struct {
	Unlocked       bool
	PassedLatihan  int
	TotalLatihan   int
	RequiredToExam int
}

func (s *Service) batchAccess(ctx context.Context, userID, batchID string) (*batchAccess, error) {
	var a batchAccess
	err := s.db.QueryRow(ctx, `
		WITH b AS (
			SELECT id, tier_id, order_index FROM batches WHERE id = $1
		),
		prev AS (
			SELECT pb.id FROM batches pb, b
			WHERE pb.tier_id = b.tier_id AND pb.order_index < b.order_index
			ORDER BY pb.order_index DESC
			LIMIT 1
		)
		SELECT
			NOT EXISTS (
				SELECT 1 FROM challenges c JOIN prev ON c.batch_id = prev.id
				WHERE c.is_exam AND NOT EXISTS (
					SELECT 1 FROM challenge_attempts a
					WHERE a.challenge_id = c.id AND a.user_id = $2 AND a.status = 'passed'
				)
			),
			(SELECT count(*) FROM challenges c WHERE c.batch_id = $1 AND NOT c.is_exam),
			(
				SELECT count(*) FROM challenges c
				WHERE c.batch_id = $1 AND NOT c.is_exam AND EXISTS (
					SELECT 1 FROM challenge_attempts a
					WHERE a.challenge_id = c.id AND a.user_id = $2 AND a.status = 'passed'
				)
			)
		FROM b
	`, batchID, userID).Scan(&a.Unlocked, &a.TotalLatihan, &a.PassedLatihan)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	a.RequiredToExam = (a.TotalLatihan*examUnlockPercent + 99) / 100
	return &a, nil
}

// applyBatchLocks isi field kunci di list challenge 1 batch.
func applyBatchLocks(items []ChallengeListItem, access *batchAccess) {
	for i := range items {
		if items[i].IsExam {
			items[i].ExamRequiredPassed = access.RequiredToExam
			items[i].ExamPassedCount = access.PassedLatihan
		}
		switch {
		case !access.Unlocked:
			items[i].Locked, items[i].LockReason = true, LockReasonBatch
		case items[i].IsExam && access.PassedLatihan < access.RequiredToExam:
			items[i].Locked, items[i].LockReason = true, LockReasonExam
		}
	}
}

// ensureChallengeUnlocked dipanggil sebelum StartChallenge bikin attempt.
func (s *Service) ensureChallengeUnlocked(ctx context.Context, userID string, batchID *string, isExam bool) error {
	if batchID == nil {
		return nil
	}
	access, err := s.batchAccess(ctx, userID, *batchID)
	if err != nil {
		return err
	}
	if !access.Unlocked {
		return ErrBatchLocked
	}
	if isExam && access.PassedLatihan < access.RequiredToExam {
		return ErrExamLocked
	}
	return nil
}
