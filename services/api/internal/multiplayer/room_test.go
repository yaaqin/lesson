package multiplayer

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"lesson/api/internal/curriculumsvc"
)

// Timer dibikin panjang biar gak nyala sendiri selama tes -- transisi phase
// dipanggil manual.
var testTiming = Timing{Cooldown: time.Hour, WinnerReveal: time.Hour, ResultsCountdown: time.Hour}

func testRoom(t *testing.T, mode string, players int) *Room {
	return testRoomWith(t, mode, players, false)
}

func testRoomWith(t *testing.T, mode string, players int, showFastest bool) *Room {
	t.Helper()
	questions := []curriculumsvc.MultiplayerQuestion{
		{ID: "q1", TierCode: "sd", Type: curriculumsvc.QuestionTypeMultipleChoice, Prompt: "2 + 3", Options: []float64{4, 5, 6, 7}, CorrectValue: 5},
		{ID: "q2", TierCode: "sd", Type: curriculumsvc.QuestionTypeEssayNumeric, Prompt: "6 x 7", CorrectValue: 42},
	}
	host := &curriculumsvc.PlayerProfile{UserID: "u0", Nickname: "host_0"}
	settings := Settings{Mode: mode, QuestionCount: 5, Format: "mixed", SecondsPerQuestion: 20, HostPlays: true, ShowFastest: showFastest}
	r := newRoom("TEST01", settings, testTiming, nil, questions, host)
	for i := 1; i < players; i++ {
		p := &curriculumsvc.PlayerProfile{UserID: fmt.Sprintf("u%d", i), Nickname: fmt.Sprintf("player_%d", i)}
		if err := r.admit(p); err != nil {
			t.Fatalf("admit %d: %v", i, err)
		}
	}
	for id := range r.members {
		if err := r.attach(id, newClient(nil)); err != nil {
			t.Fatalf("attach %s: %v", id, err)
		}
	}
	if code := r.start("u0"); code != "" {
		t.Fatalf("start: %s", code)
	}
	r.mu.Lock()
	r.beginQuestionLocked(0)
	r.mu.Unlock()
	return r
}

// Race: 10 pemain kirim jawaban benar barengan -> tepat 1 yang dapet poin.
func TestRaceSingleWinnerUnderConcurrency(t *testing.T) {
	for round := 0; round < 50; round++ {
		r := testRoom(t, ModeRace, MaxPlayers)

		var wg sync.WaitGroup
		start := make(chan struct{})
		for id := range r.members {
			wg.Add(1)
			go func() {
				defer wg.Done()
				<-start
				r.submitAnswer(id, 0, 5)
			}()
		}
		close(start)
		wg.Wait()

		r.mu.Lock()
		total := 0
		for _, m := range r.members {
			total += m.score
		}
		if total != 1 {
			t.Fatalf("round %d: total poin %d, harusnya 1", round, total)
		}
		if r.phase != PhaseCooldown || r.raceWinner == "" {
			t.Fatalf("round %d: phase %s winner %q", round, r.phase, r.raceWinner)
		}
		if r.members[r.raceWinner].score != 1 {
			t.Fatalf("round %d: pemenang gak dapet poin", round)
		}
		r.seq++ // matiin timer reveal
		r.mu.Unlock()
	}
}

func TestRaceWrongAnswerLocksPlayer(t *testing.T) {
	r := testRoom(t, ModeRace, 3)
	if code := r.submitAnswer("u1", 0, 4); code != "" {
		t.Fatalf("jawaban salah pertama ditolak: %s", code)
	}
	if code := r.submitAnswer("u1", 0, 5); code != "already_answered" {
		t.Fatalf("pemain kekunci masih bisa jawab: %q", code)
	}
	r.submitAnswer("u0", 0, 7)
	r.submitAnswer("u2", 0, 6)

	r.mu.Lock()
	defer r.mu.Unlock()
	if r.phase != PhaseCooldown {
		t.Fatalf("semua salah tapi soal belum ditutup (phase %s)", r.phase)
	}
	if r.raceWinner != "" {
		t.Fatalf("gak ada yang benar tapi ada pemenang %q", r.raceWinner)
	}
	r.seq++
}

func TestClassicScoringAndStaleAnswer(t *testing.T) {
	r := testRoom(t, ModeClassic, 3)
	r.submitAnswer("u0", 0, 5)
	r.submitAnswer("u1", 0, 4)
	if code := r.submitAnswer("u2", 1, 42); code != "question_closed" {
		t.Fatalf("jawaban buat index soal lain harus ditolak, dapet %q", code)
	}

	r.mu.Lock()
	if r.phase != PhaseQuestion {
		t.Fatalf("classic harus nunggu waktu habis, phase %s", r.phase)
	}
	if s := r.members["u0"].score; s < classicMinPoints || s > classicMaxPoints {
		t.Fatalf("poin classic di luar rentang: %d", s)
	}
	if r.members["u1"].score != 0 {
		t.Fatalf("jawaban salah dapet poin")
	}
	r.endQuestionLocked()
	r.endGameLocked()
	if r.phase != PhaseFinished {
		t.Fatalf("classic harus langsung finished, phase %s", r.phase)
	}
	res := r.results
	r.seq++
	r.mu.Unlock()

	if len(res) != 3 || res[0].Player.UserID != "u0" || res[0].Rank != 1 {
		t.Fatalf("ranking salah: %+v", res)
	}
	if res[1].Rank != 2 || res[2].Rank != 2 {
		t.Fatalf("skor sama harus rank sama: %+v", res)
	}
}

