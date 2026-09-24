package curriculumsvc

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
)

// Anti-cheating (defense in layers): apps/web nyatet sinyal mencurigakan selama
// attempt jalan (tab di-switch, window kehilangan fokus -- termasuk buka sidebar
// AI/extension, keluar fullscreen, nyoba copy soal) lalu ngirim batch ke sini.
// Pas submit, sinyal-sinyal itu diakumulasi jadi risk score -- sengaja gak
// langsung auto-fail begitu ada 1 sinyal. Durasi jawab per soal GAK dipakai
// sebagai sinyal (false positive buat orang dewasa yang ngulang tier SD).
//
// Catatan: semua sinyal ini datang dari klien, jadi sifatnya deterrence +
// detection buat user awam -- klien yang dimodif tetap bisa gak ngirim apa-apa.

const (
	SecurityEventTabHidden      = "tab_hidden"
	SecurityEventWindowBlur     = "window_blur"
	SecurityEventFullscreenExit = "fullscreen_exit"
	SecurityEventCopyBlocked    = "copy_blocked"

	RiskLevelNormal        = "normal"
	RiskLevelLowConfidence = "low_confidence"
	RiskLevelReview        = "review"

	maxSecurityEventsPerBatch   = 100
	maxSecurityEventsPerAttempt = 500
	maxSecurityEventDurationMs  = 24 * 60 * 60 * 1000
)

var ErrInvalidSecurityEvents = errors.New("security events gak valid")

type SecurityEvent struct {
	EventType     string    `json:"eventType"`
	QuestionIndex *int      `json:"questionIndex"`
	StartedAt     time.Time `json:"startedAt"`
	DurationMs    *int      `json:"durationMs"`
}

func validateSecurityEvents(events []SecurityEvent) error {
	if len(events) > maxSecurityEventsPerBatch {
		return ErrInvalidSecurityEvents
	}
	for i := range events {
		switch events[i].EventType {
		case SecurityEventTabHidden, SecurityEventWindowBlur, SecurityEventFullscreenExit, SecurityEventCopyBlocked:
		default:
			return ErrInvalidSecurityEvents
		}
		if events[i].StartedAt.IsZero() {
			return ErrInvalidSecurityEvents
		}
		if d := events[i].DurationMs; d != nil {
			clamped := min(max(*d, 0), maxSecurityEventDurationMs)
			events[i].DurationMs = &clamped
		}
	}
	return nil
}

// RecordSecurityEvents: endpoint batch (dipanggil tiap ganti soal / tiap
// beberapa event / interval). Cuma boleh ke attempt milik user sendiri yang
// masih in_progress.
func (s *Service) RecordSecurityEvents(ctx context.Context, userID, attemptID string, events []SecurityEvent) error {
	if err := validateSecurityEvents(events); err != nil {
		return err
	}

	var status string
	err := s.db.QueryRow(ctx, `
		SELECT status FROM challenge_attempts WHERE id = $1 AND user_id = $2
	`, attemptID, userID).Scan(&status)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrAttemptNotFound
		}
		return err
	}
	if status != "in_progress" {
		return ErrAttemptFinished
	}

	return s.insertSecurityEvents(ctx, attemptID, events)
}

// insertSecurityEvents: kelebihan dari cap per attempt dibuang diam-diam (biar
// klien gak bisa nge-spam tabel) -- 500 event udah jauh di atas ambang "review".
func (s *Service) insertSecurityEvents(ctx context.Context, attemptID string, events []SecurityEvent) error {
	if len(events) == 0 {
		return nil
	}

	var existing int
	if err := s.db.QueryRow(ctx,
		`SELECT count(*) FROM attempt_security_events WHERE attempt_id = $1`, attemptID,
	).Scan(&existing); err != nil {
		return err
	}
	room := maxSecurityEventsPerAttempt - existing
	if room <= 0 {
		return nil
	}
	if len(events) > room {
		events = events[:room]
	}

	batch := &pgx.Batch{}
	for _, e := range events {
		batch.Queue(`
			INSERT INTO attempt_security_events (attempt_id, event_type, question_index, started_at, duration_ms)
			VALUES ($1, $2, $3, $4, $5)
		`, attemptID, e.EventType, e.QuestionIndex, e.StartedAt, e.DurationMs)
	}
	return s.db.SendBatch(ctx, batch).Close()
}

// scoreAttemptRisk hitung risk score dari semua event yang udah kecatat buat
// attempt ini (dipanggil pas submit).
func (s *Service) scoreAttemptRisk(ctx context.Context, attemptID string) (int, string, error) {
	rows, err := s.db.Query(ctx, `
		SELECT event_type, duration_ms FROM attempt_security_events
		WHERE attempt_id = $1 ORDER BY started_at
	`, attemptID)
	if err != nil {
		return 0, "", err
	}
	defer rows.Close()

	var events []SecurityEvent
	for rows.Next() {
		var e SecurityEvent
		if err := rows.Scan(&e.EventType, &e.DurationMs); err != nil {
			return 0, "", err
		}
		events = append(events, e)
	}
	if err := rows.Err(); err != nil {
		return 0, "", err
	}

	score := computeRiskScore(events)
	return score, riskLevelForScore(score), nil
}

// computeRiskScore:
//   - tab_hidden: 4 poin + 1 poin tiap 5 detik pergi (durasi dicap 60 detik)
//   - window_blur: 2 poin + 1 poin tiap 5 detik (tab masih keliatan, mis.
//     fokus pindah ke sidebar extension / window lain di sebelahnya)
//   - fullscreen_exit: 2 kali pertama ditoleransi (1 poin), ke-3 dst 10 poin
//   - copy_blocked: 1 poin, total maksimal 5
//
// Skor dicap 100.
func computeRiskScore(events []SecurityEvent) int {
	score, fullscreenExits, copyPoints := 0, 0, 0
	awayPoints := func(e SecurityEvent) int {
		if e.DurationMs == nil {
			return 0
		}
		return min(*e.DurationMs, 60_000) / 5_000
	}
	for _, e := range events {
		switch e.EventType {
		case SecurityEventTabHidden:
			score += 4 + awayPoints(e)
		case SecurityEventWindowBlur:
			score += 2 + awayPoints(e)
		case SecurityEventFullscreenExit:
			fullscreenExits++
			if fullscreenExits <= 2 {
				score++
			} else {
				score += 10
			}
		case SecurityEventCopyBlocked:
			if copyPoints < 5 {
				copyPoints++
				score++
			}
		}
	}
	return min(score, 100)
}

// riskLevelForScore: <10 normal, 10-29 low_confidence (tetap valid, tapi
// ditandai buat leaderboard/kompetisi), >=30 review (di-hold buat review manual
// di konteks kompetitif).
func riskLevelForScore(score int) string {
	switch {
	case score >= 30:
		return RiskLevelReview
	case score >= 10:
		return RiskLevelLowConfidence
	default:
		return RiskLevelNormal
	}
}
