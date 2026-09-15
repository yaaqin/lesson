package curriculumsvc

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
)

// AdminListUsers: daftar murid (role student) buat dashboard admin platform --
// nyawa di-LEFT JOIN karena baris di tabel `lives` baru kebikin pas user pertama
// kali mulai challenge (lihat ensureDailyLivesReset), jadi default 3 kalau belum ada.
func (s *Service) AdminListUsers(ctx context.Context, page, pageSize int, search string) (*AdminUserListResult, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize
	likeSearch := "%" + search + "%"

	var total int
	if err := s.db.QueryRow(ctx, `
		SELECT count(*) FROM users u
		WHERE u.role = 'student' AND ($1 = '' OR u.display_name ILIKE $2 OR u.email ILIKE $2)
	`, search, likeSearch).Scan(&total); err != nil {
		return nil, err
	}

	rows, err := s.db.Query(ctx, `
		SELECT u.id, u.email, u.display_name, u.current_streak, u.longest_streak,
			COALESCE(l.lives_remaining, 3), u.created_at
		FROM users u
		LEFT JOIN lives l ON l.user_id = u.id
		WHERE u.role = 'student' AND ($1 = '' OR u.display_name ILIKE $2 OR u.email ILIKE $2)
		ORDER BY u.created_at DESC
		LIMIT $3 OFFSET $4
	`, search, likeSearch, pageSize, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := []AdminUserListItem{}
	for rows.Next() {
		var item AdminUserListItem
		if err := rows.Scan(
			&item.ID, &item.Email, &item.DisplayName, &item.CurrentStreak, &item.LongestStreak,
			&item.LivesRemaining, &item.CreatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return &AdminUserListResult{Items: items, Total: total, Page: page, PageSize: pageSize}, nil
}

// AdminGetUserDetail: profil lengkap 1 user buat halaman detail dashboard admin
// -- streak & nyawa (persistent per-user, lihat gameplay.go) plus ringkasan
// attempt (total & lulus) dari challenge_attempts.
func (s *Service) AdminGetUserDetail(ctx context.Context, userID string) (*AdminUserDetail, error) {
	var d AdminUserDetail
	err := s.db.QueryRow(ctx, `
		SELECT u.id, u.email, u.display_name, u.role, u.current_streak, u.longest_streak,
			u.last_active_date, COALESCE(l.lives_remaining, 3),
			COALESCE(l.last_daily_reset_at, u.created_at), u.created_at
		FROM users u
		LEFT JOIN lives l ON l.user_id = u.id
		WHERE u.id = $1
	`, userID).Scan(
		&d.ID, &d.Email, &d.DisplayName, &d.Role, &d.CurrentStreak, &d.LongestStreak,
		&d.LastActiveDate, &d.LivesRemaining, &d.LivesLastResetAt, &d.CreatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	if err := s.db.QueryRow(ctx, `
		SELECT count(*), count(*) FILTER (WHERE status = 'passed')
		FROM challenge_attempts WHERE user_id = $1
	`, userID).Scan(&d.TotalAttempts, &d.PassedAttempts); err != nil {
		return nil, err
	}

	return &d, nil
}

// AdminResetUserLives: aksi manual admin buat ngembaliin nyawa user ke penuh (3)
// tanpa nunggu reset harian (ensureDailyLivesReset di gameplay.go) -- upsert
// karena user yang belum pernah mulai challenge belum punya baris di `lives`.
func (s *Service) AdminResetUserLives(ctx context.Context, userID string) error {
	var exists bool
	if err := s.db.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM users WHERE id = $1)`, userID).Scan(&exists); err != nil {
		return err
	}
	if !exists {
		return ErrNotFound
	}

	_, err := s.db.Exec(ctx, `
		INSERT INTO lives (user_id, lives_remaining, last_daily_reset_at)
		VALUES ($1, 3, now())
		ON CONFLICT (user_id) DO UPDATE SET lives_remaining = 3, last_daily_reset_at = now()
	`, userID)
	return err
}
