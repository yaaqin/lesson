package httpserver

import (
	"errors"
	"net/http"

	"lesson/api/internal/curriculumsvc"
)

func (s *Server) handleGetAdventure(w http.ResponseWriter, r *http.Request) {
	claims, ok := s.authenticate(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	state, err := s.curriculum.GetAdventure(r.Context(), claims.Subject)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error")
		return
	}
	writeJSON(w, http.StatusOK, state)
}

func (s *Server) handleStartAdventure(w http.ResponseWriter, r *http.Request) {
	claims, ok := s.authenticate(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	result, err := s.curriculum.StartAdventureCheckpoint(r.Context(), claims.Subject)
	if err != nil {
		switch {
		case errors.Is(err, curriculumsvc.ErrAdventureCompleted):
			writeError(w, http.StatusConflict, "adventure_completed")
		case errors.Is(err, curriculumsvc.ErrNotEnoughBank):
			writeError(w, http.StatusConflict, "not_enough_questions")
		default:
			writeError(w, http.StatusInternalServerError, "internal_error")
		}
		return
	}
	writeJSON(w, http.StatusOK, result)
}

type rollbackAdventureRequest struct {
	Checkpoint int `json:"checkpoint"`
}

func (s *Server) handleRollbackAdventure(w http.ResponseWriter, r *http.Request) {
	claims, ok := s.authenticate(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req rollbackAdventureRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request")
		return
	}

	state, err := s.curriculum.RollbackAdventure(r.Context(), claims.Subject, req.Checkpoint)
	if err != nil {
		switch {
		case errors.Is(err, curriculumsvc.ErrAdventureInvalidTarget):
			writeError(w, http.StatusBadRequest, "invalid_checkpoint")
		default:
			writeError(w, http.StatusInternalServerError, "internal_error")
		}
		return
	}
	writeJSON(w, http.StatusOK, state)
}

type answerAdventureRequest struct {
	QuestionIndex int      `json:"questionIndex"`
	SelectedValue *float64 `json:"selectedValue"`
}

func (s *Server) handleAnswerAdventure(w http.ResponseWriter, r *http.Request) {
	claims, ok := s.authenticate(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req answerAdventureRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request")
		return
	}

	attemptID := r.PathValue("attemptId")
	result, err := s.curriculum.AnswerAdventureQuestion(r.Context(), claims.Subject, attemptID, req.QuestionIndex, req.SelectedValue)
	if err != nil {
		switch {
		case errors.Is(err, curriculumsvc.ErrAttemptNotFound):
			writeError(w, http.StatusNotFound, "attempt_not_found")
		case errors.Is(err, curriculumsvc.ErrAttemptFinished):
			writeError(w, http.StatusConflict, "attempt_finished")
		case errors.Is(err, curriculumsvc.ErrAdventureIndexMismatch):
			writeError(w, http.StatusConflict, "question_index_mismatch")
		default:
			writeError(w, http.StatusInternalServerError, "internal_error")
		}
		return
	}
	writeJSON(w, http.StatusOK, result)
}
