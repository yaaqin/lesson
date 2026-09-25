package multiplayer

import (
	"crypto/rand"
	"fmt"
	"log"
	"math"
	"sort"
	"sync"
	"time"

	"github.com/coder/websocket"

	"lesson/api/internal/curriculumsvc"
)

// Alur: lobby -> countdown -> [question -> (reveal) -> cooldown] x N ->
// (awaiting_results -> results_countdown) -> finished. reveal cuma di race
// kalau ShowFastest dan ada yang benar; awaiting_results/results_countdown
// cuma di race kalau hasilnya dirahasiain (ShowFastest false). closed bisa
// dari phase mana aja (host nutup room / expired).
const (
	PhaseLobby            = "lobby"
	PhaseCountdown        = "countdown"
	PhaseQuestion         = "question"
	PhaseReveal           = "reveal"
	PhaseCooldown         = "cooldown"
	PhaseAwaitingResults  = "awaiting_results"
	PhaseResultsCountdown = "results_countdown"
	PhaseFinished         = "finished"
	PhaseClosed           = "closed"
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
	timing       Timing
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

	awaitingSince time.Time
	// matchID: keisi setelah hasil game sukses disimpen (buat link share).
	matchID string
	// saveMatch: nyimpen hasil ke DB (dipasang hub). Dipanggil di goroutine
	// sendiri, gak pernah sambil megang r.mu.
	saveMatch func(curriculumsvc.MultiplayerMatchRecord) error
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

func newRoom(code string, settings Settings, timing Timing, labels []string, questions []curriculumsvc.MultiplayerQuestion, host *curriculumsvc.PlayerProfile) *Room {
	now := time.Now()
	r := &Room{
		code:         code,
		settings:     settings,
		timing:       timing,
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
	if !r.maybeEndQuestionLocked() {
		r.broadcastLocked()
	}
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
	if userID == r.hostID && r.phase == PhaseAwaitingResults {
		// Host cabut tanpa nutup room -> hasilnya dibuka buat yang lain.
		r.revealResultsLocked()
		return
	}
	if !r.maybeEndQuestionLocked() {
		r.broadcastLocked()
	}
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

// closeByHost: "kill room" -- bisa kapan aja, termasuk di tengah game. Semua
// pemain langsung dikeluarin (hasil game yang belum selesai gak disimpen).
func (r *Room) closeByHost(hostID string) string {
	r.mu.Lock()
	defer r.mu.Unlock()

	if hostID != r.hostID {
		return "forbidden"
	}
	r.closeLocked("host_closed")
	return ""
}

// revealResults: host buka hasil yang dirahasiain -> countdown -> finished.
func (r *Room) revealResults(hostID string) string {
	r.mu.Lock()
	defer r.mu.Unlock()

	if hostID != r.hostID {
		return "forbidden"
	}
	if r.phase != PhaseAwaitingResults {
		return "invalid_phase"
	}
	r.revealResultsLocked()
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

// endQuestionLocked: soal ditutup. Race + ShowFastest + ada pemenang ->
// pamer pemenangnya dulu (reveal), baru cooldown.
func (r *Room) endQuestionLocked() {
	if r.settings.Mode == ModeRace && r.settings.ShowFastest && r.raceWinner != "" {
		r.phase = PhaseReveal
		r.phaseEndsAt = time.Now().Add(r.timing.WinnerReveal)
		r.scheduleLocked(r.timing.WinnerReveal, r.beginCooldownLocked)
		r.broadcastLocked()
		return
	}
	r.beginCooldownLocked()
}

// beginCooldownLocked: jeda antar soal (kunci jawaban + hitung mundur gede),
// termasuk setelah soal terakhir sebelum hasil keluar.
func (r *Room) beginCooldownLocked() {
	r.phase = PhaseCooldown
	r.phaseEndsAt = time.Now().Add(r.timing.Cooldown)
	next := r.qIndex + 1
	if next < len(r.questions) {
		r.scheduleLocked(r.timing.Cooldown, func() { r.beginQuestionLocked(next) })
	} else {
		r.scheduleLocked(r.timing.Cooldown, r.endGameLocked)
	}
	r.broadcastLocked()
}

// hidesResults: race tanpa ShowFastest -- pemenang soal & hasil akhir
// dirahasiain sampai host buka.
func (r *Room) hidesResults() bool {
	return r.settings.Mode == ModeRace && !r.settings.ShowFastest
}

// endGameLocked: semua soal selesai. Ranking dihitung (& disimpen) sekarang,
// ditampilinnya langsung atau nunggu host.
func (r *Room) endGameLocked() {
	r.results = r.computeResultsLocked()
	r.recordMatchLocked()
	if r.hidesResults() {
		r.phase = PhaseAwaitingResults
		r.awaitingSince = time.Now()
		r.phaseEndsAt = time.Time{}
		r.broadcastLocked()
		return
	}
	r.finishLocked()
}

func (r *Room) revealResultsLocked() {
	r.phase = PhaseResultsCountdown
	r.phaseEndsAt = time.Now().Add(r.timing.ResultsCountdown)
	r.scheduleLocked(r.timing.ResultsCountdown, r.finishLocked)
	r.broadcastLocked()
}

func (r *Room) finishLocked() {
	r.phase = PhaseFinished
	r.finishedAt = time.Now()
	r.phaseEndsAt = time.Time{}
	r.broadcastLocked()
}

// recordMatchLocked: simpen hasil (siapa & rank-nya) buat kartu share. Jalan
// di goroutine; matchID baru dikirim ke klien setelah beneran kesimpen, jadi
// tombol share gak pernah ngarah ke link yang 404.
func (r *Room) recordMatchLocked() {
	if r.saveMatch == nil || len(r.results) == 0 {
		return
	}
	rec := curriculumsvc.MultiplayerMatchRecord{ID: newUUID(), Mode: r.settings.Mode, QuestionCount: len(r.questions)}
	for _, e := range r.results {
		rec.Players = append(rec.Players, curriculumsvc.MultiplayerMatchRecordPlayer{UserID: e.Player.UserID, Rank: e.Rank})
	}
	save := r.saveMatch
	go func() {
		if err := save(rec); err != nil {
			log.Printf("multiplayer: simpan match room %s gagal: %v", r.code, err)
			return
		}
		r.mu.Lock()
		defer r.mu.Unlock()
		r.matchID = rec.ID
		if r.phase != PhaseClosed {
			r.broadcastLocked()
		}
	}()
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
	} else if a.correct && r.raceWinner == "" {
		// Race: yang pertama benar menang, soal langsung ditutup.
		r.raceWinner = userID
		a.points = 1
		m.score++
		m.correct++
		m.correctMs += elapsed
		r.endQuestionLocked()
		return ""
	}
	if !r.maybeEndQuestionLocked() {
		r.broadcastLocked()
	}
	return ""
}

// maybeEndQuestionLocked: kalau semua pemain yang masih online udah jawab
// (classic) / udah salah semua (race -- yang benar udah nutup soal duluan),
// sisa waktu soal di-skip. Minimal harus ada 1 pemain online: kalau semua
// lagi putus barengan (mis. iOS nutup socket), timer yang jalan biar soalnya
// gak kelewat semua. Balikin true kalau soalnya ditutup (udah broadcast).
func (r *Room) maybeEndQuestionLocked() bool {
	if r.phase != PhaseQuestion {
		return false
	}
	online := 0
	for _, m := range r.members {
		if m.playing && !m.left && m.conn != nil {
			online++
			if _, done := r.answers[m.profile.UserID]; !done {
				return false
			}
		}
	}
	if online == 0 {
		return false
	}
	r.endQuestionLocked()
	return true
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
	case PhaseAwaitingResults:
		host := r.members[r.hostID]
		hostGone := host.left || (host.conn == nil && now.Sub(host.disconnectedAt) > lobbyHostAbsentLimit)
		if hostGone || now.Sub(r.awaitingSince) > awaitingResultsLimit {
			r.revealResultsLocked()
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
	case PhaseFinished, PhaseClosed:
		return false, false
	}
	return true, false
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

// newUUID: UUID v4 (id match, dibikin di sini biar bisa dikirim ke klien
// tanpa nunggu balikan DB).
func newUUID() string {
	b := make([]byte, 16)
	rand.Read(b)
	b[6] = b[6]&0x0f | 0x40
	b[8] = b[8]&0x3f | 0x80
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}