func TestQuestionCountOptions(t *testing.T) {
	race := QuestionCountOptions(ModeRace)
	if race[0] != 5 || race[len(race)-1] != 95 || len(race) != 10 {
		t.Fatalf("race: %v", race)
	}
	for _, n := range race {
		if n%2 == 0 {
			t.Fatalf("jumlah soal race harus ganjil: %d", n)
		}
	}
	classic := QuestionCountOptions(ModeClassic)
	if classic[0] != 10 || classic[len(classic)-1] != 100 || len(classic) != 10 {
		t.Fatalf("classic: %v", classic)
	}
}

// Classic: semua pemain online udah jawab -> sisa waktu di-skip ke cooldown.
func TestClassicSkipsWhenEveryoneAnswered(t *testing.T) {
	r := testRoom(t, ModeClassic, 3)
	r.submitAnswer("u0", 0, 5)
	r.submitAnswer("u1", 0, 4)

	// u2 putus -> yang online (u0, u1) udah jawab semua.
	r.mu.Lock()
	c := r.members["u2"].conn
	r.mu.Unlock()
	r.detach("u2", c)

	r.mu.Lock()
	defer r.mu.Unlock()
	if r.phase != PhaseCooldown {
		t.Fatalf("semua yang online udah jawab tapi phase %s", r.phase)
	}
	r.seq++
}

// Semua pemain putus barengan -> jangan skip, tunggu timer.
func TestNoSkipWhenEveryoneOffline(t *testing.T) {
	r := testRoom(t, ModeClassic, 2)
	for _, id := range []string{"u0", "u1"} {
		r.mu.Lock()
		c := r.members[id].conn
		r.mu.Unlock()
		r.detach(id, c)
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.phase != PhaseQuestion {
		t.Fatalf("semua offline tapi soal ditutup (phase %s)", r.phase)
	}
	r.seq++
}

func TestRaceShowFastestReveal(t *testing.T) {
	r := testRoomWith(t, ModeRace, 3, true)
	r.submitAnswer("u1", 0, 5)

	r.mu.Lock()
	if r.phase != PhaseReveal {
		t.Fatalf("showFastest: harus pamer pemenang dulu, phase %s", r.phase)
	}
	v := r.viewForLocked("u2", time.Now())
	if v.Reveal == nil || v.Reveal.Winner == nil || v.Reveal.Winner.UserID != "u1" {
		t.Fatalf("pemenang harus keliatan: %+v", v.Reveal)
	}
	r.beginCooldownLocked()
	if r.phase != PhaseCooldown {
		t.Fatalf("abis reveal harus cooldown, phase %s", r.phase)
	}
	r.seq++
	r.mu.Unlock()
}

// Race tanpa ShowFastest: pemenang dirahasiain, hasil nunggu host.
func TestRaceHiddenResults(t *testing.T) {
	r := testRoomWith(t, ModeRace, 3, false)
	r.submitAnswer("u1", 0, 5)

	r.mu.Lock()
	v := r.viewForLocked("u2", time.Now())
	if v.Reveal == nil || !v.Reveal.HasWinner || v.Reveal.Winner != nil {
		t.Fatalf("pemenang harus dirahasiain: %+v", v.Reveal)
	}
	r.qIndex = len(r.questions) - 1
	r.endGameLocked()
	if r.phase != PhaseAwaitingResults {
		t.Fatalf("hasil harus nunggu host, phase %s", r.phase)
	}
	if v := r.viewForLocked("u1", time.Now()); v.Results != nil {
		t.Fatalf("hasil bocor sebelum dibuka host")
	}
	r.mu.Unlock()

	if code := r.revealResults("u1"); code != "forbidden" {
		t.Fatalf("bukan host bisa buka hasil: %q", code)
	}
	if code := r.revealResults("u0"); code != "" {
		t.Fatalf("host gagal buka hasil: %q", code)
	}
	r.mu.Lock()
	if r.phase != PhaseResultsCountdown {
		t.Fatalf("abis dibuka harus countdown, phase %s", r.phase)
	}
	r.finishLocked()
	if v := r.viewForLocked("u1", time.Now()); len(v.Results) != 3 {
		t.Fatalf("hasil harus keluar: %+v", v.Results)
	}
	r.seq++
	r.mu.Unlock()
}

// Kill room: host bisa nutup room di tengah game.
func TestHostClosesMidGame(t *testing.T) {
	r := testRoom(t, ModeClassic, 3)
	if code := r.closeByHost("u1"); code != "forbidden" {
		t.Fatalf("bukan host bisa nutup room: %q", code)
	}
	if code := r.closeByHost("u0"); code != "" {
		t.Fatalf("host gagal nutup room di tengah game: %q", code)
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.phase != PhaseClosed {
		t.Fatalf("phase %s, harusnya closed", r.phase)
	}
	for id, m := range r.members {
		if m.conn != nil {
			t.Fatalf("koneksi %s belum diputus", id)
		}
	}
}
