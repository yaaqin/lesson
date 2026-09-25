package curriculumsvc

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
)

// MultiplayerConfig: jeda antar soal multiplayer, diatur admin dari dashboard
// (tabel multiplayer_config, 1 baris). Dibaca pas room dibikin.
type MultiplayerConfig struct {
	ClassicCooldownSeconds  int `json:"classicCooldownSeconds"`
	RaceCooldownSeconds     int `json:"raceCooldownSeconds"`
	RaceWinnerRevealSeconds int `json:"raceWinnerRevealSeconds"`
	ResultsCountdownSeconds int `json:"resultsCountdownSeconds"`
}

var ErrInvalidMultiplayerConfig = errors.New("pengaturan multiplayer gak valid")

// DefaultMultiplayerConfig: sama dengan default kolom di migrasi 0013.
var DefaultMultiplayerConfig = MultiplayerConfig{
	ClassicCooldownSeconds:  3,
	RaceCooldownSeconds:     3,
	RaceWinnerRevealSeconds: 1,
	ResultsCountdownSeconds: 5,
}

func (c MultiplayerConfig) Validate() error {
	inRange := func(v, min, max int) bool { return v >= min && v <= max }
	if !inRange(c.ClassicCooldownSeconds, 1, 15) || !inRange(c.RaceCooldownSeconds, 1, 15) ||
		!inRange(c.RaceWinnerRevealSeconds, 1, 5) || !inRange(c.ResultsCountdownSeconds, 1, 15) {
		return ErrInvalidMultiplayerConfig
	}
	return nil
}

func (s *Service) GetMultiplayerConfig(ctx context.Context) (*MultiplayerConfig, error) {
	var c MultiplayerConfig
	err := s.db.QueryRow(ctx, `
		SELECT classic_cooldown_seconds, race_cooldown_seconds,
			race_winner_reveal_seconds, results_countdown_seconds
		FROM multiplayer_config WHERE id
	`).Scan(&c.ClassicCooldownSeconds, &c.RaceCooldownSeconds, &c.RaceWinnerRevealSeconds, &c.ResultsCountdownSeconds)
	if errors.Is(err, pgx.ErrNoRows) {
		c = DefaultMultiplayerConfig
		return &c, nil
	}
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (s *Service) AdminUpdateMultiplayerConfig(ctx context.Context, c MultiplayerConfig) error {
	if err := c.Validate(); err != nil {
		return err
	}
	_, err := s.db.Exec(ctx, `
		INSERT INTO multiplayer_config (id, classic_cooldown_seconds, race_cooldown_seconds,
			race_winner_reveal_seconds, results_countdown_seconds, updated_at)
		VALUES (true, $1, $2, $3, $4, now())
		ON CONFLICT (id) DO UPDATE SET
			classic_cooldown_seconds = EXCLUDED.classic_cooldown_seconds,
			race_cooldown_seconds = EXCLUDED.race_cooldown_seconds,
			race_winner_reveal_seconds = EXCLUDED.race_winner_reveal_seconds,
			results_countdown_seconds = EXCLUDED.results_countdown_seconds,
			updated_at = now()
	`, c.ClassicCooldownSeconds, c.RaceCooldownSeconds, c.RaceWinnerRevealSeconds, c.ResultsCountdownSeconds)
	return err
}

// MultiplayerMatchRecord: hasil 1 game yang selesai, disimpen cuma buat kartu
// share ("aku menang lawan siapa") -- sengaja tanpa skor.
type MultiplayerMatchRecord struct {
	ID            string
	Mode          string
	QuestionCount int
	Players       []MultiplayerMatchRecordPlayer
}

type MultiplayerMatchRecordPlayer struct {
	UserID string
	Rank   int
}

func (s *Service) SaveMultiplayerMatch(ctx context.Context, m MultiplayerMatchRecord) error {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, `
		INSERT INTO multiplayer_matches (id, mode, question_count) VALUES ($1, $2, $3)
	`, m.ID, m.Mode, m.QuestionCount); err != nil {
		return err
	}
	for _, p := range m.Players {
		if _, err := tx.Exec(ctx, `
			INSERT INTO multiplayer_match_players (match_id, user_id, rank) VALUES ($1, $2, $3)
		`, m.ID, p.UserID, p.Rank); err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}

// Data publik kartu share match (/s/{nickname}/multiplayer/{matchId}).
type MatchSharePlayer struct {
	Rank     int        `json:"rank"`
	Nickname string     `json:"nickname"`
	Avatar   UserAvatar `json:"avatar"`
}

type MatchShare struct {
	ID            string             `json:"id"`
	Mode          string             `json:"mode"`
	QuestionCount int                `json:"questionCount"`
	FinishedAt    time.Time          `json:"finishedAt"`
	Players       []MatchSharePlayer `json:"players"`
}

func (s *Service) GetMatchShare(ctx context.Context, matchID string) (*MatchShare, error) {
	m := MatchShare{ID: matchID, Players: []MatchSharePlayer{}}
	err := s.db.QueryRow(ctx, `
		SELECT mode, question_count, finished_at FROM multiplayer_matches WHERE id = $1
	`, matchID).Scan(&m.Mode, &m.QuestionCount, &m.FinishedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	rows, err := s.db.Query(ctx, `
		SELECT p.rank, u.username, u.avatar_type, u.avatar_key,
			CASE WHEN u.avatar_type = 'google' THEN u.google_avatar_url END
		FROM multiplayer_match_players p
		JOIN users u ON u.id = p.user_id
		WHERE p.match_id = $1 AND u.username IS NOT NULL
		ORDER BY p.rank, u.username
	`, matchID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var p MatchSharePlayer
		if err := rows.Scan(&p.Rank, &p.Nickname, &p.Avatar.Type, &p.Avatar.Key, &p.Avatar.GoogleURL); err != nil {
			return nil, err
		}
		m.Players = append(m.Players, p)
	}
	return &m, rows.Err()
}
