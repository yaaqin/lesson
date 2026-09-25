package authsvc

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/jackc/pgx/v5"
)

var ErrGoogleEmailNotVerified = errors.New("google email not verified")

// GoogleAuthURL: link buat redirect browser ke consent screen Google.
// state dibuat & diverifikasi di layer HTTP (httpserver), bukan di sini,
// karena itu urusan cookie/CSRF bukan urusan auth service.
func (s *Service) GoogleAuthURL(state string) string {
	v := url.Values{}
	v.Set("client_id", s.google.ClientID)
	v.Set("redirect_uri", s.google.RedirectURL)
	v.Set("response_type", "code")
	v.Set("scope", "openid email profile")
	v.Set("state", state)
	v.Set("access_type", "online")
	v.Set("prompt", "select_account")
	return "https://accounts.google.com/o/oauth2/v2/auth?" + v.Encode()
}

type googleTokenResponse struct {
	AccessToken string `json:"access_token"`
}

type googleUserInfo struct {
	Sub           string `json:"sub"`
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`
	Name          string `json:"name"`
	Picture       string `json:"picture"`
}

func (s *Service) exchangeGoogleCode(ctx context.Context, code string) (*googleUserInfo, error) {
	form := url.Values{}
	form.Set("code", code)
	form.Set("client_id", s.google.ClientID)
	form.Set("client_secret", s.google.ClientSecret)
	form.Set("redirect_uri", s.google.RedirectURL)
	form.Set("grant_type", "authorization_code")

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		"https://oauth2.googleapis.com/token", strings.NewReader(form.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("google token exchange failed (%d): %s", resp.StatusCode, body)
	}

	var tokenResp googleTokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return nil, err
	}

	userReq, err := http.NewRequestWithContext(ctx, http.MethodGet,
		"https://www.googleapis.com/oauth2/v3/userinfo", nil)
	if err != nil {
		return nil, err
	}
	userReq.Header.Set("Authorization", "Bearer "+tokenResp.AccessToken)

	userResp, err := http.DefaultClient.Do(userReq)
	if err != nil {
		return nil, err
	}
	defer userResp.Body.Close()

	if userResp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(userResp.Body)
		return nil, fmt.Errorf("google userinfo failed (%d): %s", userResp.StatusCode, body)
	}

	var info googleUserInfo
	if err := json.NewDecoder(userResp.Body).Decode(&info); err != nil {
		return nil, err
	}
	return &info, nil
}

// LoginWithGoogleCode: tukar authorization code dengan profil Google, lalu:
//  1. provider_uid (google sub) udah pernah dipakai -> pakai user itu
//  2. belum, tapi email-nya udah ada akun password -> otomatis di-link
//     (simplifikasi dari FSD.md 3.1 yang aslinya nawarin konfirmasi dulu;
//     di alur redirect OAuth gak ada tempat buat nanya balik ke user)
//  3. belum ada sama sekali -> bikin user baru role student, tanpa password
func (s *Service) LoginWithGoogleCode(ctx context.Context, code string) (*TokenPair, error) {
	info, err := s.exchangeGoogleCode(ctx, code)
	if err != nil {
		return nil, err
	}
	if !info.EmailVerified {
		return nil, ErrGoogleEmailNotVerified
	}

	var userID string
	err = s.db.QueryRow(ctx, `
		SELECT user_id FROM auth_providers WHERE provider = 'google' AND provider_uid = $1
	`, info.Sub).Scan(&userID)

	switch {
	case errors.Is(err, pgx.ErrNoRows):
		err = s.db.QueryRow(ctx, `SELECT id FROM users WHERE email = $1`, info.Email).Scan(&userID)
		if errors.Is(err, pgx.ErrNoRows) {
			displayName := info.Name
			if displayName == "" {
				displayName = info.Email
			}
			if err := s.db.QueryRow(ctx, `
				INSERT INTO users (email, display_name, role)
				VALUES ($1, $2, 'student')
				RETURNING id
			`, info.Email, displayName).Scan(&userID); err != nil {
				return nil, err
			}
		} else if err != nil {
			return nil, err
		}

		if _, err := s.db.Exec(ctx, `
			INSERT INTO auth_providers (user_id, provider, provider_uid)
			VALUES ($1, 'google', $2)
		`, userID, info.Sub); err != nil {
			return nil, err
		}
	case err != nil:
		return nil, err
	}

	// Foto Google disimpen/di-refresh tiap login (URL-nya bisa berubah). User
	// yang belum pernah milih avatar (masih default) langsung pakai foto ini.
	if info.Picture != "" {
		if _, err := s.db.Exec(ctx, `
			UPDATE users
			SET google_avatar_url = $1,
				avatar_type = CASE WHEN google_avatar_url IS NULL AND avatar_type = 'character' AND avatar_key = 'fox'
					THEN 'google' ELSE avatar_type END,
				updated_at = now()
			WHERE id = $2
		`, info.Picture, userID); err != nil {
			return nil, err
		}
	}

	var email, displayName, role string
	if err := s.db.QueryRow(ctx, `
		SELECT email, display_name, role FROM users WHERE id = $1
	`, userID).Scan(&email, &displayName, &role); err != nil {
		return nil, err
	}

	return s.issueTokens(ctx, userID, email, displayName, role)
}
