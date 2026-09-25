package multiplayer

import (
	"fmt"
	"sync"
	"testing"

	"lesson/api/internal/curriculumsvc"
)

func testRoom(t *testing.T, mode string, players int) *Room {
	t.Helper()
	questions := []curriculumsvc.MultiplayerQuestion{
		{ID: "q1", TierCode: "sd", Type: curriculumsvc.QuestionTypeMultipleChoice, Prompt: "2 + 3", Options: []float64{4, 5, 6, 7}, CorrectValue: 5},
		{ID: "q2", TierCode: "sd", Type: curriculumsvc.QuestionTypeEssayNumeric, Prompt: "6 x 7", CorrectValue: 42},
	}
	host := &curriculumsvc.PlayerProfile{UserID: "u0", Nickname: "host_0"}
	r := newRoom("TEST01", Settings{Mode: mode, QuestionCount: 5, Format: "mixed", SecondsPerQuestion: 20, HostPlays: true}, nil, questions, host)
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
		if r.phase != PhaseReveal || r.raceWinner == "" {
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
	if r.phase != PhaseReveal {
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
	r.finishLocked()
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
