package curriculumsvc

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
)

// Share: data publik (tanpa login) buat kartu share + halaman undangan di
// apps/web (/s/{nickname}/{kind}). Dicari lewat nickname, jadi user tanpa
// nickname gak bisa di-share. Rank diambil dari server (bukan dari URL) biar
// gak bisa dipalsuin.
type ShareRank struct {
	Rank  int `json:"rank"`
	Total int `json:"total"`
}

type ShareTier struct {
	Code   string     `json:"code"`
	Name   string     `json:"name"`
	Points int        `json:"points"`
	Rank   *ShareRank `json:"rank"`
}

type ShareAdventure struct {
	CheckpointsCleared int        `json:"checkpointsCleared"`
	TotalCheckpoints   int        `json:"totalCheckpoints"`
	Completed          bool       `json:"completed"`
	TimePercent        float64    `json:"timePercent"`
	Rank               *ShareRank `json:"rank"`
}

type ShareProfile struct {
	Nickname string     `json:"nickname"`
	Avatar   UserAvatar `json:"avatar"`
	// CurrentStreak udah dicek masih nyambung (main hari ini/kemarin), beda
	// sama kolom users.current_streak yang baru di-reset pas main lagi.
	CurrentStreak int             `json:"currentStreak"`
	LongestStreak int             `json:"longestStreak"`
	Tiers         []ShareTier     `json:"tiers"`
	Adventure     *ShareAdventure `json:"adventure"`
}

func (s *Service) GetShareProfile(ctx context.Context, nickname string) (*ShareProfile, error) {
	n := NormalizeNickname(nickname)
	if n == "" {
		return nil, ErrNotFound
	}

	var (
		userID         string
		p              ShareProfile
		lastActiveDate *time.Time
	)
	err := s.db.QueryRow(ctx, `
		SELECT id, username, avatar_type, avatar_key,
			CASE WHEN avatar_type = 'google' THEN google_avatar_url END,
			current_streak, longest_streak, last_active_date
		FROM users WHERE username = $1
	`, n).Scan(&userID, &p.Nickname, &p.Avatar.Type, &p.Avatar.Key, &p.Avatar.GoogleURL,
		&p.CurrentStreak, &p.LongestStreak, &lastActiveDate)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	yesterday := truncateToUTCDate(time.Now()).AddDate(0, 0, -1)
	if lastActiveDate == nil || truncateToUTCDate(*lastActiveDate).Before(yesterday) {
		p.CurrentStreak = 0
	}

	p.Tiers = []ShareTier{}
	for _, code := range TierLeaderboardCodes {
		board, err := s.GetTierLeaderboard(ctx, userID, code)
		if err != nil {
			return nil, err
		}
		t := ShareTier{Code: code, Name: board.TierName}
		if e := findTierEntry(board); e != nil {
			t.Points = e.Points
			t.Rank = &ShareRank{Rank: e.Rank, Total: board.TotalRanked}
		}
		p.Tiers = append(p.Tiers, t)
	}

	adv, err := s.GetAdventureLeaderboard(ctx, userID)
	if err != nil {
		return nil, err
	}
	if e := findAdventureEntry(adv); e != nil {
		p.Adventure = &ShareAdventure{
			CheckpointsCleared: e.CheckpointsCleared,
			TotalCheckpoints:   adv.TotalCheckpoints,
			Completed:          e.Completed,
			TimePercent:        e.TimePercent,
			Rank:               &ShareRank{Rank: e.Rank, Total: adv.TotalRanked},
		}
	}

	return &p, nil
}

func findTierEntry(b *TierLeaderboard) *TierLeaderboardEntry {
	if b.Me != nil {
		return b.Me
	}
	for i := range b.Entries {
		if b.Entries[i].IsMe {
			return &b.Entries[i]
		}
	}
	return nil
}

func findAdventureEntry(b *AdventureLeaderboard) *AdventureLeaderboardEntry {
	if b.Me != nil {
		return b.Me
	}
	for i := range b.Entries {
		if b.Entries[i].IsMe {
			return &b.Entries[i]
		}
	}
	return nil
}
