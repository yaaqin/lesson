// Package multiplayer: mode main bareng lewat WebSocket. State room disimpen
// in-memory (bukan DB) -- cukup selama API jalan 1 instance. Kalau nanti
// di-scale jadi beberapa instance, state room harus dipindah ke store bareng
// (mis. Redis) atau room di-pin ke 1 instance.
//
// Dua mode:
//   - classic: semua pemain dapet soal yang sama barengan. Jawab kapan aja
//     selama waktu soal jalan; soal ditutup pas waktunya habis ATAU semua
//     pemain yang online udah jawab. Benar = 500..1000 poin tergantung
//     kecepatan (dihitung dari jam server, bukan klaim klien).
//   - race: adu cepet. Jawaban benar PERTAMA dapet 1 poin dan soal langsung
//     ditutup. Salah = kekunci buat soal itu. Jumlah soal ganjil. Kalau
//     ShowFastest, pemenang soal dipamerin sebentar (reveal) sebelum
//     cooldown; kalau gak, nama pemenang dirahasiain dan hasil akhir baru
//     keluar setelah host klik "tampilkan hasil" (+ countdown).
//
// Antar soal selalu ada cooldown (lamanya diatur admin, lihat
// curriculumsvc.MultiplayerConfig). Ranking cuma ditampilin di akhir game;
// yang disimpen ke DB cuma rank tiap pemain buat kartu share.
package multiplayer

import (
	"errors"
	"regexp"
	"slices"
	"time"

	"lesson/api/internal/curriculumsvc"
)

const (
	ModeClassic = "classic"
	ModeRace    = "race"

	MaxPlayers = 10
	MinPlayers = 2
	maxBatches = 30

	classicMaxPoints = 1000
	classicMinPoints = 500

	countdownDuration = 3 * time.Second

	// Pemain yang putus di lobby dikeluarin kalau gak balik dalam waktu ini
	// (iOS mutus WebSocket tiap app ditinggal sebentar, jadi jangan langsung).
	lobbyDisconnectGrace = 45 * time.Second
	// Host yang ninggalin lobby selama ini -> room ditutup.
	lobbyHostAbsentLimit = 3 * time.Minute
	lobbyMaxAge          = 60 * time.Minute
	finishedRoomTTL      = 10 * time.Minute
	// Hasil yang dirahasiain (race tanpa ShowFastest) otomatis dibuka kalau
	// host gak ada selama lobbyHostAbsentLimit atau gak klik-klik selama ini.
	awaitingResultsLimit = 10 * time.Minute
	ticketTTL            = 30 * time.Second
)

// SecondsPerQuestionOptions: pilihan waktu per soal buat pembuat room.
var SecondsPerQuestionOptions = []int{10, 15, 20, 30, 45, 60}

var (
	ErrNotPremium         = errors.New("cuma user premium yang bisa bikin room")
	ErrInvalidSettings    = errors.New("pengaturan room gak valid")
	ErrInvalidSource      = errors.New("batch sumber soal gak valid")
	ErrRoomNotFound       = errors.New("room gak ketemu")
	ErrRoomFull           = errors.New("room udah penuh")
	ErrGameStarted        = errors.New("game udah mulai")
	ErrKicked             = errors.New("dikeluarin dari room")
	ErrAlreadyHosting     = errors.New("masih ada game kamu yang lagi jalan")
	ErrNicknameRequired   = curriculumsvc.ErrNicknameRequired
	ErrNotEnoughQuestions = curriculumsvc.ErrNotEnoughBank
	errInvalidTicket      = errors.New("ticket gak valid")
)

type Settings struct {
	Mode               string   `json:"mode"`
	QuestionCount      int      `json:"questionCount"`
	Format             string   `json:"format"`
	SecondsPerQuestion int      `json:"secondsPerQuestion"`
	BatchIDs           []string `json:"batchIds"`
	HostPlays          bool     `json:"hostPlays"`
	// ShowFastest: cuma kepake di race (classic selalu true).
	ShowFastest bool `json:"showFastest"`
}

// Timing: durasi jeda antar soal, snapshot dari config admin pas room dibikin.
type Timing struct {
	Cooldown         time.Duration
	WinnerReveal     time.Duration
	ResultsCountdown time.Duration
}

func timingFromConfig(c *curriculumsvc.MultiplayerConfig, mode string) Timing {
	cooldown := c.ClassicCooldownSeconds
	if mode == ModeRace {
		cooldown = c.RaceCooldownSeconds
	}
	return Timing{
		Cooldown:         time.Duration(cooldown) * time.Second,
		WinnerReveal:     time.Duration(c.RaceWinnerRevealSeconds) * time.Second,
		ResultsCountdown: time.Duration(c.ResultsCountdownSeconds) * time.Second,
	}
}

// QuestionCountOptions: classic 10, 20, ..., 100; race ganjil 5, 15, ..., 95.
func QuestionCountOptions(mode string) []int {
	start := 10
	if mode == ModeRace {
		start = 5
	}
	opts := []int{}
	for n := start; n <= 100; n += 10 {
		opts = append(opts, n)
	}
	return opts
}

var uuidPattern = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)

func (s *Settings) validate() error {
	if s.Mode != ModeClassic && s.Mode != ModeRace {
		return ErrInvalidSettings
	}
	if s.Mode == ModeClassic {
		s.ShowFastest = true
	}
	if !slices.Contains(QuestionCountOptions(s.Mode), s.QuestionCount) {
		return ErrInvalidSettings
	}
	switch s.Format {
	case curriculumsvc.MultiplayerFormatMC, curriculumsvc.MultiplayerFormatEssay, curriculumsvc.MultiplayerFormatMixed:
	default:
		return ErrInvalidSettings
	}
	if !slices.Contains(SecondsPerQuestionOptions, s.SecondsPerQuestion) {
		return ErrInvalidSettings
	}
	if len(s.BatchIDs) == 0 || len(s.BatchIDs) > maxBatches {
		return ErrInvalidSource
	}
	seen := map[string]bool{}
	ids := make([]string, 0, len(s.BatchIDs))
	for _, id := range s.BatchIDs {
		if !uuidPattern.MatchString(id) {
			return ErrInvalidSource
		}
		if !seen[id] {
			seen[id] = true
			ids = append(ids, id)
		}
	}
	s.BatchIDs = ids
	return nil
}
