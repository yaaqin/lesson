package multiplayer

import (
	"math"
	"sort"
	"sync"
	"time"

	"github.com/coder/websocket"

	"lesson/api/internal/curriculumsvc"
)

const (
	PhaseLobby     = "lobby"
	PhaseCountdown = "countdown"
	PhaseQuestion  = "question"
	PhaseReveal    = "reveal"
	PhaseFinished  = "finished"
	PhaseClosed    = "closed"
)

// Room: satu game. SEMUA perubahan state lewat r.mu -- ini sekaligus yang
// nanganin race condition mode adu cepet: jawaban diproses satu-satu sesuai
// urutan nyampe di server, jawaban benar pertama yang dapet lock nutup soal
// (phase pindah ke reveal), jadi jawaban benar berikutnya otomatis ditolak.
//
// Pergantian phase dijadwalin pakai time.AfterFunc + nomor urut (seq): tiap
// transisi naikin seq, timer yang seq-nya udah basi (mis. soal race ditutup
// duluan karena ada yang benar) gak ngapa-ngapain pas nyala.
type Room struct {
	code         string
	settings     Settings
	sourceLabels []string
	questions    []curriculumsvc.MultiplayerQuestion
	hostID       string
	createdAt    time.Time

	mu          sync.Mutex
	members     map[string]*member
	joinOrder   []string
	banned      map[string]bool
	phase       string
	seq         int
	timer       *time.Timer
	phaseEndsAt time.Time
	qIndex      int
	qStartedAt  time.Time
	answers     map[string]*answer
	raceWinner  string
	finishedAt  time.Time
	results     []ResultEntry
}

type member struct {
	profile        curriculumsvc.PlayerProfile
	playing        bool // false = host yang cuma nonton
	conn           *client
	disconnectedAt time.Time
	left           bool
	score          int
	correct        int
	correctMs      int64 // total waktu jawab soal yang benar (tie-break)
}

type answer struct {
	value     float64
	correct   bool
	elapsedMs int64
	points    int
}

func newRoom(code string, settings Settings, labels []string, questions []curriculumsvc.MultiplayerQuestion, host *curriculumsvc.PlayerProfile) *Room {
	now := time.Now()
	r := &Room{
		code:         code,
		settings:     settings,
		sourceLabels: labels,
		questions:    questions,
		hostID:       host.UserID,
		createdAt:    now,
		members:      map[string]*member{},
		banned:       map[string]bool{},
		phase:        PhaseLobby,
	}
	r.members[host.UserID] = &member{profile: *host, playing: settings.HostPlays, disconnectedAt: now}
	r.joinOrder = append(r.joinOrder, host.UserID)
	return r
}

// ---------- masuk / keluar ----------

// admit: daftarin user ke room (dipanggil pas minta ticket, sebelum konek WS).
// User yang udah jadi member boleh masuk lagi kapan aja (reconnect).
func (r *Room) admit(p *curriculumsvc.PlayerProfile) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.phase == PhaseClosed {
		return ErrRoomNotFound
	}
	if r.banned[p.UserID] {
		return ErrKicked
	}
	if m, ok := r.members[p.UserID]; ok {
		m.left = false
		m.profile = *p
		return nil
	}
	if r.phase != PhaseLobby {
		return ErrGameStarted
	}
	if r.playingCountLocked() >= MaxPlayers {
		return ErrRoomFull
	}
	r.members[p.UserID] = &member{profile: *p, playing: true, disconnectedAt: time.Now()}
	r.joinOrder = append(r.joinOrder, p.UserID)
	r.broadcastLocked()
	return nil
}

func (r *Room) attach(userID string, c *client) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	m, ok := r.members[userID]
	if !ok || r.phase == PhaseClosed {
		return ErrRoomNotFound
	}
	if m.conn != nil && m.conn != c {
		// Tab/device lain dari akun yang sama -> koneksi lama diganti.
		m.conn.shutdown(closeReplaced, "replaced")
	}
	m.conn = c
	m.left = false
	r.broadcastLocked()
	return nil
}

func (r *Room) detach(userID string, c *client) {
	r.mu.Lock()
	defer r.mu.Unlock()

	m, ok := r.members[userID]
	if !ok || m.conn != c {
		return
	}
	m.conn = nil
	m.disconnectedAt = time.Now()
	r.broadcastLocked()
	r.maybeEndRaceQuestionLocked()
}

func (r *Room) leave(userID string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	m, ok := r.members[userID]
	if !ok {
		return
	}
	if r.phase == PhaseLobby {
		if userID == r.hostID {
			r.closeLocked("host_closed")
			return
		}
		r.removeMemberLocked(userID)
	} else {
		m.left = true
	}
	if m.conn != nil {
		m.conn.shutdown(websocket.StatusNormalClosure, "left")
		m.conn = nil
	}
	r.broadcastLocked()
	r.maybeEndRaceQuestionLocked()
}

