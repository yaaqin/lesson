package httpserver

import (
	"errors"
	"net/http"
	"strings"

	"lesson/api/internal/authsvc"
)

// registerAuthRoutes: shared auth — dipakai userApp maupun dashboard
// (satu sistem akun untuk end-user, org owner/admin/teacher, dan platform admin).
func registerAuthRoutes(mux *http.ServeMux, s *Server) {
	mux.HandleFunc("POST /auth/register", s.notImplemented)
	mux.HandleFunc("POST /auth/login", s.handleLogin)
	mux.HandleFunc("GET /auth/google", s.notImplemented)
	mux.HandleFunc("GET /auth/google/callback", s.notImplemented)
	mux.HandleFunc("POST /auth/set-password", s.notImplemented)
	mux.HandleFunc("POST /auth/refresh", s.handleRefresh)
	mux.HandleFunc("POST /auth/logout", s.handleLogout)
}

type loginRequest struct {
	Identifier string `json:"identifier"`
	Password   string `json:"password"`
}

type refreshRequest struct {
	RefreshToken string `json:"refreshToken"`
}

type userResponse struct {
	ID          string `json:"id"`
	Email       string `json:"email"`
	DisplayName string `json:"displayName"`
	Role        string `json:"role"`
}

type tokenResponse struct {
	AccessToken  string       `json:"accessToken"`
	RefreshToken string       `json:"refreshToken"`
	User         userResponse `json:"user"`
}

func toTokenResponse(pair *authsvc.TokenPair) tokenResponse {
	return tokenResponse{
		AccessToken:  pair.AccessToken,
		RefreshToken: pair.RefreshToken,
		User: userResponse{
			ID:          pair.User.ID,
			Email:       pair.User.Email,
			DisplayName: pair.User.DisplayName,
			Role:        pair.User.Role,
		},
	}
}

func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := decodeJSON(r, &req); err != nil || req.Identifier == "" || req.Password == "" {
		writeError(w, http.StatusBadRequest, "invalid_request")
		return
	}

	pair, err := s.auth.Login(r.Context(), req.Identifier, req.Password)
	if err != nil {
		if errors.Is(err, authsvc.ErrInvalidCredentials) {
			writeError(w, http.StatusUnauthorized, "invalid_credentials")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal_error")
		return
	}

	writeJSON(w, http.StatusOK, toTokenResponse(pair))
}

func (s *Server) handleRefresh(w http.ResponseWriter, r *http.Request) {
	var req refreshRequest
	if err := decodeJSON(r, &req); err != nil || req.RefreshToken == "" {
		writeError(w, http.StatusBadRequest, "invalid_request")
		return
	}

	pair, err := s.auth.Refresh(r.Context(), req.RefreshToken)
	if err != nil {
		if errors.Is(err, authsvc.ErrInvalidRefreshToken) {
			writeError(w, http.StatusUnauthorized, "invalid_refresh_token")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal_error")
		return
	}

	writeJSON(w, http.StatusOK, toTokenResponse(pair))
}

func (s *Server) handleLogout(w http.ResponseWriter, r *http.Request) {
	claims, ok := s.authenticate(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	if err := s.auth.Logout(r.Context(), claims.Subject); err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// authenticate ambil & validasi access token dari header Authorization: Bearer <token>.
func (s *Server) authenticate(r *http.Request) (*authsvc.AccessClaims, bool) {
	header := r.Header.Get("Authorization")
	token, found := strings.CutPrefix(header, "Bearer ")
	if !found || token == "" {
		return nil, false
	}

	claims, err := s.auth.ParseAccessToken(token)
	if err != nil {
		return nil, false
	}
	return claims, true
}

// authenticateRole sama seperti authenticate, tapi sekalian mastiin role di
// access token termasuk salah satu dari allowedRoles.
func (s *Server) authenticateRole(r *http.Request, allowedRoles ...string) (*authsvc.AccessClaims, bool) {
	claims, ok := s.authenticate(r)
	if !ok {
		return nil, false
	}
	for _, role := range allowedRoles {
		if claims.Role == role {
			return claims, true
		}
	}
	return nil, false
}
