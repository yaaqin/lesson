package httpserver

import "net/http"

// registerAppRoutes: dipakai apps/web (userApp / end-user).
// Kurikulum, gameplay platform, sisi murid dari ujian organisasi, dan profil user sendiri.
func registerAppRoutes(mux *http.ServeMux, s *Server) {
	// Curriculum (read-only)
	mux.HandleFunc("GET /app/tiers", s.notImplemented)
	mux.HandleFunc("GET /app/tiers/{tierId}/batches", s.notImplemented)
	mux.HandleFunc("GET /app/tiers/{tierId}/categories", s.notImplemented)
	mux.HandleFunc("GET /app/batches/{batchId}/challenges", s.notImplemented)
	mux.HandleFunc("GET /app/challenges/{challengeId}", s.notImplemented)

	// Gameplay (platform challenge)
	mux.HandleFunc("POST /app/challenges/{challengeId}/start", s.notImplemented)
	mux.HandleFunc("POST /app/attempts/{attemptId}/submit", s.notImplemented)

	// Ujian organisasi — sisi murid (pembuatan/pengelolaan ujian ada di routes_org.go)
	mux.HandleFunc("POST /app/exams/{examId}/start", s.notImplemented)
	mux.HandleFunc("POST /app/organization-exam-attempts/{attemptId}/submit", s.notImplemented)

	// Undangan organisasi (murid menerima undangan lewat userApp)
	mux.HandleFunc("GET /app/invites/{token}", s.notImplemented)
	mux.HandleFunc("POST /app/invites/{token}/accept", s.notImplemented)

	// Me
	mux.HandleFunc("GET /app/me", s.notImplemented)
	mux.HandleFunc("GET /app/me/progress", s.notImplemented)
	mux.HandleFunc("GET /app/me/lives", s.notImplemented)
	mux.HandleFunc("GET /app/me/organizations", s.notImplemented)
	mux.HandleFunc("POST /app/me/lives/reset", s.notImplemented)
	mux.HandleFunc("PATCH /app/me/preferences", s.notImplemented)
}
