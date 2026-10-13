package httpserver

import (
	"errors"
	"net/http"

	"lesson/api/internal/coursesvc"
)

// registerCourseRoutes: Course Coding Dasar di apps/web (/coding/course) --
// mode belajar materi, kepisah dari kuis (gak ada nyawa).
func registerCourseRoutes(mux *http.ServeMux, s *Server) {
	mux.HandleFunc("GET /app/course/modules", s.handleCourseModules)
	mux.HandleFunc("GET /app/course/levels/{levelId}", s.handleCourseLevel)
	mux.HandleFunc("POST /app/course/levels/{levelId}/start", s.handleCourseStartLevel)
	mux.HandleFunc("POST /app/course/modules/{moduleCode}/exam/start", s.handleCourseStartExam)
	mux.HandleFunc("POST /app/course/attempts/{attemptId}/submit", s.handleCourseSubmit)
}

func writeCourseError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, coursesvc.ErrNotFound):
		writeError(w, http.StatusNotFound, "not_found")
	case errors.Is(err, coursesvc.ErrLocked):
		writeError(w, http.StatusForbidden, "locked")
	case errors.Is(err, coursesvc.ErrEmptyBank):
		writeError(w, http.StatusConflict, "not_enough_questions")
	case errors.Is(err, coursesvc.ErrAttemptNotFound):
		writeError(w, http.StatusNotFound, "attempt_not_found")
	case errors.Is(err, coursesvc.ErrAttemptFinished):
		writeError(w, http.StatusConflict, "attempt_already_submitted")
	default:
		writeError(w, http.StatusInternalServerError, "internal_error")
	}
}

func (s *Server) handleCourseModules(w http.ResponseWriter, r *http.Request) {
	claims, ok := s.authenticate(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	modules, err := s.course.ListModules(r.Context(), claims.Subject)
	if err != nil {
		writeCourseError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, modules)
}

func (s *Server) handleCourseLevel(w http.ResponseWriter, r *http.Request) {
	claims, ok := s.authenticate(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	level, err := s.course.GetLevel(r.Context(), claims.Subject, r.PathValue("levelId"))
	if err != nil {
		writeCourseError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, level)
}

func (s *Server) handleCourseStartLevel(w http.ResponseWriter, r *http.Request) {
	claims, ok := s.authenticate(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	result, err := s.course.StartLevel(r.Context(), claims.Subject, r.PathValue("levelId"))
	if err != nil {
		writeCourseError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (s *Server) handleCourseStartExam(w http.ResponseWriter, r *http.Request) {
	claims, ok := s.authenticate(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	result, err := s.course.StartExam(r.Context(), claims.Subject, r.PathValue("moduleCode"))
	if err != nil {
		writeCourseError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, result)
}

type courseSubmitRequest struct {
	Answers []coursesvc.Answer `json:"answers"`
}

func (s *Server) handleCourseSubmit(w http.ResponseWriter, r *http.Request) {
	claims, ok := s.authenticate(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	var req courseSubmitRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request")
		return
	}
	result, err := s.course.Submit(r.Context(), claims.Subject, r.PathValue("attemptId"), req.Answers)
	if err != nil {
		writeCourseError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, result)
}
