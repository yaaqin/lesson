package httpserver

import (
	"errors"
	"net/http"

	"lesson/api/internal/curriculumsvc"
)

type updateProfileRequest struct {
	Nickname   *string `json:"nickname"`
	AvatarType *string `json:"avatarType"`
	AvatarKey  *string `json:"avatarKey"`
}

func (s *Server) handleUpdateProfile(w http.ResponseWriter, r *http.Request) {
	claims, ok := s.authenticate(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req updateProfileRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request")
		return
	}

	me, err := s.curriculum.UpdateProfile(r.Context(), claims.Subject, curriculumsvc.ProfileUpdate{
		Nickname:   req.Nickname,
		AvatarType: req.AvatarType,
		AvatarKey:  req.AvatarKey,
	})
	if err != nil {
		switch {
		case errors.Is(err, curriculumsvc.ErrInvalidNickname):
			writeError(w, http.StatusBadRequest, "invalid_nickname")
		case errors.Is(err, curriculumsvc.ErrNicknameTaken):
			writeError(w, http.StatusConflict, "nickname_taken")
		case errors.Is(err, curriculumsvc.ErrInvalidAvatar):
			writeError(w, http.StatusBadRequest, "invalid_avatar")
		case errors.Is(err, curriculumsvc.ErrNoGoogleAvatar):
			writeError(w, http.StatusBadRequest, "no_google_avatar")
		case errors.Is(err, curriculumsvc.ErrNothingToUpdate):
			writeError(w, http.StatusBadRequest, "invalid_request")
		default:
			writeError(w, http.StatusInternalServerError, "internal_error")
		}
		return
	}
	writeJSON(w, http.StatusOK, me)
}

func (s *Server) handleCheckNickname(w http.ResponseWriter, r *http.Request) {
	claims, ok := s.authenticate(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	result, err := s.curriculum.CheckNickname(r.Context(), claims.Subject, r.URL.Query().Get("nickname"))
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error")
		return
	}
	writeJSON(w, http.StatusOK, result)
}
