package httpserver

import (
	"errors"
	"net/http"

	"lesson/api/internal/curriculumsvc"
)

// registerAppRoutes: dipakai apps/web (userApp / end-user).
// Kurikulum, gameplay platform, sisi murid dari ujian organisasi, dan profil user sendiri.
func registerAppRoutes(mux *http.ServeMux, s *Server) {
	// Curriculum (read-only)
	mux.HandleFunc("GET /app/tiers", s.handleListTiers)
	mux.HandleFunc("GET /app/tiers/{tierCode}/batches", s.handleListBatches)
	mux.HandleFunc("GET /app/tiers/{tierCode}/categories", s.handleListCategories)
	mux.HandleFunc("GET /app/tiers/{tierCode}/leaderboard", s.handleGetTierLeaderboard)
	mux.HandleFunc("GET /app/batches/{batchId}/challenges", s.handleListChallenges)
	mux.HandleFunc("GET /app/categories/{categoryId}/challenges", s.handleListChallengesByCategory)
	mux.HandleFunc("GET /app/challenges/{challengeId}", s.notImplemented)

	// Gameplay (platform challenge)
	mux.HandleFunc("POST /app/challenges/{challengeId}/start", s.handleStartChallenge)
	mux.HandleFunc("POST /app/attempts/{attemptId}/submit", s.handleSubmitAttempt)
	mux.HandleFunc("POST /app/attempts/{attemptId}/security-events", s.handleRecordSecurityEvents)

	// Adventure (mode jalan terus lintas jenjang, lihat curriculumsvc/adventure.go)
	mux.HandleFunc("GET /app/adventure", s.handleGetAdventure)
	mux.HandleFunc("GET /app/adventure/leaderboard", s.handleGetAdventureLeaderboard)
	mux.HandleFunc("POST /app/adventure/start", s.handleStartAdventure)
	mux.HandleFunc("POST /app/adventure/rollback", s.handleRollbackAdventure)
	mux.HandleFunc("POST /app/adventure/attempts/{attemptId}/answer", s.handleAnswerAdventure)

	// Ujian organisasi — sisi murid (pembuatan/pengelolaan ujian ada di routes_org.go)
	mux.HandleFunc("POST /app/exams/{examId}/start", s.notImplemented)
	mux.HandleFunc("POST /app/organization-exam-attempts/{attemptId}/submit", s.notImplemented)

	// Undangan organisasi (murid menerima undangan lewat userApp)
	mux.HandleFunc("GET /app/invites/{token}", s.notImplemented)
	mux.HandleFunc("POST /app/invites/{token}/accept", s.notImplemented)

	// Me
	mux.HandleFunc("GET /app/me", s.handleMe)
	mux.HandleFunc("PATCH /app/me/profile", s.handleUpdateProfile)
	mux.HandleFunc("GET /app/me/nickname-check", s.handleCheckNickname)
	mux.HandleFunc("GET /app/me/progress", s.notImplemented)
	mux.HandleFunc("GET /app/me/lives", s.notImplemented)
	mux.HandleFunc("GET /app/me/organizations", s.notImplemented)
	mux.HandleFunc("POST /app/me/lives/reset", s.notImplemented)
	mux.HandleFunc("PATCH /app/me/preferences", s.notImplemented)
}

func (s *Server) handleListTiers(w http.ResponseWriter, r *http.Request) {
	tiers, err := s.curriculum.ListTiers(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error")
		return
	}
	writeJSON(w, http.StatusOK, tiers)
}

func (s *Server) handleListBatches(w http.ResponseWriter, r *http.Request) {
	tierCode := r.PathValue("tierCode")
	batches, err := s.curriculum.ListBatchesByTierCode(r.Context(), tierCode)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error")
		return
	}
	writeJSON(w, http.StatusOK, batches)
}

func (s *Server) handleListCategories(w http.ResponseWriter, r *http.Request) {
	tierCode := r.PathValue("tierCode")
	categories, err := s.curriculum.ListCategoriesByTierCode(r.Context(), tierCode)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error")
		return
	}
	writeJSON(w, http.StatusOK, categories)
}

