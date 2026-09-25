package httpserver

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"

	"lesson/api/internal/authsvc"
	"lesson/api/internal/curriculumsvc"
	"lesson/api/internal/multiplayer"
)

type Server struct {
	db             *pgxpool.Pool
	auth           *authsvc.Service
	curriculum     *curriculumsvc.Service
	multiplayer    *multiplayer.Hub
	mux            *http.ServeMux
	allowedOrigins map[string]struct{}
	// wsOriginPatterns: ALLOWED_ORIGINS apa adanya ("https://host") --
	// coder/websocket nyocokin pola yang ada "://" ke scheme://host origin.
	wsOriginPatterns []string
	webAppURL        string
}

func New(db *pgxpool.Pool, auth *authsvc.Service, curriculum *curriculumsvc.Service, mp *multiplayer.Hub, allowedOrigins []string, webAppURL string) *Server {
	originSet := make(map[string]struct{}, len(allowedOrigins))
	for _, o := range allowedOrigins {
		originSet[o] = struct{}{}
	}

	s := &Server{
		db:               db,
		auth:             auth,
		curriculum:       curriculum,
		multiplayer:      mp,
		mux:              http.NewServeMux(),
		allowedOrigins:   originSet,
		wsOriginPatterns: allowedOrigins,
		webAppURL:        webAppURL,
	}
	s.routes()
	return s
}

func (s *Server) Handler() http.Handler {
	return s.withCORS(s.mux)
}

func (s *Server) withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if _, ok := s.allowedOrigins[origin]; ok {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Vary", "Origin")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		}
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (s *Server) routes() {
	s.mux.HandleFunc("GET /health", s.handleHealth)

	api := http.NewServeMux()
	registerAuthRoutes(api, s)        // shared: dipakai userApp & dashboard (login/register akun)
	registerAppRoutes(api, s)         // dipakai apps/web (userApp / end-user)
	registerMultiplayerRoutes(api, s) // apps/web mode main bareng (WebSocket)
	registerOrgRoutes(api, s)         // dipakai apps/dashboard -> /org (organization owner/admin/teacher)
	registerAdminRoutes(api, s)       // dipakai apps/dashboard -> /admin (platform admin/superadmin)
	registerPublicRoutes(api, s)      // tanpa login: kartu share & halaman undangan apps/web

	s.mux.Handle("/api/v1/", http.StripPrefix("/api/v1", api))
}

func (s *Server) notImplemented(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusNotImplemented, map[string]string{"error": "not_implemented"})
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	status := "ok"
	if err := s.db.Ping(context.Background()); err != nil {
		status = "db_unreachable"
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": status})
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(body)
}

func decodeJSON(r *http.Request, dst any) error {
	return json.NewDecoder(r.Body).Decode(dst)
}

func writeError(w http.ResponseWriter, status int, code string) {
	writeJSON(w, status, map[string]string{"error": code})
}
