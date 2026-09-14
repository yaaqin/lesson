package authsvc

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrInvalidCredentials  = errors.New("invalid credentials")
	ErrInvalidRefreshToken = errors.New("invalid refresh token")
)

type User struct {
	ID          string
	Email       string
	DisplayName string
	Role        string
}

type TokenPair struct {
	AccessToken  string
	RefreshToken string
	User         User
}

type GoogleConfig struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
}

type Service struct {
	db              *pgxpool.Pool
	jwtSecret       string
	accessTokenTTL  time.Duration
	refreshTokenTTL time.Duration
	google          GoogleConfig
}

func NewService(db *pgxpool.Pool, jwtSecret string, accessTokenTTL, refreshTokenTTL time.Duration, google GoogleConfig) *Service {
	return &Service{
		db:              db,
		jwtSecret:       jwtSecret,
		accessTokenTTL:  accessTokenTTL,
		refreshTokenTTL: refreshTokenTTL,
		google:          google,
	}
}

// Login mencocokkan identifier (email atau username) + password terhadap
// users.password_hash (bcrypt). Sukses -> terbitkan access + refresh token baru,
// sekaligus menggantikan sesi/refresh token lama milik user ini kalau ada
// (lihat issueTokens -> 1 user cuma boleh 1 sesi aktif).
func (s *Service) Login(ctx context.Context, identifier, password string) (*TokenPair, error) {
	row := s.db.QueryRow(ctx, `
		SELECT id, email, display_name, role, password_hash
		FROM users
		WHERE email = $1 OR username = $1
	`, identifier)

	var (
		id, email, displayName, role string
		passwordHash                 *string
	)
	if err := row.Scan(&id, &email, &displayName, &role, &passwordHash); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}

	if passwordHash == nil || !VerifyPassword(*passwordHash, password) {
		return nil, ErrInvalidCredentials
	}

	return s.issueTokens(ctx, id, email, displayName, role)
}

// Refresh menukar refresh token yang valid & belum kedaluwarsa dengan pasangan
// token baru (rotate). Kalau refresh token sudah ditimpa sesi login lain
// (device/browser lain), hash-nya tidak akan cocok lagi -> ErrInvalidRefreshToken.
func (s *Service) Refresh(ctx context.Context, refreshToken string) (*TokenPair, error) {
	hash := hashRefreshToken(refreshToken)

	row := s.db.QueryRow(ctx, `
		SELECT u.id, u.email, u.display_name, u.role
		FROM user_sessions s
		JOIN users u ON u.id = s.user_id
		WHERE s.refresh_token_hash = $1 AND s.expires_at > now()
	`, hash)

	var id, email, displayName, role string
	if err := row.Scan(&id, &email, &displayName, &role); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrInvalidRefreshToken
		}
		return nil, err
	}

	return s.issueTokens(ctx, id, email, displayName, role)
}

func (s *Service) Logout(ctx context.Context, userID string) error {
	_, err := s.db.Exec(ctx, `DELETE FROM user_sessions WHERE user_id = $1`, userID)
	return err
}

func (s *Service) ParseAccessToken(tokenString string) (*AccessClaims, error) {
	return parseAccessToken(s.jwtSecret, tokenString)
}

func (s *Service) issueTokens(ctx context.Context, id, email, displayName, role string) (*TokenPair, error) {
	accessToken, err := generateAccessToken(s.jwtSecret, s.accessTokenTTL, id, email, role)
	if err != nil {
		return nil, err
	}

	refreshToken, err := generateRefreshToken()
	if err != nil {
		return nil, err
	}

	_, err = s.db.Exec(ctx, `
		INSERT INTO user_sessions (user_id, refresh_token_hash, expires_at, updated_at)
		VALUES ($1, $2, $3, now())
		ON CONFLICT (user_id) DO UPDATE
		SET refresh_token_hash = EXCLUDED.refresh_token_hash,
		    expires_at = EXCLUDED.expires_at,
		    updated_at = now()
	`, id, hashRefreshToken(refreshToken), time.Now().Add(s.refreshTokenTTL))
	if err != nil {
		return nil, err
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User:         User{ID: id, Email: email, DisplayName: displayName, Role: role},
	}, nil
}
