package httpserver

import (
	"errors"
	"net/http"
	"strconv"

	"lesson/api/internal/curriculumsvc"
)

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

	// Soal (bank soal) per challenge — prompt & opsi diedit bareng dalam 1 payload,
	// gak dipisah endpoint /options sendiri (bentuk form-nya emang selalu 1 kesatuan).
	mux.HandleFunc("GET /admin/challenges/{id}/questions", s.handleAdminListQuestions)
	mux.HandleFunc("POST /admin/challenges/{id}/questions", s.handleAdminCreateQuestion)
	mux.HandleFunc("PUT /admin/questions/{id}", s.handleAdminUpdateQuestion)
	mux.HandleFunc("DELETE /admin/questions/{id}", s.handleAdminDeleteQuestion)

	mux.HandleFunc("GET /admin/organizations", s.notImplemented)
	mux.HandleFunc("GET /admin/stats", s.notImplemented)

	// Daftar & detail murid + reset nyawa manual (nyawa & streak persistent
	// per-user, lihat curriculumsvc/gameplay.go & admin_users.go).
	mux.HandleFunc("GET /admin/users", s.handleAdminListUsers)
	mux.HandleFunc("GET /admin/users/{id}", s.handleAdminGetUser)
	mux.HandleFunc("POST /admin/users/{id}/lives/reset", s.handleAdminResetUserLives)
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

func (s *Server) handleAdminListQuestions(w http.ResponseWriter, r *http.Request) {
	if _, ok := s.authenticateRole(r, "admin", "superadmin"); !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	challengeID := r.PathValue("id")
	questions, err := s.curriculum.AdminListQuestions(r.Context(), challengeID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error")
		return
	}
	writeJSON(w, http.StatusOK, questions)
}

type questionOptionRequest struct {
	Value     float64 `json:"value"`
	IsCorrect bool    `json:"isCorrect"`
}

type upsertQuestionRequest struct {
	Type               string                  `json:"type"`
	Prompt             string                  `json:"prompt"`
	Status             string                  `json:"status"`
	Options            []questionOptionRequest `json:"options"`
	CorrectAnswerValue *float64                `json:"correctAnswerValue"`
}

func normalizeQuestionType(t string) string {
	if t == curriculumsvc.QuestionTypeEssayNumeric {
		return curriculumsvc.QuestionTypeEssayNumeric
	}
	return curriculumsvc.QuestionTypeMultipleChoice
}

func toAdminOptions(input []questionOptionRequest) []curriculumsvc.AdminQuestionOption {
	options := make([]curriculumsvc.AdminQuestionOption, len(input))
	for i, o := range input {
		options[i] = curriculumsvc.AdminQuestionOption{Value: o.Value, IsCorrect: o.IsCorrect}
	}
	return options
}

func normalizeStatus(status string) string {
	switch status {
	case "draft", "published", "archived":
		return status
	default:
		return "published"
	}
}

func (s *Server) handleAdminCreateQuestion(w http.ResponseWriter, r *http.Request) {
	if _, ok := s.authenticateRole(r, "admin", "superadmin"); !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req upsertQuestionRequest
	if err := decodeJSON(r, &req); err != nil || req.Prompt == "" {
		writeError(w, http.StatusBadRequest, "invalid_request")
		return
	}

	challengeID := r.PathValue("id")
	id, err := s.curriculum.AdminCreateQuestion(
		r.Context(), challengeID, normalizeQuestionType(req.Type), req.Prompt, normalizeStatus(req.Status),
		toAdminOptions(req.Options), req.CorrectAnswerValue,
	)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "validation_error", "message": err.Error()})
		return
	}
	writeJSON(w, http.StatusCreated, map[string]string{"id": id})
}

func (s *Server) handleAdminUpdateQuestion(w http.ResponseWriter, r *http.Request) {
	if _, ok := s.authenticateRole(r, "admin", "superadmin"); !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req upsertQuestionRequest
	if err := decodeJSON(r, &req); err != nil || req.Prompt == "" {
		writeError(w, http.StatusBadRequest, "invalid_request")
		return
	}

	questionID := r.PathValue("id")
	if err := s.curriculum.AdminUpdateQuestion(
		r.Context(), questionID, normalizeQuestionType(req.Type), req.Prompt, normalizeStatus(req.Status),
		toAdminOptions(req.Options), req.CorrectAnswerValue,
	); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "validation_error", "message": err.Error()})
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) handleAdminDeleteQuestion(w http.ResponseWriter, r *http.Request) {
	if _, ok := s.authenticateRole(r, "admin", "superadmin"); !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	questionID := r.PathValue("id")
	if err := s.curriculum.AdminDeleteQuestion(r.Context(), questionID); err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) handleAdminListUsers(w http.ResponseWriter, r *http.Request) {
	if _, ok := s.authenticateRole(r, "admin", "superadmin"); !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("pageSize"))
	search := r.URL.Query().Get("q")

	result, err := s.curriculum.AdminListUsers(r.Context(), page, pageSize, search)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error")
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (s *Server) handleAdminGetUser(w http.ResponseWriter, r *http.Request) {
	if _, ok := s.authenticateRole(r, "admin", "superadmin"); !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	userID := r.PathValue("id")
	detail, err := s.curriculum.AdminGetUserDetail(r.Context(), userID)
	if err != nil {
		switch {
		case errors.Is(err, curriculumsvc.ErrNotFound):
			writeError(w, http.StatusNotFound, "not_found")
		default:
			writeError(w, http.StatusInternalServerError, "internal_error")
		}
		return
	}
	writeJSON(w, http.StatusOK, detail)
}

func (s *Server) handleAdminResetUserLives(w http.ResponseWriter, r *http.Request) {
	if _, ok := s.authenticateRole(r, "admin", "superadmin"); !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	userID := r.PathValue("id")
	if err := s.curriculum.AdminResetUserLives(r.Context(), userID); err != nil {
		switch {
		case errors.Is(err, curriculumsvc.ErrNotFound):
			writeError(w, http.StatusNotFound, "not_found")
		default:
			writeError(w, http.StatusInternalServerError, "internal_error")
		}
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