func (r *Room) removeMemberLocked(userID string) {
	delete(r.members, userID)
	for i, id := range r.joinOrder {
		if id == userID {
			r.joinOrder = append(r.joinOrder[:i], r.joinOrder[i+1:]...)
			break
		}
	}
}

// ---------- aksi host ----------

func (r *Room) kick(hostID, targetID string) string {
	r.mu.Lock()
	defer r.mu.Unlock()

	if hostID != r.hostID || targetID == r.hostID {
		return "forbidden"
	}
	if r.phase != PhaseLobby {
		return "game_started"
	}
	m, ok := r.members[targetID]
	if !ok {
		return ""
	}
	r.banned[targetID] = true
	if m.conn != nil {
		m.conn.sendJSON(map[string]string{"type": "kicked"})
		m.conn.shutdown(closeKicked, "kicked")
	}
	r.removeMemberLocked(targetID)
	r.broadcastLocked()
	return ""
}

func (r *Room) setHostPlays(hostID string, plays bool) string {
	r.mu.Lock()
	defer r.mu.Unlock()

	if hostID != r.hostID {
		return "forbidden"
	}
	if r.phase != PhaseLobby {
		return "game_started"
	}
	m := r.members[r.hostID]
	if plays && !m.playing && r.playingCountLocked() >= MaxPlayers {
		return "room_full"
	}
	m.playing = plays
	r.settings.HostPlays = plays
	r.broadcastLocked()
	return ""
}

func (r *Room) start(hostID string) string {
	r.mu.Lock()
	defer r.mu.Unlock()

	if hostID != r.hostID {
		return "forbidden"
	}
	if r.phase != PhaseLobby {
		return "game_started"
	}
	if r.playingCountLocked() < MinPlayers {
		return "not_enough_players"
	}
	r.phase = PhaseCountdown
	r.phaseEndsAt = time.Now().Add(countdownDuration)
	r.scheduleLocked(countdownDuration, func() { r.beginQuestionLocked(0) })
	r.broadcastLocked()
	return ""
}

func (r *Room) closeByHost(hostID string) string {
	r.mu.Lock()
	defer r.mu.Unlock()

	if hostID != r.hostID {
		return "forbidden"
	}
	if r.phase != PhaseLobby && r.phase != PhaseFinished {
		return "game_started"
	}
	r.closeLocked("host_closed")
	return ""
}

// closeLocked: kabarin semua klien, putus koneksinya, room dibuang hub di
// putaran janitor berikutnya (lookup-nya udah nolak room closed).
func (r *Room) closeLocked(reason string) {
	if r.phase == PhaseClosed {
		return
	}
	r.phase = PhaseClosed
	r.seq++
	if r.timer != nil {
		r.timer.Stop()
	}
	for _, m := range r.members {
		if m.conn != nil {
			m.conn.sendJSON(map[string]string{"type": "closed", "reason": reason})
			m.conn.shutdown(websocket.StatusNormalClosure, reason)
			m.conn = nil
		}
	}
}

// ---------- alur soal ----------

func (r *Room) scheduleLocked(d time.Duration, fn func()) {
	r.seq++
	seq := r.seq
	if r.timer != nil {
		r.timer.Stop()
	}
	r.timer = time.AfterFunc(d, func() {
		r.mu.Lock()
		defer r.mu.Unlock()
		if r.seq != seq || r.phase == PhaseClosed {
			return
		}
		fn()
	})
}

func (r *Room) secondsPerQuestion() time.Duration {
	return time.Duration(r.settings.SecondsPerQuestion) * time.Second
}

func (r *Room) beginQuestionLocked(index int) {
	now := time.Now()
	r.phase = PhaseQuestion
	r.qIndex = index
	r.qStartedAt = now
	r.phaseEndsAt = now.Add(r.secondsPerQuestion())
	r.answers = map[string]*answer{}
	r.raceWinner = ""
	r.scheduleLocked(r.secondsPerQuestion(), r.endQuestionLocked)
	r.broadcastLocked()
}

func (r *Room) endQuestionLocked() {
	reveal := classicRevealDuration
	if r.settings.Mode == ModeRace {
		reveal = raceRevealDuration
	}
	r.phase = PhaseReveal
	r.phaseEndsAt = time.Now().Add(reveal)
	next := r.qIndex + 1
	if next < len(r.questions) {
		r.scheduleLocked(reveal, func() { r.beginQuestionLocked(next) })
	} else {
		r.scheduleLocked(reveal, r.finishLocked)
	}
	r.broadcastLocked()
}

func (r *Room) finishLocked() {
	r.phase = PhaseFinished
	r.finishedAt = time.Now()
	r.phaseEndsAt = time.Time{}
	r.results = r.computeResultsLocked()
	r.broadcastLocked()
}

