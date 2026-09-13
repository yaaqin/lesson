package httpserver

import "net/http"

// registerAuthRoutes: shared auth — dipakai userApp maupun dashboard
// (satu sistem akun untuk end-user, org owner/admin/teacher, dan platform admin).
func registerAuthRoutes(mux *http.ServeMux, s *Server) {
	mux.HandleFunc("POST /auth/register", s.notImplemented)
	mux.HandleFunc("POST /auth/login", s.notImplemented)
	mux.HandleFunc("GET /auth/google", s.notImplemented)
	mux.HandleFunc("GET /auth/google/callback", s.notImplemented)
	mux.HandleFunc("POST /auth/set-password", s.notImplemented)
	mux.HandleFunc("POST /auth/refresh", s.notImplemented)
	mux.HandleFunc("POST /auth/logout", s.notImplemented)
}
