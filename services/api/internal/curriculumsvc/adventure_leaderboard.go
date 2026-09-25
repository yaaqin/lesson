package curriculumsvc

import (
	"context"
	"time"
)

// Leaderboard Adventure: top N user dengan journey terjauh, bisa diliat semua
// user yang login. Cuma attempt lulus yang masih berlaku (belum di-rollback)
// yang dihitung -- 1 per checkpoint.
//
// Urutan ranking:
//  1. Checkpoint yang udah lulus (makin banyak makin atas).
//  2. Persentase waktu (makin kecil makin atas) = total waktu dipakai (udah
//     dipotong bonus) / total waktu yang disediain x 100. Waktu dipakai per
//     CP = durasi attempt lulus - sisa jatah salah x AdventureBonusSecondsPerFail
//     (min 0). Waktu disediain = jumlah time limit semua soal di CP itu.
//  3. Yang nyampe duluan (checkpoint terakhirnya lulus lebih awal).
//
// User yang belum punya nickname gak ditampilin (nama asli gak dibuka ke publik).
const AdventureLeaderboardSize = 100

type AdventureLeaderboardEntry struct {
	Rank               int        `json:"rank"`
	Nickname           string     `json:"nickname"`
	Avatar             UserAvatar `json:"avatar"`
	CheckpointsCleared int        `json:"checkpointsCleared"`
	QuestionsCleared   int        `json:"questionsCleared"`
	Completed          bool       `json:"completed"`
	TimeAllowedSeconds int        `json:"timeAllowedSeconds"`
	AdjustedSeconds    int        `json:"adjustedSeconds"`
	TimePercent        float64    `json:"timePercent"`
	LastClearedAt      time.Time  `json:"lastClearedAt"`
	IsMe               bool       `json:"isMe"`
}

type AdventureLeaderboard struct {
	Entries []AdventureLeaderboardEntry `json:"entries"`
	// Me keisi kalau user sendiri ada di ranking tapi di luar top N.
	Me                  *AdventureLeaderboardEntry `json:"me"`
	TotalRanked         int                        `json:"totalRanked"`
	TotalCheckpoints    int                        `json:"totalCheckpoints"`
	BonusSecondsPerFail int                        `json:"bonusSecondsPerFail"`
}

// adventureTimeScore: waktu setelah dipotong bonus sisa jatah salah, dan
// persentasenya terhadap waktu yang disediain. Rumusnya harus sama dengan
// query GetAdventureLeaderboard.
func adventureTimeScore(durationSeconds, failsRemaining, allowedSeconds int) (adjusted int, percent float64) {
	adjusted = max(durationSeconds-failsRemaining*AdventureBonusSecondsPerFail, 0)
	if allowedSeconds > 0 {
		percent = float64(adjusted) / float64(allowedSeconds) * 100
	}
	return adjusted, percent
}

func (s *Service) GetAdventureLeaderboard(ctx context.Context, userID string) (*AdventureLeaderboard, error) {
	rows, err := s.db.Query(ctx, `
		WITH valid AS (
			SELECT DISTINCT ON (a.user_id, a.checkpoint_no)
				a.user_id,
				a.time_allowed_seconds AS allowed,
				GREATEST(
					floor(EXTRACT(EPOCH FROM a.completed_at - a.started_at))::int
						- (a.max_fails - a.fails_used) * $1,
					0
				) AS adjusted,
				a.completed_at
			FROM adventure_checkpoint_attempts a
			WHERE a.status = 'passed' AND a.rolled_back_at IS NULL
			ORDER BY a.user_id, a.checkpoint_no, a.completed_at DESC
		),
		per_user AS (
			SELECT user_id, count(*)::int AS cleared, sum(allowed)::int AS allowed,
				sum(adjusted)::int AS adjusted, max(completed_at) AS last_at
			FROM valid
			GROUP BY user_id
		),
		ranked AS (
			SELECT p.*, u.username, u.avatar_type, u.avatar_key, u.google_avatar_url,
				CASE WHEN p.allowed > 0 THEN p.adjusted::float8 / p.allowed * 100 ELSE 0 END AS pct
			FROM per_user p
			JOIN users u ON u.id = p.user_id
			WHERE u.username IS NOT NULL
		),
		numbered AS (
			SELECT *,
				row_number() OVER (ORDER BY cleared DESC, pct ASC, last_at ASC, user_id)::int AS rank,
				count(*) OVER ()::int AS total
			FROM ranked
		)
		SELECT rank, total, user_id = $3, username, avatar_type, avatar_key,
			CASE WHEN avatar_type = 'google' THEN google_avatar_url END,
			cleared, allowed, adjusted, pct, last_at
		FROM numbered
		WHERE rank <= $2 OR user_id = $3
		ORDER BY rank
	`, AdventureBonusSecondsPerFail, AdventureLeaderboardSize, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	board := &AdventureLeaderboard{
		Entries:             []AdventureLeaderboardEntry{},
		TotalCheckpoints:    AdventureTotalCheckpoints,
		BonusSecondsPerFail: AdventureBonusSecondsPerFail,
	}
	for rows.Next() {
		var e AdventureLeaderboardEntry
		if err := rows.Scan(&e.Rank, &board.TotalRanked, &e.IsMe, &e.Nickname,
			&e.Avatar.Type, &e.Avatar.Key, &e.Avatar.GoogleURL,
			&e.CheckpointsCleared, &e.TimeAllowedSeconds, &e.AdjustedSeconds, &e.TimePercent,
			&e.LastClearedAt); err != nil {
			return nil, err
		}
		e.QuestionsCleared = e.CheckpointsCleared * AdventureQuestionsPerCheckpoint
		e.Completed = e.CheckpointsCleared >= AdventureTotalCheckpoints
		if e.Rank <= AdventureLeaderboardSize {
			board.Entries = append(board.Entries, e)
		} else {
			board.Me = &e
		}
	}
	return board, rows.Err()
}
