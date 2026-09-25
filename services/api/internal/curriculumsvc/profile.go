package curriculumsvc

import (
	"context"
	"errors"
	"regexp"
	"slices"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// Nickname = users.username: 6-20 karakter, cuma a-z, 0-9, "_" dan "." --
// sama persis dengan CHECK constraint di migrations/0009_user_profile.sql.
var nicknamePattern = regexp.MustCompile(`^[a-z0-9_.]{6,20}$`)

const (
	AvatarTypeCharacter = "character"
	AvatarTypeGoogle    = "google"
)

// AvatarCharacters: id karakter kartun yang boleh dipilih. Harus sinkron sama
// apps/web src/lib/avatars.ts (yang nentuin gambar/warnanya).
var AvatarCharacters = []string{
	"fox", "panda", "tiger", "frog", "monkey", "penguin", "lion", "koala",
	"rabbit", "bear", "unicorn", "octopus", "cat", "dog", "owl", "dino",
}

var (
	ErrInvalidNickname = errors.New("nickname gak valid")
	ErrNicknameTaken   = errors.New("nickname udah dipakai user lain")
	ErrInvalidAvatar   = errors.New("avatar gak valid")
	ErrNoGoogleAvatar  = errors.New("akun ini gak punya foto Google")
	ErrNothingToUpdate = errors.New("gak ada yang diubah")
)

// NormalizeNickname: trim + lowercase (input "Budi_123" dianggap "budi_123"),
// balikin "" kalau formatnya gak valid.
func NormalizeNickname(raw string) string {
	n := strings.ToLower(strings.TrimSpace(raw))
	if !nicknamePattern.MatchString(n) {
		return ""
	}
	return n
}

type NicknameCheck struct {
	Nickname  string `json:"nickname"`
	Valid     bool   `json:"valid"`
	Available bool   `json:"available"`
}

// CheckNickname: nickname milik user sendiri dianggap available.
func (s *Service) CheckNickname(ctx context.Context, userID, raw string) (*NicknameCheck, error) {
	n := NormalizeNickname(raw)
	if n == "" {
		return &NicknameCheck{Nickname: strings.ToLower(strings.TrimSpace(raw))}, nil
	}
	var taken bool
	if err := s.db.QueryRow(ctx, `
		SELECT EXISTS(SELECT 1 FROM users WHERE username = $1 AND id <> $2)
	`, n, userID).Scan(&taken); err != nil {
		return nil, err
	}
	return &NicknameCheck{Nickname: n, Valid: true, Available: !taken}, nil
}

type ProfileUpdate struct {
	Nickname   *string
	AvatarType *string
	AvatarKey  *string
}

func (s *Service) UpdateProfile(ctx context.Context, userID string, in ProfileUpdate) (*MeInfo, error) {
	if in.Nickname == nil && in.AvatarType == nil {
		return nil, ErrNothingToUpdate
	}

	var nickname *string
	if in.Nickname != nil {
		n := NormalizeNickname(*in.Nickname)
		if n == "" {
			return nil, ErrInvalidNickname
		}
		nickname = &n
	}

	var avatarType, avatarKey *string
	if in.AvatarType != nil {
		switch *in.AvatarType {
		case AvatarTypeCharacter:
			if in.AvatarKey == nil || !slices.Contains(AvatarCharacters, *in.AvatarKey) {
				return nil, ErrInvalidAvatar
			}
			avatarKey = in.AvatarKey
		case AvatarTypeGoogle:
			var googleURL *string
			if err := s.db.QueryRow(ctx,
				`SELECT google_avatar_url FROM users WHERE id = $1`, userID,
			).Scan(&googleURL); err != nil {
				if errors.Is(err, pgx.ErrNoRows) {
					return nil, ErrNotFound
				}
				return nil, err
			}
			if googleURL == nil || *googleURL == "" {
				return nil, ErrNoGoogleAvatar
			}
		default:
			return nil, ErrInvalidAvatar
		}
		avatarType = in.AvatarType
	}

	// Unique index users.username yang jadi penentu akhir (bukan CheckNickname)
	// biar aman dari race 2 user rebutan nickname yang sama.
	_, err := s.db.Exec(ctx, `
		UPDATE users
		SET username = COALESCE($1, username),
			avatar_type = COALESCE($2, avatar_type),
			avatar_key = COALESCE($3, avatar_key),
			updated_at = now()
		WHERE id = $4
	`, nickname, avatarType, avatarKey, userID)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, ErrNicknameTaken
		}
		return nil, err
	}

	return s.GetMe(ctx, userID)
}

// ThemePreferences: nilai enum theme_preference_type (migrations/0001).
var ThemePreferences = []string{"light", "dark", "system"}

var ErrInvalidThemePreference = errors.New("pilihan tema gak valid")

// UpdatePreferences: sekarang cuma tema tampilan apps/web.
func (s *Service) UpdatePreferences(ctx context.Context, userID, themePreference string) (*MeInfo, error) {
	if !slices.Contains(ThemePreferences, themePreference) {
		return nil, ErrInvalidThemePreference
	}
	if _, err := s.db.Exec(ctx, `
		UPDATE users SET theme_preference = $1::theme_preference_type, updated_at = now() WHERE id = $2
	`, themePreference, userID); err != nil {
		return nil, err
	}
	return s.GetMe(ctx, userID)
}
