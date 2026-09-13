package httpserver

import "net/http"

// registerAdminRoutes: dipakai apps/dashboard -> area /admin
// (platform admin/superadmin: bank soal global, kurikulum, dan daftar organisasi terdaftar).
func registerAdminRoutes(mux *http.ServeMux, s *Server) {
	mux.HandleFunc("POST /admin/tiers", s.notImplemented)
	mux.HandleFunc("PUT /admin/tiers/{id}", s.notImplemented)
	mux.HandleFunc("DELETE /admin/tiers/{id}", s.notImplemented)

	mux.HandleFunc("POST /admin/batches", s.notImplemented)
	mux.HandleFunc("PUT /admin/batches/{id}", s.notImplemented)
	mux.HandleFunc("DELETE /admin/batches/{id}", s.notImplemented)

	mux.HandleFunc("POST /admin/categories", s.notImplemented)
	mux.HandleFunc("PUT /admin/categories/{id}", s.notImplemented)
	mux.HandleFunc("DELETE /admin/categories/{id}", s.notImplemented)

	mux.HandleFunc("POST /admin/challenges", s.notImplemented)
	mux.HandleFunc("PUT /admin/challenges/{id}", s.notImplemented)
	mux.HandleFunc("DELETE /admin/challenges/{id}", s.notImplemented)

	mux.HandleFunc("POST /admin/questions", s.notImplemented)
	mux.HandleFunc("PUT /admin/questions/{id}", s.notImplemented)
	mux.HandleFunc("DELETE /admin/questions/{id}", s.notImplemented)
	mux.HandleFunc("POST /admin/questions/{id}/options", s.notImplemented)
	mux.HandleFunc("PUT /admin/questions/{id}/options", s.notImplemented)
	mux.HandleFunc("DELETE /admin/questions/{id}/options", s.notImplemented)

	mux.HandleFunc("GET /admin/organizations", s.notImplemented)
	mux.HandleFunc("GET /admin/stats", s.notImplemented)
}
