package httpserver

import (
	"errors"
	"net/http"

	"lesson/api/internal/curriculumsvc"
)

// registerPublicRoutes: tanpa login -- dipakai halaman undangan & kartu share
// di apps/web (/s/{nickname}/{kind}), termasuk dari server Next (OG image).
func registerPublicRoutes(mux *http.ServeMux, s *Server) {
	mux.HandleFunc("GET /public/share/{nickname}", s.handleGetShareProfile)
}

func (s *Server) handleGetShareProfile(w http.ResponseWriter, r *http.Request) {
	profile, err := s.curriculum.GetShareProfile(r.Context(), r.PathValue("nickname"))
	if err != nil {
		if errors.Is(err, curriculumsvc.ErrNotFound) {
			writeError(w, http.StatusNotFound, "user_not_found")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal_error")
		return
	}
	writeJSON(w, http.StatusOK, profile)
}
