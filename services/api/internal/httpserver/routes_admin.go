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

	// Pohon Tier -> Batch -> Challenge + edit waktu (FSD.md belum punya endpoint
	// ini secara eksplisit; ditambahkan sesuai request: "waktu per soal bisa
	// diedit di dashboard").
	mux.HandleFunc("GET /admin/curriculum", s.handleAdminListCurriculum)
	mux.HandleFunc("PATCH /admin/challenges/{id}/timing", s.handleAdminUpdateChallengeTiming)

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

func (s *Server) handleAdminListCurriculum(w http.ResponseWriter, r *http.Request) {
	if _, ok := s.authenticateRole(r, "admin", "superadmin"); !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	tiers, err := s.curriculum.AdminListCurriculum(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error")
		return
	}
	writeJSON(w, http.StatusOK, tiers)
}

type updateChallengeTimingRequest struct {
	TimeLimitSeconds int `json:"timeLimitSeconds"`
}

func (s *Server) handleAdminUpdateChallengeTiming(w http.ResponseWriter, r *http.Request) {
	if _, ok := s.authenticateRole(r, "admin", "superadmin"); !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req updateChallengeTimingRequest
	if err := decodeJSON(r, &req); err != nil || req.TimeLimitSeconds <= 0 {
		writeError(w, http.StatusBadRequest, "invalid_request")
		return
	}

	challengeID := r.PathValue("id")
	if err := s.curriculum.AdminUpdateChallengeTiming(r.Context(), challengeID, req.TimeLimitSeconds); err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
