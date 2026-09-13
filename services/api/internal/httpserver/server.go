package httpserver

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Server struct {
	db  *pgxpool.Pool
	mux *http.ServeMux
}

func New(db *pgxpool.Pool) *Server {
	s := &Server{db: db, mux: http.NewServeMux()}
	s.routes()
	return s
}

func (s *Server) Handler() http.Handler {
	return s.mux
}

func (s *Server) routes() {
	s.mux.HandleFunc("GET /health", s.handleHealth)

	api := http.NewServeMux()
	registerAuthRoutes(api, s)  // shared: dipakai userApp & dashboard (login/register akun)
	registerAppRoutes(api, s)   // dipakai apps/web (userApp / end-user)
	registerOrgRoutes(api, s)   // dipakai apps/dashboard -> /org (organization owner/admin/teacher)
	registerAdminRoutes(api, s) // dipakai apps/dashboard -> /admin (platform admin/superadmin)

	s.mux.Handle("/api/v1/", http.StripPrefix("/api/v1", api))
}

func (s *Server) notImplemented(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNotImplemented)
	json.NewEncoder(w).Encode(map[string]string{"error": "not_implemented"})
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	status := "ok"
	if err := s.db.Ping(context.Background()); err != nil {
		status = "db_unreachable"
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": status})
}
