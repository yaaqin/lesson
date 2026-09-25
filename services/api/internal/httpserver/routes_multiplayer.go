package httpserver

import (
	"errors"
	"net/http"

	"github.com/coder/websocket"

	"lesson/api/internal/curriculumsvc"
	"lesson/api/internal/multiplayer"
)

// registerMultiplayerRoutes: mode main bareng (apps/web /multiplayer). Alur
// klien: POST rooms (host premium) / POST rooms/{code}/join -> dapet ticket ->
// buka WebSocket GET /app/multiplayer/ws?ticket=... Semua aksi game (mulai,
// jawab, kick) lewat WebSocket, lihat internal/multiplayer.
func registerMultiplayerRoutes(mux *http.ServeMux, s *Server) {
	mux.HandleFunc("GET /app/multiplayer/options", s.handleMultiplayerOptions)
	mux.HandleFunc("POST /app/multiplayer/rooms", s.handleCreateMultiplayerRoom)
	mux.HandleFunc("POST /app/multiplayer/rooms/{code}/join", s.handleJoinMultiplayerRoom)
	mux.HandleFunc("GET /app/multiplayer/ws", s.handleMultiplayerWS)

	// Kuota bikin room: sisa kesempatan + minta tambahan ke admin.
	mux.HandleFunc("GET /app/multiplayer/quota", s.handleGetRoomQuota)
	mux.HandleFunc("POST /app/multiplayer/quota-requests", s.handleCreateRoomQuotaRequest)
}

func (s *Server) handleGetRoomQuota(w http.ResponseWriter, r *http.Request) {
	claims, ok := s.authenticate(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	status, err := s.curriculum.GetRoomQuotaStatus(r.Context(), claims.Subject)
	if err != nil {
		if errors.Is(err, curriculumsvc.ErrNotFound) {
			writeError(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal_error")
		return
	}
	writeJSON(w, http.StatusOK, status)
}

type createRoomQuotaRequestBody struct {
	Message string `json:"message"`
}

func (s *Server) handleCreateRoomQuotaRequest(w http.ResponseWriter, r *http.Request) {
	claims, ok := s.authenticate(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var body createRoomQuotaRequestBody
	if err := decodeJSON(r, &body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request")
		return
	}

	req, err := s.curriculum.CreateRoomQuotaRequest(r.Context(), claims.Subject, body.Message)
	if err != nil {
		switch {
		case errors.Is(err, curriculumsvc.ErrInvalidQuotaRequest):
			writeError(w, http.StatusBadRequest, "invalid_message")
		case errors.Is(err, curriculumsvc.ErrRoomQuotaPending):
			writeError(w, http.StatusConflict, "request_pending")
		case errors.Is(err, curriculumsvc.ErrAlreadyPremium):
			writeError(w, http.StatusConflict, "already_premium")
		case errors.Is(err, curriculumsvc.ErrNotFound):
			writeError(w, http.StatusUnauthorized, "unauthorized")
		default:
			writeError(w, http.StatusInternalServerError, "internal_error")
		}
		return
	}
	writeJSON(w, http.StatusCreated, req)
}

func (s *Server) handleMultiplayerOptions(w http.ResponseWriter, r *http.Request) {
	if _, ok := s.authenticate(r); !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	opts, err := s.multiplayer.Options(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error")
		return
	}
	writeJSON(w, http.StatusOK, opts)
}

func (s *Server) handleCreateMultiplayerRoom(w http.ResponseWriter, r *http.Request) {
	claims, ok := s.authenticate(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req multiplayer.Settings
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request")
		return
	}

	code, err := s.multiplayer.CreateRoom(r.Context(), claims.Subject, req)
	if err != nil {
		writeMultiplayerError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]string{"code": code})
}

func (s *Server) handleJoinMultiplayerRoom(w http.ResponseWriter, r *http.Request) {
	claims, ok := s.authenticate(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	result, err := s.multiplayer.Join(r.Context(), claims.Subject, r.PathValue("code"))
	if err != nil {
		writeMultiplayerError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, result)
}

// handleMultiplayerWS: auth-nya lewat ticket sekali pakai (bukan header),
// karena WebSocket di browser gak bisa ngirim header Authorization.
func (s *Server) handleMultiplayerWS(w http.ResponseWriter, r *http.Request) {
	userID, code, err := s.multiplayer.RedeemTicket(r.URL.Query().Get("ticket"))
	if err != nil {
		writeError(w, http.StatusUnauthorized, "invalid_ticket")
		return
	}

	conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{OriginPatterns: s.wsOriginPatterns})
	if err != nil {
		return // Accept udah nulis response error-nya sendiri
	}
	// r.Context() tetap hidup selama koneksi ke-hijack masih jalan.
	s.multiplayer.Serve(r.Context(), conn, userID, code)
}

func writeMultiplayerError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, multiplayer.ErrNoRoomQuota):
		writeError(w, http.StatusForbidden, "no_room_quota")
	case errors.Is(err, multiplayer.ErrInvalidSettings):
		writeError(w, http.StatusBadRequest, "invalid_settings")
	case errors.Is(err, multiplayer.ErrInvalidSource):
		writeError(w, http.StatusBadRequest, "invalid_source")
	case errors.Is(err, multiplayer.ErrNotEnoughQuestions):
		writeError(w, http.StatusUnprocessableEntity, "not_enough_questions")
	case errors.Is(err, multiplayer.ErrAlreadyHosting):
		writeError(w, http.StatusConflict, "already_hosting")
	case errors.Is(err, multiplayer.ErrRoomNotFound):
		writeError(w, http.StatusNotFound, "room_not_found")
	case errors.Is(err, multiplayer.ErrRoomFull):
		writeError(w, http.StatusConflict, "room_full")
	case errors.Is(err, multiplayer.ErrGameStarted):
		writeError(w, http.StatusConflict, "game_started")
	case errors.Is(err, multiplayer.ErrKicked):
		writeError(w, http.StatusForbidden, "kicked")
	case errors.Is(err, multiplayer.ErrNicknameRequired):
		writeError(w, http.StatusForbidden, "nickname_required")
	case errors.Is(err, curriculumsvc.ErrNotFound):
		writeError(w, http.StatusUnauthorized, "unauthorized")
	default:
		writeError(w, http.StatusInternalServerError, "internal_error")
	}
}