// handleGetTierLeaderboard: top 100 achievement per jenjang (sd/smp/smk/kampus),
// bisa diliat semua user yang login.
func (s *Server) handleGetTierLeaderboard(w http.ResponseWriter, r *http.Request) {
	claims, ok := s.authenticate(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	board, err := s.curriculum.GetTierLeaderboard(r.Context(), claims.Subject, r.PathValue("tierCode"))
	if err != nil {
		if errors.Is(err, curriculumsvc.ErrTierNoLeaderboard) {
			writeError(w, http.StatusNotFound, "leaderboard_not_found")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal_error")
		return
	}
	writeJSON(w, http.StatusOK, board)
}

func (s *Server) handleListChallengesByCategory(w http.ResponseWriter, r *http.Request) {
	claims, ok := s.authenticate(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	categoryID := r.PathValue("categoryId")
	challenges, err := s.curriculum.ListChallengesByCategory(r.Context(), categoryID, claims.Subject)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error")
		return
	}
	writeJSON(w, http.StatusOK, challenges)
}

func (s *Server) handleListChallenges(w http.ResponseWriter, r *http.Request) {
	claims, ok := s.authenticate(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	batchID := r.PathValue("batchId")
	challenges, err := s.curriculum.ListChallenges(r.Context(), batchID, claims.Subject)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error")
		return
	}
	writeJSON(w, http.StatusOK, challenges)
}

func (s *Server) handleStartChallenge(w http.ResponseWriter, r *http.Request) {
	claims, ok := s.authenticate(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	challengeID := r.PathValue("challengeId")
	result, err := s.curriculum.StartChallenge(r.Context(), claims.Subject, challengeID)
	if err != nil {
		switch {
		case errors.Is(err, curriculumsvc.ErrNotFound):
			writeError(w, http.StatusNotFound, "challenge_not_found")
		case errors.Is(err, curriculumsvc.ErrNoLives):
			writeError(w, http.StatusForbidden, "no_lives")
		case errors.Is(err, curriculumsvc.ErrBatchLocked):
			writeError(w, http.StatusForbidden, "batch_locked")
		case errors.Is(err, curriculumsvc.ErrExamLocked):
			writeError(w, http.StatusForbidden, "exam_locked")
		case errors.Is(err, curriculumsvc.ErrNotEnoughBank):
			writeError(w, http.StatusConflict, "not_enough_questions")
		default:
			writeError(w, http.StatusInternalServerError, "internal_error")
		}
		return
	}
	writeJSON(w, http.StatusOK, result)
}

type submitAttemptRequest struct {
	Answers        []curriculumsvc.SubmitAnswer  `json:"answers"`
	SecurityEvents []curriculumsvc.SecurityEvent `json:"securityEvents"`
}

func (s *Server) handleSubmitAttempt(w http.ResponseWriter, r *http.Request) {
	claims, ok := s.authenticate(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req submitAttemptRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request")
		return
	}

	attemptID := r.PathValue("attemptId")
	result, err := s.curriculum.SubmitAttempt(r.Context(), claims.Subject, attemptID, req.Answers, req.SecurityEvents)
	if err != nil {
		switch {
		case errors.Is(err, curriculumsvc.ErrInvalidSecurityEvents):
			writeError(w, http.StatusBadRequest, "invalid_security_events")
		case errors.Is(err, curriculumsvc.ErrAttemptNotFound):
			writeError(w, http.StatusNotFound, "attempt_not_found")
		case errors.Is(err, curriculumsvc.ErrAttemptFinished):
			writeError(w, http.StatusConflict, "attempt_already_submitted")
		default:
			writeError(w, http.StatusInternalServerError, "internal_error")
		}
		return
	}
	writeJSON(w, http.StatusOK, result)
}

type recordSecurityEventsRequest struct {
	Events []curriculumsvc.SecurityEvent `json:"events"`
}

func (s *Server) handleRecordSecurityEvents(w http.ResponseWriter, r *http.Request) {
	claims, ok := s.authenticate(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req recordSecurityEventsRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request")
		return
	}

	attemptID := r.PathValue("attemptId")
	err := s.curriculum.RecordSecurityEvents(r.Context(), claims.Subject, attemptID, req.Events)
	if err != nil {
		switch {
		case errors.Is(err, curriculumsvc.ErrInvalidSecurityEvents):
			writeError(w, http.StatusBadRequest, "invalid_security_events")
		case errors.Is(err, curriculumsvc.ErrAttemptNotFound):
			writeError(w, http.StatusNotFound, "attempt_not_found")
		case errors.Is(err, curriculumsvc.ErrAttemptFinished):
			writeError(w, http.StatusConflict, "attempt_already_submitted")
		default:
			writeError(w, http.StatusInternalServerError, "internal_error")
		}
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) handleMe(w http.ResponseWriter, r *http.Request) {
	claims, ok := s.authenticate(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	me, err := s.curriculum.GetMe(r.Context(), claims.Subject)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error")
		return
	}
	writeJSON(w, http.StatusOK, me)
}
