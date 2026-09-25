package multiplayer

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"strings"
	"sync"
	"time"

	"lesson/api/internal/curriculumsvc"
)

// Tanpa huruf/angka yang gampang ketuker (0/O, 1/I/L) biar kode room enak
// didikte.
const codeAlphabet = "ABCDEFGHJKMNPQRSTUVWXYZ23456789"
const codeLength = 6

// Hub: daftar semua room yang hidup + ticket WebSocket.
type Hub struct {
	curriculum *curriculumsvc.Service

	mu      sync.Mutex
	rooms   map[string]*Room
	tickets map[string]ticket
}

// ticket: izin sekali pakai buat buka WebSocket. Browser gak bisa nempel
// header Authorization di WebSocket, jadi klien minta ticket dulu lewat HTTP
// biasa (yang udah lewat auth + auto refresh token), baru konek ?ticket=...
type ticket struct {
	userID  string
	code    string
	expires time.Time
}

func NewHub(curriculum *curriculumsvc.Service) *Hub {
	h := &Hub{
		curriculum: curriculum,
		rooms:      map[string]*Room{},
		tickets:    map[string]ticket{},
	}
	go h.janitor()
	return h
}

func (h *Hub) janitor() {
	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()
	for now := range ticker.C {
		h.mu.Lock()
		rooms := make(map[string]*Room, len(h.rooms))
		for code, r := range h.rooms {
			rooms[code] = r
		}
		for t, tk := range h.tickets {
			if now.After(tk.expires) {
				delete(h.tickets, t)
			}
		}
		h.mu.Unlock()

		// sweep ngunci room -- dijalanin di luar h.mu biar urutan lock
		// konsisten (room gak pernah ngunci hub).
		for code, r := range rooms {
			if r.sweep(now) {
				h.mu.Lock()
				delete(h.rooms, code)
				h.mu.Unlock()
			}
		}
	}
}

func (h *Hub) lookup(code string) *Room {
	h.mu.Lock()
	r := h.rooms[strings.ToUpper(code)]
	h.mu.Unlock()
	if r == nil {
		return nil
	}
	r.mu.Lock()
	closed := r.phase == PhaseClosed
	r.mu.Unlock()
	if closed {
		return nil
	}
	return r
}

type SourceOptions struct {
	Tiers                     []curriculumsvc.MultiplayerSourceTier `json:"tiers"`
	ClassicQuestionCounts     []int                                 `json:"classicQuestionCounts"`
	RaceQuestionCounts        []int                                 `json:"raceQuestionCounts"`
	SecondsPerQuestionOptions []int                                 `json:"secondsPerQuestionOptions"`
	MaxPlayers                int                                   `json:"maxPlayers"`
}

// Options: bahan form "Bikin Room" -- sumber soal dibaca dari DB tiap kali,
// jadi batch baru langsung muncul.
func (h *Hub) Options(ctx context.Context) (*SourceOptions, error) {
	tiers, err := h.curriculum.ListMultiplayerSources(ctx)
	if err != nil {
		return nil, err
	}
	return &SourceOptions{
		Tiers:                     tiers,
		ClassicQuestionCounts:     QuestionCountOptions(ModeClassic),
		RaceQuestionCounts:        QuestionCountOptions(ModeRace),
		SecondsPerQuestionOptions: SecondsPerQuestionOptions,
		MaxPlayers:                MaxPlayers,
	}, nil
}

