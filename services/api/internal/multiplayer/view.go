package multiplayer

import (
	"time"

	"lesson/api/internal/curriculumsvc"
)

// View: snapshot state room buat SATU klien, dikirim utuh tiap ada perubahan
// (pemain <= 10, jadi murah). Klien tinggal render ulang dari snapshot --
// reconnect pun cukup nunggu snapshot pertama. Kunci jawaban & skor pemain
// lain gak pernah ikut sebelum waktunya.
type View struct {
	Type      string   `json:"type"` // "state"
	ServerNow int64    `json:"serverNow"`
	Room      RoomInfo `json:"room"`
	You       YouInfo  `json:"you"`
	Phase     string   `json:"phase"`
	// PhaseEndsAt: epoch ms kapan countdown/soal/reveal selesai (0 = gak ada).
	PhaseEndsAt   int64         `json:"phaseEndsAt"`
	Players       []PlayerView  `json:"players"`
	Question      *QuestionView `json:"question,omitempty"`
	AnsweredCount int           `json:"answeredCount"`
	PlayingCount  int           `json:"playingCount"`
	MyAnswer      *MyAnswerView `json:"myAnswer,omitempty"`
	Reveal        *RevealView   `json:"reveal,omitempty"`
	Results       []ResultEntry `json:"results,omitempty"`
	// MatchID: id hasil yang udah kesimpen (buat link share), cuma di finished.
	MatchID string `json:"matchId,omitempty"`
}

type RoomInfo struct {
	Code               string   `json:"code"`
	Mode               string   `json:"mode"`
	QuestionCount      int      `json:"questionCount"`
	Format             string   `json:"format"`
	SecondsPerQuestion int      `json:"secondsPerQuestion"`
	ShowFastest        bool     `json:"showFastest"`
	Sources            []string `json:"sources"`
	HostID             string   `json:"hostId"`
	MaxPlayers         int      `json:"maxPlayers"`
	MinPlayers         int      `json:"minPlayers"`
}

type YouInfo struct {
	UserID  string `json:"userId"`
	IsHost  bool   `json:"isHost"`
	Playing bool   `json:"playing"`
}

type PlayerView struct {
	curriculumsvc.PlayerProfile
	IsHost    bool `json:"isHost"`
	Playing   bool `json:"playing"`
	Connected bool `json:"connected"`
	Left      bool `json:"left"`
}

type QuestionView struct {
	Index    int       `json:"index"`
	Total    int       `json:"total"`
	TierCode string    `json:"tierCode"`
	Type     string    `json:"type"`
	Prompt   string    `json:"prompt"`
	Options  []float64 `json:"options,omitempty"`
}

// MyAnswerView: Correct keisi setelah soal ditutup (classic) atau langsung
// setelah jawab (race, biar yang salah tau dirinya kekunci).
type MyAnswerView struct {
	Value   float64 `json:"value"`
	Correct *bool   `json:"correct,omitempty"`
	Points  *int    `json:"points,omitempty"`
}

type RevealView struct {
	CorrectValue  float64 `json:"correctValue"`
	CorrectCount  int     `json:"correctCount"`
	AnsweredCount int     `json:"answeredCount"`
	// Race: ada yang dapet poin soal ini atau gak. Winner (siapa orangnya)
	// cuma dikirim kalau room-nya ShowFastest.
	HasWinner bool                         `json:"hasWinner"`
	Winner    *curriculumsvc.PlayerProfile `json:"winner,omitempty"`
}

type ResultEntry struct {
	Rank         int                         `json:"rank"`
	Player       curriculumsvc.PlayerProfile `json:"player"`
	Score        int                         `json:"score"`
	CorrectCount int                         `json:"correctCount"`
	AvgCorrectMs int64                       `json:"avgCorrectMs"`
	Left         bool                        `json:"left"`
	correctMs    int64
}

func (r *Room) viewForLocked(userID string, now time.Time) View {
	me := r.members[userID]
	v := View{
		Type:      "state",
		ServerNow: now.UnixMilli(),
		Room: RoomInfo{
			Code:               r.code,
			Mode:               r.settings.Mode,
			QuestionCount:      len(r.questions),
			Format:             r.settings.Format,
			SecondsPerQuestion: r.settings.SecondsPerQuestion,
			ShowFastest:        r.settings.ShowFastest,
			Sources:            r.sourceLabels,
			HostID:             r.hostID,
			MaxPlayers:         MaxPlayers,
			MinPlayers:         MinPlayers,
		},
		You:          YouInfo{UserID: userID, IsHost: userID == r.hostID, Playing: me != nil && me.playing},
		Phase:        r.phase,
		PlayingCount: r.playingCountLocked(),
	}
	if !r.phaseEndsAt.IsZero() {
		v.PhaseEndsAt = r.phaseEndsAt.UnixMilli()
	}
	for _, id := range r.joinOrder {
		m := r.members[id]
		v.Players = append(v.Players, PlayerView{
			PlayerProfile: m.profile,
			IsHost:        id == r.hostID,
			Playing:       m.playing,
			Connected:     m.conn != nil,
			Left:          m.left,
		})
	}

	closed := r.phase == PhaseReveal || r.phase == PhaseCooldown
	if r.phase == PhaseQuestion || closed {
		q := r.questions[r.qIndex]
		v.Question = &QuestionView{
			Index: r.qIndex, Total: len(r.questions), TierCode: q.TierCode,
			Type: q.Type, Prompt: q.Prompt, Options: q.Options,
		}
		v.AnsweredCount = len(r.answers)

		if a, ok := r.answers[userID]; ok {
			v.MyAnswer = &MyAnswerView{Value: a.value}
			if closed || r.settings.Mode == ModeRace {
				correct, points := a.correct, a.points
				v.MyAnswer.Correct = &correct
				v.MyAnswer.Points = &points
			}
		}

		if closed {
			rv := &RevealView{CorrectValue: q.CorrectValue, AnsweredCount: len(r.answers)}
			for _, a := range r.answers {
				if a.correct {
					rv.CorrectCount++
				}
			}
			if r.settings.Mode == ModeRace {
				// Race: yang dihitung benar cuma pemenangnya (jawaban benar
				// yang telat udah ditolak sebelum masuk r.answers).
				if w, ok := r.members[r.raceWinner]; ok {
					rv.HasWinner = true
					if r.settings.ShowFastest {
						winner := w.profile
						rv.Winner = &winner
					}
				}
			}
			v.Reveal = rv
		}
	}

	if r.phase == PhaseFinished {
		v.Results = r.results
		v.MatchID = r.matchID
	}
	return v
}

func (r *Room) broadcastLocked() {
	now := time.Now()
	for id, m := range r.members {
		if m.conn != nil {
			m.conn.sendJSON(r.viewForLocked(id, now))
		}
	}
}