// submitAnswer: index wajib sama dengan soal yang lagi jalan -- jawaban telat
// buat soal sebelumnya (paket nyangkut di jaringan) dibuang. Satu pemain cuma
// boleh jawab sekali per soal.
func (r *Room) submitAnswer(userID string, index int, value float64) string {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.phase != PhaseQuestion || index != r.qIndex {
		return "question_closed"
	}
	m, ok := r.members[userID]
	if !ok || !m.playing || m.left {
		return "not_playing"
	}
	if _, done := r.answers[userID]; done {
		return "already_answered"
	}
	now := time.Now()
	if !now.Before(r.phaseEndsAt) {
		return "question_closed"
	}

	q := r.questions[r.qIndex]
	elapsed := now.Sub(r.qStartedAt).Milliseconds()
	a := &answer{value: value, correct: floatEquals(q.CorrectValue, value), elapsedMs: elapsed}
	r.answers[userID] = a

	if r.settings.Mode == ModeClassic {
		if a.correct {
			limitMs := float64(r.secondsPerQuestion().Milliseconds())
			speed := 1 - math.Min(float64(elapsed)/limitMs, 1)
			a.points = classicMinPoints + int(math.Round(float64(classicMaxPoints-classicMinPoints)*speed))
			m.score += a.points
			m.correct++
			m.correctMs += elapsed
		}
		r.broadcastLocked()
		return ""
	}

	// Race: yang pertama benar menang, soal langsung ditutup.
	if a.correct && r.raceWinner == "" {
		r.raceWinner = userID
		a.points = 1
		m.score++
		m.correct++
		m.correctMs += elapsed
		r.endQuestionLocked()
		return ""
	}
	r.broadcastLocked()
	r.maybeEndRaceQuestionLocked()
	return ""
}

// maybeEndRaceQuestionLocked: race -- kalau semua pemain yang masih online
// udah salah (kekunci), gak ada gunanya nunggu timer.
func (r *Room) maybeEndRaceQuestionLocked() {
	if r.settings.Mode != ModeRace || r.phase != PhaseQuestion {
		return
	}
	for _, m := range r.members {
		if m.playing && !m.left && m.conn != nil {
			if _, done := r.answers[m.profile.UserID]; !done {
				return
			}
		}
	}
	r.endQuestionLocked()
}

// ---------- janitor ----------

// sweep: dipanggil hub berkala. Balikin true kalau room udah boleh dibuang.
func (r *Room) sweep(now time.Time) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	switch r.phase {
	case PhaseClosed:
		return true
	case PhaseFinished:
		if now.Sub(r.finishedAt) > finishedRoomTTL {
			r.closeLocked("expired")
			return true
		}
	case PhaseLobby:
		host := r.members[r.hostID]
		if now.Sub(r.createdAt) > lobbyMaxAge || (host.conn == nil && now.Sub(host.disconnectedAt) > lobbyHostAbsentLimit) {
			r.closeLocked("expired")
			return true
		}
		changed := false
		for id, m := range r.members {
			if id != r.hostID && m.conn == nil && now.Sub(m.disconnectedAt) > lobbyDisconnectGrace {
				r.removeMemberLocked(id)
				changed = true
			}
		}
		if changed {
			r.broadcastLocked()
		}
	}
	return false
}

func (r *Room) isActiveHostedBy(userID string) (active, inLobby bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.hostID != userID {
		return false, false
	}
	switch r.phase {
	case PhaseLobby:
		return true, true
	case PhaseCountdown, PhaseQuestion, PhaseReveal:
		return true, false
	}
	return false, false
}

// ---------- helper ----------

func (r *Room) playingCountLocked() int {
	n := 0
	for _, m := range r.members {
		if m.playing && !m.left {
			n++
		}
	}
	return n
}

func (r *Room) computeResultsLocked() []ResultEntry {
	entries := []ResultEntry{}
	for _, id := range r.joinOrder {
		m := r.members[id]
		if !m.playing {
			continue
		}
		var avg int64
		if m.correct > 0 {
			avg = m.correctMs / int64(m.correct)
		}
		entries = append(entries, ResultEntry{
			Player:       m.profile,
			Score:        m.score,
			CorrectCount: m.correct,
			AvgCorrectMs: avg,
			Left:         m.left,
			correctMs:    m.correctMs,
		})
	}
	// Skor -> jumlah benar -> total waktu jawab benar (lebih cepet di atas).
	sort.SliceStable(entries, func(i, j int) bool {
		a, b := entries[i], entries[j]
		if a.Score != b.Score {
			return a.Score > b.Score
		}
		if a.CorrectCount != b.CorrectCount {
			return a.CorrectCount > b.CorrectCount
		}
		return a.correctMs < b.correctMs
	})
	for i := range entries {
		if i > 0 && sameStanding(entries[i], entries[i-1]) {
			entries[i].Rank = entries[i-1].Rank
		} else {
			entries[i].Rank = i + 1
		}
	}
	return entries
}

func sameStanding(a, b ResultEntry) bool {
	return a.Score == b.Score && a.CorrectCount == b.CorrectCount && a.correctMs == b.correctMs
}

func floatEquals(a, b float64) bool {
	return math.Abs(a-b) < 1e-9
}