// CreateRoom: user premium bebas bikin; user biasa motong 1 kesempatan
// (users.room_quota) -- dipotong paling akhir, jadi room yang gagal kebikin
// gak makan kuota. Soal langsung diambil di sini (bukan pas mulai) biar
// racikan yang soalnya kurang ketauan dari awal.
func (h *Hub) CreateRoom(ctx context.Context, hostID string, settings Settings) (string, error) {
	host, err := h.curriculum.GetPlayerProfile(ctx, hostID)
	if err != nil {
		return "", err
	}
	if err := settings.validate(); err != nil {
		return "", err
	}

	sources, err := h.curriculum.ListMultiplayerSources(ctx)
	if err != nil {
		return "", err
	}
	labelByID := map[string]string{}
	for _, t := range sources {
		for _, b := range t.Batches {
			labelByID[b.ID] = t.TierName + " · " + b.Name
		}
	}
	labels := make([]string, 0, len(settings.BatchIDs))
	for _, id := range settings.BatchIDs {
		label, ok := labelByID[id]
		if !ok {
			return "", ErrInvalidSource
		}
		labels = append(labels, label)
	}

	questions, err := h.curriculum.PickMultiplayerQuestions(ctx, settings.BatchIDs, settings.QuestionCount, settings.Format)
	if err != nil {
		return "", err
	}
	cfg, err := h.curriculum.GetMultiplayerConfig(ctx)
	if err != nil {
		return "", err
	}

	// Satu host cuma boleh punya 1 room aktif. Lobby lama ditutup otomatis,
	// tapi game yang lagi jalan gak diganggu.
	h.mu.Lock()
	existing := make([]*Room, 0)
	for _, r := range h.rooms {
		existing = append(existing, r)
	}
	h.mu.Unlock()
	lobbies := []*Room{}
	for _, r := range existing {
		if active, inLobby := r.isActiveHostedBy(hostID); active {
			if !inLobby {
				return "", ErrAlreadyHosting
			}
			lobbies = append(lobbies, r)
		}
	}

	if !host.IsPremium {
		if err := h.curriculum.ConsumeRoomQuota(ctx, hostID); err != nil {
			return "", err
		}
	}
	for _, r := range lobbies {
		r.mu.Lock()
		r.closeLocked("host_closed")
		r.mu.Unlock()
	}

	h.mu.Lock()
	defer h.mu.Unlock()
	code := h.newCodeLocked()
	room := newRoom(code, settings, timingFromConfig(cfg, settings.Mode), labels, questions, host)
	room.saveMatch = h.saveMatch
	h.rooms[code] = room
	return code, nil
}

func (h *Hub) saveMatch(rec curriculumsvc.MultiplayerMatchRecord) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return h.curriculum.SaveMultiplayerMatch(ctx, rec)
}

// MatchShare: data publik kartu share hasil game.
func (h *Hub) MatchShare(ctx context.Context, matchID string) (*curriculumsvc.MatchShare, error) {
	if !uuidPattern.MatchString(matchID) {
		return nil, curriculumsvc.ErrNotFound
	}
	return h.curriculum.GetMatchShare(ctx, matchID)
}

func (h *Hub) newCodeLocked() string {
	for {
		b := make([]byte, codeLength)
		rand.Read(b)
		for i := range b {
			b[i] = codeAlphabet[int(b[i])%len(codeAlphabet)]
		}
		code := string(b)
		if _, taken := h.rooms[code]; !taken {
			return code
		}
	}
}

type JoinResult struct {
	Code   string `json:"code"`
	Ticket string `json:"ticket"`
}

// Join: daftarin user ke room (atau masuk lagi kalau udah member) lalu kasih
// ticket WebSocket.
func (h *Hub) Join(ctx context.Context, userID, code string) (*JoinResult, error) {
	room := h.lookup(code)
	if room == nil {
		return nil, ErrRoomNotFound
	}
	profile, err := h.curriculum.GetPlayerProfile(ctx, userID)
	if err != nil {
		return nil, err
	}
	if err := room.admit(profile); err != nil {
		return nil, err
	}

	raw := make([]byte, 24)
	rand.Read(raw)
	t := hex.EncodeToString(raw)
	h.mu.Lock()
	h.tickets[t] = ticket{userID: userID, code: room.code, expires: time.Now().Add(ticketTTL)}
	h.mu.Unlock()
	return &JoinResult{Code: room.code, Ticket: t}, nil
}

// RedeemTicket: sekali pakai.
func (h *Hub) RedeemTicket(t string) (userID, code string, err error) {
	h.mu.Lock()
	defer h.mu.Unlock()
	tk, ok := h.tickets[t]
	if !ok || time.Now().After(tk.expires) {
		return "", "", errInvalidTicket
	}
	delete(h.tickets, t)
	return tk.userID, tk.code, nil
}
