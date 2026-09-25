package curriculumsvc

import (
	"context"
	"errors"
	"slices"
	"time"
)

// Leaderboard per jenjang: top N user berdasarkan achievement di jenjang itu,
// bisa diliat semua user yang login. Tier "umum" (puzzle) belum ikut.
//
// Poin = jumlah jawaban benar. Tiap challenge cuma diambil attempt terbaiknya
// (ngulang challenge yang sama gak nambah poin, cuma bisa naikin skor
// terbaiknya):
//   - Challenge biasa: attempt yang selesai, lulus atau gak.
//   - Ujian: cuma attempt yang lulus -- ujian gagal gak nambah poin.
//
// Poin sama -> yang nyampe duluan (attempt terbaik terakhirnya lebih awal).
// User yang belum punya nickname gak ditampilin.
const TierLeaderboardSize = 100

// TierLeaderboardCodes: jenjang yang punya leaderboard.
var TierLeaderboardCodes = []string{"sd", "smp", "smk", "kampus"}

var ErrTierNoLeaderboard = errors.New("jenjang ini gak punya leaderboard")

type TierLeaderboardEntry struct {
	Rank                int        `json:"rank"`
	Nickname            string     `json:"nickname"`
	Avatar              UserAvatar `json:"avatar"`
	Points              int        `json:"points"`
	ChallengesCompleted int        `json:"challengesCompleted"`
	ExamsPassed         int        `json:"examsPassed"`
	LastScoredAt        time.Time  `json:"lastScoredAt"`
	IsMe                bool       `json:"isMe"`
}

type TierLeaderboard struct {
	TierCode string                 `json:"tierCode"`
	TierName string                 `json:"tierName"`
	Entries  []TierLeaderboardEntry `json:"entries"`
	// Me keisi kalau user sendiri ada di ranking tapi di luar top N.
	Me          *TierLeaderboardEntry `json:"me"`
	TotalRanked int                   `json:"totalRanked"`
}

func (s *Service) GetTierLeaderboard(ctx context.Context, userID, tierCode string) (*TierLeaderboard, error) {
	if !slices.Contains(TierLeaderboardCodes, tierCode) {
		return nil, ErrTierNoLeaderboard
	}

	board := &TierLeaderboard{TierCode: tierCode, Entries: []TierLeaderboardEntry{}}
	if err := s.db.QueryRow(ctx, `SELECT name FROM tiers WHERE code = $1`, tierCode).Scan(&board.TierName); err != nil {
		return nil, err
	}

	rows, err := s.db.Query(ctx, `
		WITH best AS (
			SELECT DISTINCT ON (a.user_id, a.challenge_id)
				a.user_id, a.correct_count AS points, c.is_exam, a.completed_at
			FROM challenge_attempts a
			JOIN challenges c ON c.id = a.challenge_id
			JOIN batches b ON b.id = c.batch_id
			JOIN tiers t ON t.id = b.tier_id
			WHERE t.code = $1 AND a.correct_count > 0
				AND (a.status = 'passed' OR (a.status = 'failed' AND NOT c.is_exam))
			ORDER BY a.user_id, a.challenge_id, a.correct_count DESC, a.completed_at ASC
		),
		per_user AS (
			SELECT user_id, sum(points)::int AS points, count(*)::int AS challenges,
				(count(*) FILTER (WHERE is_exam))::int AS exams, max(completed_at) AS last_at
			FROM best
			GROUP BY user_id
		),
		numbered AS (
			SELECT p.*, u.username, u.avatar_type, u.avatar_key, u.google_avatar_url,
				row_number() OVER (ORDER BY p.points DESC, p.last_at ASC, p.user_id)::int AS rank,
				count(*) OVER ()::int AS total
			FROM per_user p
			JOIN users u ON u.id = p.user_id
			WHERE u.username IS NOT NULL
		)
		SELECT rank, total, user_id = $3, username, avatar_type, avatar_key,
			CASE WHEN avatar_type = 'google' THEN google_avatar_url END,
			points, challenges, exams, last_at
		FROM numbered
		WHERE rank <= $2 OR user_id = $3
		ORDER BY rank
	`, tierCode, TierLeaderboardSize, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var e TierLeaderboardEntry
		if err := rows.Scan(&e.Rank, &board.TotalRanked, &e.IsMe, &e.Nickname,
			&e.Avatar.Type, &e.Avatar.Key, &e.Avatar.GoogleURL,
			&e.Points, &e.ChallengesCompleted, &e.ExamsPassed, &e.LastScoredAt); err != nil {
			return nil, err
		}
		if e.Rank <= TierLeaderboardSize {
			board.Entries = append(board.Entries, e)
		} else {
			board.Me = &e
		}
	}
	return board, rows.Err()
}
