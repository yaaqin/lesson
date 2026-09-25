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
	// Hasil game multiplayer (/s/{nickname}/multiplayer/{matchId}) -- cuma
	// siapa aja yang main & rank-nya, tanpa skor.
	mux.HandleFunc("GET /public/multiplayer/matches/{id}", s.handleGetMatchShare)
}

func (s *Server) handleGetMatchShare(w http.ResponseWriter, r *http.Request) {
	match, err := s.multiplayer.MatchShare(r.Context(), r.PathValue("id"))
	if err != nil {
		if errors.Is(err, curriculumsvc.ErrNotFound) {
			writeError(w, http.StatusNotFound, "match_not_found")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal_error")
		return
	}
	writeJSON(w, http.StatusOK, match)
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
