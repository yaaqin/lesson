package curriculumsvc

import (
	"context"
	"errors"
	"math/rand/v2"

	"github.com/jackc/pgx/v5"
)

// Bagian DB dari mode multiplayer (game-nya sendiri jalan in-memory di
// internal/multiplayer). Sumber soal = batch kurikulum yang dipilih pembuat
// room, dibaca langsung dari tabel tiers/batches -- batch baru dari dashboard
// otomatis kepake tanpa ubah kode.

const (
	MultiplayerFormatMC    = "mc"
	MultiplayerFormatEssay = "essay"
	MultiplayerFormatMixed = "mixed"
)

type MultiplayerSourceBatch struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	// QuestionCount: soal yang kepake buat format essay/campuran (semua soal
	// PG bisa disajiin sebagai isian karena jawabannya angka). MCCount: yang
	// kepake buat format PG doang (soal essay asli gak punya opsi).
	QuestionCount int `json:"questionCount"`
	MCCount       int `json:"mcCount"`
}

type MultiplayerSourceTier struct {
	TierCode string                   `json:"tierCode"`
	TierName string                   `json:"tierName"`
	Batches  []MultiplayerSourceBatch `json:"batches"`
}

// ListMultiplayerSources: semua tier yang pake batch (tier puzzle kayak
// "umum" gak ikut) beserta batch-nya dan jumlah soal published di dalemnya.
func (s *Service) ListMultiplayerSources(ctx context.Context) ([]MultiplayerSourceTier, error) {
	rows, err := s.db.Query(ctx, `
		SELECT t.code, t.name, b.id, b.name,
			count(q.id),
			count(q.id) FILTER (WHERE q.question_type = 'multiple_choice')
		FROM tiers t
		JOIN batches b ON b.tier_id = t.id
		LEFT JOIN challenges c ON c.batch_id = b.id
		LEFT JOIN questions q ON q.challenge_id = c.id AND q.status = 'published'
			AND q.question_type IN ('multiple_choice', 'essay_numeric')
		WHERE t.uses_batch
		GROUP BY t.code, t.name, t.order_index, b.id, b.name, b.order_index
		ORDER BY t.order_index, b.order_index
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tiers := []MultiplayerSourceTier{}
	for rows.Next() {
		var tierCode, tierName string
		var b MultiplayerSourceBatch
		if err := rows.Scan(&tierCode, &tierName, &b.ID, &b.Name, &b.QuestionCount, &b.MCCount); err != nil {
			return nil, err
		}
		if b.QuestionCount == 0 {
			continue
		}
		if n := len(tiers); n == 0 || tiers[n-1].TierCode != tierCode {
			tiers = append(tiers, MultiplayerSourceTier{TierCode: tierCode, TierName: tierName})
		}
		last := &tiers[len(tiers)-1]
		last.Batches = append(last.Batches, b)
	}
	return tiers, rows.Err()
}

// MultiplayerQuestion: soal yang udah siap dimainin. Type = cara nyajiinnya
// (soal PG bisa jadi essay_numeric kalau format room isian/campuran), bukan
// tipe aslinya di bank. CorrectValue cuma dipegang server.
type MultiplayerQuestion struct {
	ID           string
	TierCode     string
	Type         string
	Prompt       string
	Options      []float64
	CorrectValue float64
}

type multiplayerBankQuestion struct {
	id, tierCode, batchID, qType, prompt string
	correct                              *float64
	options                              []QuestionOption
}

// PickMultiplayerQuestions: ambil `count` soal acak dari batch-batch yang
// dipilih, dibagi rata antar batch (round-robin) biar racikan campuran SD +
// SMP beneran kecampur, bukan didominasi batch yang bank soalnya paling gede.
// Soal dengan teks sama (bank procedural bisa dobel) cuma diambil sekali.
func (s *Service) PickMultiplayerQuestions(ctx context.Context, batchIDs []string, count int, format string) ([]MultiplayerQuestion, error) {
	rows, err := s.db.Query(ctx, `
		SELECT t.code, b.id, q.id, q.question_type, q.prompt, q.correct_answer_value,
			qo.option_value, qo.is_correct
		FROM questions q
		JOIN challenges c ON c.id = q.challenge_id
		JOIN batches b ON b.id = c.batch_id
		JOIN tiers t ON t.id = b.tier_id
		LEFT JOIN question_options qo ON qo.question_id = q.id
		WHERE b.id = ANY($1::uuid[]) AND q.status = 'published'
			AND q.question_type IN ('multiple_choice', 'essay_numeric')
		ORDER BY q.id
	`, batchIDs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	byID := map[string]*multiplayerBankQuestion{}
	order := []string{}
	for rows.Next() {
		var (
			tierCode, batchID, qID, qType, prompt string
			correct, optValue                     *float64
			isCorrect                             *bool
		)
		if err := rows.Scan(&tierCode, &batchID, &qID, &qType, &prompt, &correct, &optValue, &isCorrect); err != nil {
			return nil, err
		}
		q, ok := byID[qID]
		if !ok {
			q = &multiplayerBankQuestion{id: qID, tierCode: tierCode, batchID: batchID, qType: qType, prompt: prompt, correct: correct}
			byID[qID] = q
			order = append(order, qID)
		}
		if optValue != nil && isCorrect != nil {
			q.options = append(q.options, QuestionOption{Value: *optValue, IsCorrect: *isCorrect})
		}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	pools := map[string][]*multiplayerBankQuestion{}
	for _, id := range order {
		q := byID[id]
		if _, ok := multiplayerCorrectValue(q); !ok {
			continue
		}
		if format == MultiplayerFormatMC && (q.qType != QuestionTypeMultipleChoice || len(q.options) < 2) {
			continue
		}
		pools[q.batchID] = append(pools[q.batchID], q)
	}
	for _, pool := range pools {
		rand.Shuffle(len(pool), func(i, j int) { pool[i], pool[j] = pool[j], pool[i] })
	}

	picked := make([]*multiplayerBankQuestion, 0, count)
	seenPrompt := map[string]bool{}
	for len(picked) < count {
		progressed := false
		for _, batchID := range batchIDs {
			pool := pools[batchID]
			for len(pool) > 0 && len(picked) < count {
				q := pool[0]
				pool = pool[1:]
				if seenPrompt[q.prompt] {
					continue
				}
				seenPrompt[q.prompt] = true
				picked = append(picked, q)
				progressed = true
				break
			}
			pools[batchID] = pool
		}
		if !progressed {
			return nil, ErrNotEnoughBank
		}
	}
	rand.Shuffle(len(picked), func(i, j int) { picked[i], picked[j] = picked[j], picked[i] })

	// Campuran: separuh soal jadi isian. Soal essay asli udah pasti isian, sisa
	// jatahnya diambil dari soal PG yang disajiin tanpa opsi.
	essayQuota := 0
	switch format {
	case MultiplayerFormatEssay:
		essayQuota = count
	case MultiplayerFormatMixed:
		essayQuota = count / 2
		for _, q := range picked {
			if q.qType == QuestionTypeEssayNumeric || len(q.options) < 2 {
				essayQuota--
			}
		}
	}

	out := make([]MultiplayerQuestion, len(picked))
	for i, q := range picked {
		correct, _ := multiplayerCorrectValue(q)
		mq := MultiplayerQuestion{ID: q.id, TierCode: q.tierCode, Prompt: q.prompt, CorrectValue: correct}
		asEssay := format == MultiplayerFormatEssay || q.qType == QuestionTypeEssayNumeric || len(q.options) < 2
		if !asEssay && essayQuota > 0 {
			asEssay = true
			essayQuota--
		}
		if asEssay {
			mq.Type = QuestionTypeEssayNumeric
		} else {
			mq.Type = QuestionTypeMultipleChoice
			mq.Options = make([]float64, len(q.options))
			for j, opt := range q.options {
				mq.Options[j] = opt.Value
			}
			rand.Shuffle(len(mq.Options), func(a, b int) { mq.Options[a], mq.Options[b] = mq.Options[b], mq.Options[a] })
		}
		out[i] = mq
	}
	return out, nil
}

// multiplayerCorrectValue: soal PG -> nilai opsi yang is_correct (fallback
// correct_answer_value), soal essay -> correct_answer_value.
func multiplayerCorrectValue(q *multiplayerBankQuestion) (float64, bool) {
	for _, opt := range q.options {
		if opt.IsCorrect {
			return opt.Value, true
		}
	}
	if q.correct != nil {
		return *q.correct, true
	}
	return 0, false
}

// PlayerProfile: identitas publik yang ditampilin di room multiplayer.
type PlayerProfile struct {
	UserID    string     `json:"userId"`
	Nickname  string     `json:"nickname"`
	Avatar    UserAvatar `json:"avatar"`
	IsPremium bool       `json:"isPremium"`
}

var ErrNicknameRequired = errors.New("user belum punya nickname")

func (s *Service) GetPlayerProfile(ctx context.Context, userID string) (*PlayerProfile, error) {
	p := PlayerProfile{UserID: userID}
	var nickname *string
	err := s.db.QueryRow(ctx, `
		SELECT username, avatar_type, avatar_key, google_avatar_url, is_premium
		FROM users WHERE id = $1
	`, userID).Scan(&nickname, &p.Avatar.Type, &p.Avatar.Key, &p.Avatar.GoogleURL, &p.IsPremium)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	if nickname == nil {
		return nil, ErrNicknameRequired
	}
	p.Nickname = *nickname
	return &p, nil
}
