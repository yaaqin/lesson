package httpserver

import "net/http"

// registerOrgRoutes: dipakai apps/dashboard -> area /org
// (organization owner/admin/teacher: kelola organisasi, grouping/kelas, undangan, dan ujian).
func registerOrgRoutes(mux *http.ServeMux, s *Server) {
	// Organization
	mux.HandleFunc("POST /org/organizations/register", s.notImplemented)
	mux.HandleFunc("GET /org/organizations/{id}", s.notImplemented)
	mux.HandleFunc("PATCH /org/organizations/{id}", s.notImplemented)
	mux.HandleFunc("GET /org/organizations/{id}/members", s.notImplemented)
	mux.HandleFunc("POST /org/organizations/{id}/invites", s.notImplemented)
	mux.HandleFunc("PATCH /org/organization-members/{id}", s.notImplemented)

	// Grouping / kelas
	mux.HandleFunc("POST /org/organizations/{id}/groups", s.notImplemented)
	mux.HandleFunc("GET /org/organizations/{id}/groups", s.notImplemented)
	mux.HandleFunc("PATCH /org/groups/{id}", s.notImplemented)
	mux.HandleFunc("DELETE /org/groups/{id}", s.notImplemented)

	// Ujian organisasi — sisi guru (start/submit oleh murid ada di routes_app.go)
	mux.HandleFunc("POST /org/organizations/{id}/exams", s.notImplemented)
	mux.HandleFunc("GET /org/organizations/{id}/exams", s.notImplemented)
	mux.HandleFunc("PATCH /org/exams/{id}", s.notImplemented)
	mux.HandleFunc("DELETE /org/exams/{id}", s.notImplemented)
	mux.HandleFunc("POST /org/exams/{id}/questions", s.notImplemented)
	mux.HandleFunc("GET /org/question-bank", s.notImplemented) // scope=global -> guru ambil dari bank soal publik
}
