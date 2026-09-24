package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math/rand/v2"
	"sort"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"lesson/api/internal/config"
)

// Seed konten kurikulum: 3 tier (SD, SMP, SMK/SMA), masing-masing 1 batch,
// 5 challenge biasa + 1 challenge ujian, soal digenerate procedural (bukan
// hardcoded) biar variatif tiap kali database di-reset dari nol.
//
// Idempotent per-challenge: kalau sebuah challenge (by nama, dalam batch yang
// sama) udah ada, soalnya TIDAK digenerate ulang -- supaya rerun gak nimpa
// time_limit_seconds yang udah diedit admin dari dashboard.

type option struct {
	value     float64
	isCorrect bool
}

type question struct {
	prompt  string
	options []option
}

type challengeSpec struct {
	name                  string
	isExam                bool
	questionCountRequired int
	passThresholdPercent  int
	optionCount           int
	timeLimitSeconds      int
	questions             []question
}

type batchSpec struct {
	name       string
	challenges []challengeSpec
}

type tierSpec struct {
	code    string
	name    string
	batches []batchSpec
}

func main() {
	cfg := config.Load()
	ctx := context.Background()

	pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("connect db: %v", err)
	}
	defer pool.Close()

	tiers := []tierSpec{sdTierSpec(), smpTierSpec(), smkTierSpec(), kampusTierSpec()}

	for tierOrder, tier := range tiers {
		tierID, err := getOrCreateTier(ctx, pool, tier.code, tier.name, tierOrder)
		if err != nil {
			log.Fatalf("tier %s: %v", tier.code, err)
		}

		fmt.Printf("%s:\n", tier.name)
		for batchOrder, batch := range tier.batches {
			batchID, err := getOrCreateBatch(ctx, pool, tierID, batch.name, batchOrder)
			if err != nil {
				log.Fatalf("batch %s: %v", batch.name, err)
			}

			fmt.Printf("  %s:\n", batch.name)
			for order, spec := range batch.challenges {
				challengeID, existed, err := getOrCreateChallenge(ctx, pool, batchID, spec, order)
				if err != nil {
					log.Fatalf("challenge %s: %v", spec.name, err)
				}
				if existed {
					fmt.Printf("    - %s (sudah ada, dilewati)\n", spec.name)
					continue
				}
				if err := insertQuestions(ctx, pool, challengeID, spec.questions); err != nil {
					log.Fatalf("insert questions for %s: %v", spec.name, err)
				}
				kind := "challenge"
				if spec.isExam {
					kind = "ujian"
				}
				fmt.Printf("    - %s [%s]: %d soal di bank\n", spec.name, kind, len(spec.questions))
			}
		}
	}

	fmt.Println("\nSeed kurikulum selesai.")
}

func getOrCreateTier(ctx context.Context, pool *pgxpool.Pool, code, name string, orderIndex int) (string, error) {
	var id string
	err := pool.QueryRow(ctx, `SELECT id FROM tiers WHERE code = $1`, code).Scan(&id)
	if err == nil {
		return id, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return "", err
	}
	err = pool.QueryRow(ctx, `
		INSERT INTO tiers (code, name, uses_batch, order_index)
		VALUES ($1, $2, true, $3)
		RETURNING id
	`, code, name, orderIndex).Scan(&id)
	return id, err
}

func getOrCreateBatch(ctx context.Context, pool *pgxpool.Pool, tierID, name string, orderIndex int) (string, error) {
	var id string
	err := pool.QueryRow(ctx, `
		SELECT id FROM batches WHERE tier_id = $1 AND name = $2
	`, tierID, name).Scan(&id)
	if err == nil {
		return id, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return "", err
	}
	err = pool.QueryRow(ctx, `
		INSERT INTO batches (tier_id, name, order_index)
		VALUES ($1, $2, $3)
		RETURNING id
	`, tierID, name, orderIndex).Scan(&id)
	return id, err
}

func getOrCreateChallenge(ctx context.Context, pool *pgxpool.Pool, batchID string, spec challengeSpec, orderIndex int) (id string, existed bool, err error) {
	err = pool.QueryRow(ctx, `
		SELECT id FROM challenges WHERE batch_id = $1 AND name = $2
	`, batchID, spec.name).Scan(&id)
	if err == nil {
		return id, true, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return "", false, err
	}
	err = pool.QueryRow(ctx, `
		INSERT INTO challenges (
			batch_id, name, order_index, is_exam,
			question_count_required, pass_threshold_percent, option_count, time_limit_seconds
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id
	`, batchID, spec.name, orderIndex, spec.isExam,
		spec.questionCountRequired, spec.passThresholdPercent, spec.optionCount, spec.timeLimitSeconds,
	).Scan(&id)
	return id, false, err
}

func insertQuestions(ctx context.Context, pool *pgxpool.Pool, challengeID string, questions []question) error {
	for _, q := range questions {
		var questionID string
		err := pool.QueryRow(ctx, `
			INSERT INTO questions (challenge_id, question_type, prompt, correct_answer_value, status)
			VALUES ($1, 'multiple_choice', $2, $3, 'published')
			RETURNING id
		`, challengeID, q.prompt, correctValueOf(q.options)).Scan(&questionID)
		if err != nil {
			return err
		}
		for _, opt := range q.options {
			if _, err := pool.Exec(ctx, `
				INSERT INTO question_options (question_id, option_value, is_correct)
				VALUES ($1, $2, $3)
			`, questionID, opt.value, opt.isCorrect); err != nil {
				return err
			}
		}
	}
	return nil
}

func correctValueOf(opts []option) float64 {
	for _, o := range opts {
		if o.isCorrect {
			return o.value
		}
	}
	return 0
}

// buildChallenge generate `bankSize` soal dari generator `gen`, dipakai jadi
// bank soal satu challenge (question_count_required diambil random dari bank ini).
func buildChallenge(
	name string, isExam bool, required, bankSize, passPercent, optionCount, timeLimitSeconds int,
	gen func(optionCount int) question,
) challengeSpec {
	questions := make([]question, bankSize)
	for i := range questions {
		questions[i] = gen(optionCount)
	}
	return challengeSpec{
		name:                  name,
		isExam:                isExam,
		questionCountRequired: required,
		passThresholdPercent:  passPercent,
		optionCount:           optionCount,
		timeLimitSeconds:      timeLimitSeconds,
		questions:             questions,
	}
}

// buildOptions susun opsi jawaban: 1 benar + distraktor dari `candidates`
// (dedup, gak boleh sama dengan correct), diacak posisinya. Kalau candidates
// kurang, ditambal otomatis pakai offset dari correct yang belum kepake.
func buildOptions(correct float64, candidates []float64, optionCount int, allowNegative bool) []option {
	seen := map[float64]bool{correct: true}
	distinct := make([]float64, 0, len(candidates))
	for _, c := range candidates {
		if !allowNegative && c < 0 {
			continue
		}
		if !seen[c] {
			seen[c] = true
			distinct = append(distinct, c)
		}
	}
	rand.Shuffle(len(distinct), func(i, j int) { distinct[i], distinct[j] = distinct[j], distinct[i] })

	need := optionCount - 1
	if len(distinct) > need {
		distinct = distinct[:need]
	}
	for offset := 1.0; len(distinct) < need; offset++ {
		up := correct + offset
		if !seen[up] {
			seen[up] = true
			distinct = append(distinct, up)
		}
		if len(distinct) >= need {
			break
		}
		down := correct - offset
		if (allowNegative || down >= 0) && !seen[down] {
			seen[down] = true
			distinct = append(distinct, down)
		}
	}

	options := make([]option, 0, optionCount)
	// "+ 0" menormalkan -0 jadi 0 (biar gak tampil "-0" di UI)
	options = append(options, option{value: correct + 0, isCorrect: true})
	for _, d := range distinct {
		options = append(options, option{value: d + 0, isCorrect: false})
	}
	rand.Shuffle(len(options), func(i, j int) { options[i], options[j] = options[j], options[i] })
	return options
}

// ---------- SD ----------

func genAddition(optionCount int) question {
	a, b := rand.IntN(20)+1, rand.IntN(20)+1
	correct := float64(a + b)
	return question{
		prompt:  fmt.Sprintf("%d + %d", a, b),
		options: buildOptions(correct, []float64{correct - 1, correct + 1, correct - 2, correct + 2, correct + 3}, optionCount, false),
	}
}

func genSubtraction(optionCount int) question {
	a := rand.IntN(20) + 5
	b := rand.IntN(a) + 1
	correct := float64(a - b)
	return question{
		prompt:  fmt.Sprintf("%d - %d", a, b),
		options: buildOptions(correct, []float64{correct - 1, correct + 1, correct - 2, correct + 2, correct + 3}, optionCount, false),
	}
}

func genMultiplication(optionCount int) question {
	a, b := rand.IntN(10)+1, rand.IntN(10)+1
	correct := float64(a * b)
	return question{
		prompt:  fmt.Sprintf("%d × %d", a, b),
		options: buildOptions(correct, []float64{correct - float64(a), correct + float64(a), correct - float64(b), correct + float64(b), correct + 2}, optionCount, false),
	}
}

func genDivision(optionCount int) question {
	b := rand.IntN(10) + 1
	q := rand.IntN(10) + 1
	a := b * q
	correct := float64(q)
	return question{
		prompt:  fmt.Sprintf("%d ÷ %d", a, b),
		options: buildOptions(correct, []float64{correct - 1, correct + 1, correct - 2, correct + 2, correct + 3}, optionCount, false),
	}
}

func genSDMixed(optionCount int) question {
	switch rand.IntN(4) {
	case 0:
		return genAddition(optionCount)
	case 1:
		return genSubtraction(optionCount)
	case 2:
		return genMultiplication(optionCount)
	default:
		return genDivision(optionCount)
	}
}

func sdTierSpec() tierSpec {
	return tierSpec{
		code: "sd", name: "SD",
		batches: []batchSpec{
			{
				name: "Batch 1 — Semester 1",
				challenges: []challengeSpec{
					buildChallenge("Penjumlahan", false, 5, 8, 70, 3, 20, genAddition),
					buildChallenge("Pengurangan", false, 5, 8, 70, 3, 20, genSubtraction),
					buildChallenge("Perkalian", false, 5, 8, 70, 3, 20, genMultiplication),
					buildChallenge("Pembagian", false, 5, 8, 70, 3, 20, genDivision),
					buildChallenge("Campuran Dasar", false, 5, 8, 70, 3, 20, genSDMixed),
					buildChallenge("Ujian Semester 1 SD", true, 8, 12, 70, 3, 300, genSDMixed),
				},
			},
			{
				name:       "Batch 2 — Soal Campuran",
				challenges: sdCampuranChallenges(),
			},
			{
				name:       "Batch 3 — Campuran 4 Angka",
				challenges: sdCampuranFourDigitChallenges(),
			},
			{
				name:       "Batch 4 — Campuran 4 Variabel",
				challenges: sdCampuranFourVariableChallenges(),
			},
			{
				name:       "Batch 5 — Campuran 5 Variabel",
				challenges: sdCampuranFiveVariableChallenges(),
			},
		},
	}
}

// ---------- SD - Soal Campuran (batch 2, gacha gado-gado lebih susah) ----------
//
// Sama pola-nya kayak SMP/SMK batch 2: tiap level udah campur operasi (tambah-
// kurang dua langkah, urutan operasi kali-tambah / bagi-kurang, soal cerita
// gabungan), bank 16 soal per level, question_count_required 10 -> 10 dari 16
// diambil tiap main. Ujian banknya lebih gede: 35 soal, wajib jawab 25.

func genAddSubTwoStepHard(optionCount int) question {
	a := rand.IntN(31) + 20 // 20-50
	b := rand.IntN(21) + 10 // 10-30
	c := rand.IntN(a+b-1) + 1
	correct := float64(a + b - c)
	return question{
		prompt:  fmt.Sprintf("%d + %d - %d = ?", a, b, c),
		options: buildOptions(correct, []float64{correct - 1, correct + 1, correct - 2, correct + 2, correct + 3}, optionCount, false),
	}
}

func genMulAddHard(optionCount int) question {
	a, b := rand.IntN(8)+2, rand.IntN(8)+2
	c := rand.IntN(20) + 1
	correct := float64(a*b + c)
	return question{
		prompt:  fmt.Sprintf("%d × %d + %d = ?", a, b, c),
		options: buildOptions(correct, []float64{correct - 1, correct + 1, correct - float64(a), correct + float64(b), correct + 2}, optionCount, false),
	}
}

func genDivSubHard(optionCount int) question {
	b := rand.IntN(8) + 2
	q := rand.IntN(10) + 5
	a := b * q
	c := rand.IntN(q)
	correct := float64(q - c)
	return question{
		prompt:  fmt.Sprintf("%d ÷ %d - %d = ?", a, b, c),
		options: buildOptions(correct, []float64{correct - 1, correct + 1, correct - 2, correct + 2, correct + float64(b)}, optionCount, false),
	}
}

func genWordAddSubHard(optionCount int) question {
	names := []string{"Andi", "Budi", "Sari", "Dewi", "Rian"}
	name := names[rand.IntN(len(names))]
	start := rand.IntN(41) + 20 // 20-60
	gain := rand.IntN(21) + 5   // 5-25
	give := rand.IntN(start+gain-1) + 1
	correct := float64(start + gain - give)
	return question{
		prompt: fmt.Sprintf(
			"%s punya %d kelereng. Dia dapat %d kelereng lagi dari temannya, lalu memberikan %d kelereng ke adiknya. Sisa kelereng %s sekarang?",
			name, start, gain, give, name,
		),
		options: buildOptions(correct, []float64{correct - 1, correct + 1, correct - 2, correct + 2, correct + 5}, optionCount, false),
	}
}

func genWordMulDivHard(optionCount int) question {
	kids := []int{2, 3, 4, 5}[rand.IntN(4)]
	perKid := rand.IntN(8) + 3 // 3-10
	total := kids * perKid

	boxes := kids
	for _, d := range []int{8, 7, 6, 5, 4, 3, 2} {
		if total%d == 0 {
			boxes = d
			break
		}
	}
	perBox := total / boxes
	correct := float64(perKid)
	return question{
		prompt: fmt.Sprintf(
			"Sebuah toko punya %d kotak permen, tiap kotak isi %d permen. Semua permen dibagi rata ke %d anak. Setiap anak dapat berapa permen?",
			boxes, perBox, kids,
		),
		options: buildOptions(correct, []float64{correct - 1, correct + 1, correct - 2, correct + 2, float64(total)}, optionCount, false),
	}
}

func genSDCampuranMixed(optionCount int) question {
	switch rand.IntN(5) {
	case 0:
		return genAddSubTwoStepHard(optionCount)
	case 1:
		return genMulAddHard(optionCount)
	case 2:
		return genDivSubHard(optionCount)
	case 3:
		return genWordAddSubHard(optionCount)
	default:
		return genWordMulDivHard(optionCount)
	}
}

func sdCampuranChallenges() []challengeSpec {
	challenges := make([]challengeSpec, 0, 11)
	for level := 1; level <= 10; level++ {
		name := fmt.Sprintf("Campuran Level %d", level)
		challenges = append(challenges, buildChallenge(name, false, 10, 16, 70, 3, 25, genSDCampuranMixed))
	}
	challenges = append(challenges, buildChallenge("Ujian Soal Campuran SD", true, 25, 35, 70, 3, 900, genSDCampuranMixed))
	return challenges
}

// ---------- SD - Campuran 4 Angka (batch 3, angka digedein sampe ribuan) ----------
//
// Sama 5 tipe soal kayak batch 2 (tambah-kurang dua langkah, kali-tambah,
// bagi-kurang, cerita tambah-kurang, cerita kali-bagi), tapi rentang angkanya
// dinaikin sampe 4 digit (ribuan) biar lebih menantang. Struktur level & ujian
// tetep sama: 10 level (bank 16, wajib 10) + 1 ujian (bank 35, wajib 25).

func genAddSubTwoStepFourDigit(optionCount int) question {
	a := rand.IntN(4000) + 2000 // 2000-5999
	b := rand.IntN(3000) + 1000 // 1000-3999
	c := rand.IntN(a+b-1) + 1
	correct := float64(a + b - c)
	return question{
		prompt:  fmt.Sprintf("%d + %d - %d = ?", a, b, c),
		options: buildOptions(correct, []float64{correct - 1, correct + 1, correct - 2, correct + 2, correct + 3}, optionCount, false),
	}
}

func genMulAddFourDigit(optionCount int) question {
	a := rand.IntN(70) + 20 // 20-89
	b := rand.IntN(70) + 20 // 20-89
	c := rand.IntN(300) + 1 // 1-300
	correct := float64(a*b + c)
	return question{
		prompt:  fmt.Sprintf("%d × %d + %d = ?", a, b, c),
		options: buildOptions(correct, []float64{correct - 1, correct + 1, correct - float64(a), correct + float64(b), correct + 10}, optionCount, false),
	}
}

func genDivSubFourDigit(optionCount int) question {
	b := rand.IntN(8) + 2     // 2-9
	q := rand.IntN(700) + 300 // 300-999
	a := b * q
	c := rand.IntN(q)
	correct := float64(q - c)
	return question{
		prompt:  fmt.Sprintf("%d ÷ %d - %d = ?", a, b, c),
		options: buildOptions(correct, []float64{correct - 1, correct + 1, correct - 2, correct + 2, correct + float64(b)}, optionCount, false),
	}
}

func genWordAddSubFourDigit(optionCount int) question {
	start := rand.IntN(4000) + 2000 // 2000-5999 (stok awal)
	gain := rand.IntN(2000) + 500   // 500-2499 (kiriman baru)
	give := rand.IntN(start+gain-1) + 1
	correct := float64(start + gain - give)
	return question{
		prompt: fmt.Sprintf(
			"Sebuah toko ATK punya stok %d buku. Toko itu menerima kiriman baru %d buku, lalu %d buku terjual. Sisa stok buku di toko sekarang?",
			start, gain, give,
		),
		options: buildOptions(correct, []float64{correct - 1, correct + 1, correct - 2, correct + 2, correct + 10}, optionCount, false),
	}
}

func genWordMulDivFourDigit(optionCount int) question {
	kids := []int{4, 5, 8, 10}[rand.IntN(4)]
	perKid := rand.IntN(600) + 200 // 200-799
	total := kids * perKid

	boxes := kids
	for _, d := range []int{20, 16, 15, 12, 10, 8, 6, 5, 4, 3, 2} {
		if total%d == 0 {
			boxes = d
			break
		}
	}
	perBox := total / boxes
	correct := float64(perKid)
	return question{
		prompt: fmt.Sprintf(
			"Sebuah pabrik memproduksi %d dus permen, tiap dus isi %d permen. Semua permen itu dibagi rata ke %d toko. Setiap toko dapat berapa permen?",
			boxes, perBox, kids,
		),
		options: buildOptions(correct, []float64{correct - 1, correct + 1, correct - 2, correct + 2, float64(total)}, optionCount, false),
	}
}

func genSDCampuranFourDigitMixed(optionCount int) question {
	switch rand.IntN(5) {
	case 0:
		return genAddSubTwoStepFourDigit(optionCount)
	case 1:
		return genMulAddFourDigit(optionCount)
	case 2:
		return genDivSubFourDigit(optionCount)
	case 3:
		return genWordAddSubFourDigit(optionCount)
	default:
		return genWordMulDivFourDigit(optionCount)
	}
}

func sdCampuranFourDigitChallenges() []challengeSpec {
	challenges := make([]challengeSpec, 0, 11)
	for level := 1; level <= 10; level++ {
		name := fmt.Sprintf("Campuran 4 Angka Level %d", level)
		challenges = append(challenges, buildChallenge(name, false, 10, 16, 70, 3, 25, genSDCampuranFourDigitMixed))
	}
	challenges = append(challenges, buildChallenge("Ujian Campuran 4 Angka SD", true, 25, 35, 70, 3, 900, genSDCampuranFourDigitMixed))
	return challenges
}

// ---------- SD - Campuran 4 Variabel (batch 4, 4 angka + 3 operasi sekaligus) ----------
//
// Bukan naikin jumlah digit, tapi jumlah ANGKA yang dioperasikan dalam satu
// soal: a op b op c op d (contoh: 100 ÷ 4 × 75 + 333), ngikutin urutan operasi
// baku (kali/bagi didahulukan baru tambah/kurang). Opsi jawaban dinaikin jadi
// 4 pilihan. Struktur tetep 10 level (bank 16, wajib 10) + 1 ujian (bank 35,
// wajib 25).

func genDivMulAddFourVar(optionCount int) question {
	b := rand.IntN(8) + 2   // 2-9
	q := rand.IntN(50) + 10 // 10-59
	a := b * q              // a ÷ b = q (pas)
	c := rand.IntN(90) + 10 // 10-99
	d := rand.IntN(500) + 1 // 1-500
	correct := float64(q*c + d)
	return question{
		prompt:  fmt.Sprintf("%d ÷ %d × %d + %d = ?", a, b, c, d),
		options: buildOptions(correct, []float64{correct - 1, correct + 1, correct - 2, correct + 2, correct + 10, correct - 10}, optionCount, false),
	}
}

func genMulSubDivFourVar(optionCount int) question {
	a := rand.IntN(40) + 10 // 10-49
	b := rand.IntN(40) + 10 // 10-49
	product := a * b
	d := rand.IntN(8) + 2          // 2-9
	qc := rand.IntN(product-1) + 1 // 1..product-1
	c := d * qc                    // c ÷ d = qc (pas)
	correct := float64(product - qc)
	return question{
		prompt:  fmt.Sprintf("%d × %d - %d ÷ %d = ?", a, b, c, d),
		options: buildOptions(correct, []float64{correct - 1, correct + 1, correct - 2, correct + 2, correct + 10, correct - 10}, optionCount, false),
	}
}

func genAddMulSubFourVar(optionCount int) question {
	b := rand.IntN(40) + 10 // 10-49
	c := rand.IntN(40) + 10 // 10-49
	bc := b * c
	a := rand.IntN(1000) + 100 // 100-1099
	d := rand.IntN(a+bc-1) + 1
	correct := float64(a + bc - d)
	return question{
		prompt:  fmt.Sprintf("%d + %d × %d - %d = ?", a, b, c, d),
		options: buildOptions(correct, []float64{correct - 1, correct + 1, correct - 2, correct + 2, correct + 10, correct - 10}, optionCount, false),
	}
}

func genSubDivMulFourVar(optionCount int) question {
	c := rand.IntN(8) + 2   // 2-9
	q := rand.IntN(40) + 10 // 10-49
	b := c * q              // b ÷ c = q (pas)
	d := rand.IntN(20) + 2  // 2-21
	qd := q * d
	a := qd + rand.IntN(1000) + 1
	correct := float64(a - qd)
	return question{
		prompt:  fmt.Sprintf("%d - %d ÷ %d × %d = ?", a, b, c, d),
		options: buildOptions(correct, []float64{correct - 1, correct + 1, correct - 2, correct + 2, correct + 10, correct - 10}, optionCount, false),
	}
}

func genWordTwoMulAddFourVar(optionCount int) question {
	a := rand.IntN(15) + 5  // 5-19 karung toko A
	b := rand.IntN(40) + 10 // 10-49 kg per karung toko A
	c := rand.IntN(15) + 5  // 5-19 karung toko B
	d := rand.IntN(40) + 10 // 10-49 kg per karung toko B
	correct := float64(a*b + c*d)
	return question{
		prompt: fmt.Sprintf(
			"Toko A menjual %d karung beras, tiap karung isi %d kg. Toko B menjual %d karung beras, tiap karung isi %d kg. Total berat beras kedua toko (kg) = ?",
			a, b, c, d,
		),
		options: buildOptions(correct, []float64{correct - 1, correct + 1, correct - 2, correct + 2, float64(a * b), float64(c * d)}, optionCount, false),
	}
}

func genSDCampuranFourVariableMixed(optionCount int) question {
	switch rand.IntN(5) {
	case 0:
		return genDivMulAddFourVar(optionCount)
	case 1:
		return genMulSubDivFourVar(optionCount)
	case 2:
		return genAddMulSubFourVar(optionCount)
	case 3:
		return genSubDivMulFourVar(optionCount)
	default:
		return genWordTwoMulAddFourVar(optionCount)
	}
}

func sdCampuranFourVariableChallenges() []challengeSpec {
	challenges := make([]challengeSpec, 0, 11)
	for level := 1; level <= 10; level++ {
		name := fmt.Sprintf("Campuran 4 Variabel Level %d", level)
		challenges = append(challenges, buildChallenge(name, false, 10, 16, 70, 4, 30, genSDCampuranFourVariableMixed))
	}
	challenges = append(challenges, buildChallenge("Ujian Campuran 4 Variabel SD", true, 25, 35, 70, 4, 900, genSDCampuranFourVariableMixed))
	return challenges
}

// ---------- SD - Campuran 5 Variabel (batch 5, 5 angka + 4 operasi, sampe 5 digit) ----------
//
// Naik satu tingkat lagi dari batch 4: 5 angka dioperasikan sekaligus (4
// operasi), salah satu angkanya dibikin bisa nyampe 5 digit (10.000-99.999).
// Tetep ngikutin urutan operasi baku (kali/bagi didahulukan, dihitung
// kiri-ke-kanan kalau presedensinya sama). Opsi jawaban naik jadi 5 pilihan,
// bank per level naik jadi 25, wajib jawab 15. Ujian: bank 50, wajib 35.

func genDivMulAddSubFiveVar(optionCount int) question {
	b := rand.IntN(8) + 2       // 2-9
	q := rand.IntN(9000) + 1000 // 1000-9999
	a := b * q                  // a ÷ b = q (pas, a bisa nyampe 5 digit)
	c := rand.IntN(8) + 2       // 2-9
	afterMul := q * c
	d := rand.IntN(4901) + 100 // 100-5000
	running := afterMul + d
	e := rand.IntN(running-1) + 1
	correct := float64(running - e)
	return question{
		prompt:  fmt.Sprintf("%d ÷ %d × %d + %d - %d = ?", a, b, c, d, e),
		options: buildOptions(correct, []float64{correct - 1, correct + 1, correct - 2, correct + 2, correct + 10, correct - 10, correct + 20}, optionCount, false),
	}
}

func genMulAddSubDivFiveVar(optionCount int) question {
	a := rand.IntN(90) + 10    // 10-99
	b := rand.IntN(90) + 10    // 10-99
	c := rand.IntN(8901) + 100 // 100-9000
	running := a*b + c
	e := rand.IntN(8) + 2 // 2-9
	qd := rand.IntN(min(running-1, 9999)) + 1
	d := e * qd // d ÷ e = qd (pas, d bisa nyampe 5 digit)
	correct := float64(running - qd)
	return question{
		prompt:  fmt.Sprintf("%d × %d + %d - %d ÷ %d = ?", a, b, c, d, e),
		options: buildOptions(correct, []float64{correct - 1, correct + 1, correct - 2, correct + 2, correct + 10, correct - 10, correct + 20}, optionCount, false),
	}
}

func genAddMulSubDivFiveVar(optionCount int) question {
	a := rand.IntN(8001) + 1000 // 1000-9000
	b := rand.IntN(90) + 10     // 10-99
	c := rand.IntN(90) + 10     // 10-99
	running := a + b*c
	e := rand.IntN(8) + 2 // 2-9
	qde := rand.IntN(min(running-1, 9999)) + 1
	d := e * qde // d ÷ e = qde (pas, d bisa nyampe 5 digit)
	correct := float64(running - qde)
	return question{
		prompt:  fmt.Sprintf("%d + %d × %d - %d ÷ %d = ?", a, b, c, d, e),
		options: buildOptions(correct, []float64{correct - 1, correct + 1, correct - 2, correct + 2, correct + 10, correct - 10, correct + 20}, optionCount, false),
	}
}

func genSubDivMulAddFiveVar(optionCount int) question {
	c := rand.IntN(8) + 2       // 2-9
	q := rand.IntN(9000) + 1000 // 1000-9999
	b := c * q                  // b ÷ c = q (pas, b bisa nyampe 5 digit)
	d := rand.IntN(8) + 2       // 2-9
	qd := q * d
	a := qd + rand.IntN(5000) + 1
	e := rand.IntN(4901) + 100 // 100-5000
	correct := float64(a - qd + e)
	return question{
		prompt:  fmt.Sprintf("%d - %d ÷ %d × %d + %d = ?", a, b, c, d, e),
		options: buildOptions(correct, []float64{correct - 1, correct + 1, correct - 2, correct + 2, correct + 10, correct - 10, correct + 20}, optionCount, false),
	}
}

func genWordWarehouseFiveVar(optionCount int) question {
	start := rand.IntN(40001) + 10000 // 10000-50000 (stok awal, 5 digit)
	gainA := rand.IntN(8001) + 1000   // 1000-9000
	gainB := rand.IntN(4501) + 500    // 500-5000
	running1 := start + gainA + gainB
	out := rand.IntN(running1-1) + 1
	running2 := running1 - out
	damaged := rand.IntN(running2 + 1) // 0..running2
	correct := float64(running2 - damaged)
	return question{
		prompt: fmt.Sprintf(
			"Sebuah gudang punya stok awal %d karung beras. Gudang menerima kiriman %d karung dari pemasok A dan %d karung dari pemasok B, lalu %d karung dikirim ke toko lain dan %d karung rusak harus dibuang. Sisa karung beras di gudang sekarang?",
			start, gainA, gainB, out, damaged,
		),
		options: buildOptions(correct, []float64{correct - 1, correct + 1, correct - 2, correct + 2, correct + 10, correct - 10, correct + 20}, optionCount, false),
	}
}

func genSDCampuranFiveVariableMixed(optionCount int) question {
	switch rand.IntN(5) {
	case 0:
		return genDivMulAddSubFiveVar(optionCount)
	case 1:
		return genMulAddSubDivFiveVar(optionCount)
	case 2:
		return genAddMulSubDivFiveVar(optionCount)
	case 3:
		return genSubDivMulAddFiveVar(optionCount)
	default:
		return genWordWarehouseFiveVar(optionCount)
	}
}

func sdCampuranFiveVariableChallenges() []challengeSpec {
	challenges := make([]challengeSpec, 0, 11)
	for level := 1; level <= 10; level++ {
		name := fmt.Sprintf("Campuran 5 Variabel Level %d", level)
		challenges = append(challenges, buildChallenge(name, false, 15, 25, 70, 5, 40, genSDCampuranFiveVariableMixed))
	}
	challenges = append(challenges, buildChallenge("Ujian Campuran 5 Variabel SD", true, 35, 50, 70, 5, 1260, genSDCampuranFiveVariableMixed))
	return challenges
}

// ---------- SMP ----------

func genLinearEquation(optionCount int) question {
	x := rand.IntN(21) - 10
	a := rand.IntN(20) + 1
	b := a + x
	correct := float64(x)
	return question{
		prompt:  fmt.Sprintf("x + %d = %d, x = ?", a, b),
		options: buildOptions(correct, []float64{correct - 1, correct + 1, correct - 2, correct + 2, correct + 3}, optionCount, true),
	}
}

func gcd(a, b int) int {
	for b != 0 {
		a, b = b, a%b
	}
	if a < 0 {
		return -a
	}
	return a
}

func genFractionSimplify(optionCount int) question {
	rn := rand.IntN(4) + 1
	rd := rn + rand.IntN(4) + 1
	for gcd(rn, rd) != 1 {
		rd++
	}
	k := rand.IntN(3) + 2
	num, den := rn*k, rd*k
	correct := float64(rn)
	return question{
		prompt:  fmt.Sprintf("Sederhanakan %d/%d, pembilangnya jadi?", num, den),
		options: buildOptions(correct, []float64{correct - 1, correct + 1, correct + 2, float64(num), float64(den)}, optionCount, false),
	}
}

func genPercentage(optionCount int) question {
	percents := []int{10, 20, 25, 50, 75}
	p := percents[rand.IntN(len(percents))]
	base := (rand.IntN(10) + 1) * 4
	correct := float64(p * base / 100)
	return question{
		prompt:  fmt.Sprintf("%d%% dari %d = ?", p, base),
		options: buildOptions(correct, []float64{correct - 1, correct + 1, correct - 2, correct + 2, correct + 5}, optionCount, false),
	}
}

func genRatio(optionCount int) question {
	a, b := rand.IntN(5)+1, rand.IntN(5)+1
	m := rand.IntN(5) + 2
	c := a * m
	correct := float64(b * m)
	return question{
		prompt:  fmt.Sprintf("%d : %d = %d : n, n = ?", a, b, c),
		options: buildOptions(correct, []float64{correct - 1, correct + 1, correct - float64(m), correct + float64(m), correct + 2}, optionCount, false),
	}
}

func genIntegerOps(optionCount int) question {
	a, b := rand.IntN(30)+1, rand.IntN(30)+1
	correct := float64(a - b)
	return question{
		prompt:  fmt.Sprintf("%d - %d", a, b),
		options: buildOptions(correct, []float64{correct - 1, correct + 1, correct - 2, correct + 2, -correct}, optionCount, true),
	}
}

func genSMPMixed(optionCount int) question {
	switch rand.IntN(5) {
	case 0:
		return genLinearEquation(optionCount)
	case 1:
		return genFractionSimplify(optionCount)
	case 2:
		return genPercentage(optionCount)
	case 3:
		return genRatio(optionCount)
	default:
		return genIntegerOps(optionCount)
	}
}

func smpTierSpec() tierSpec {
	return tierSpec{
		code: "smp", name: "SMP",
		batches: []batchSpec{
			{
				name: "Batch 1 — Semester 1",
				challenges: []challengeSpec{
					buildChallenge("Aljabar Dasar", false, 5, 8, 70, 4, 25, genLinearEquation),
					buildChallenge("Pecahan", false, 5, 8, 70, 4, 25, genFractionSimplify),
					buildChallenge("Persentase", false, 5, 8, 70, 4, 25, genPercentage),
					buildChallenge("Perbandingan", false, 5, 8, 70, 4, 25, genRatio),
					buildChallenge("Bilangan Bulat", false, 5, 8, 70, 4, 25, genIntegerOps),
					buildChallenge("Ujian Semester 1 SMP", true, 8, 12, 70, 4, 480, genSMPMixed),
				},
			},
			{
				name:       "Batch 2 — Soal Campuran",
				challenges: smpCampuranChallenges(),
			},
			{
				name:       "Batch 3 — Campuran 3 Variabel",
				challenges: smpCampuranThreeVariableChallenges(),
			},
			{
				name:       "Batch 4 — Campuran 4 Variabel",
				challenges: smpCampuranFourVariableChallenges(),
			},
			{
				name:       "Batch 5 — Campuran 5 Variabel",
				challenges: smpCampuranFiveVariableChallenges(),
			},
		},
	}
}

// ---------- SMP - Soal Campuran (batch 2, gacha gado-gado lebih susah) ----------
//
// Bukan per-topik kayak batch 1 -- tiap level udah campur (aljabar 2 langkah,
// pecahan berpenyebut beda, persentase naik/turun, rasio soal cerita, operasi
// hitung campuran), banknya lebih gede (16 per level) biar variasi gacha-nya
// berasa, question_count_required 10 -> pemain harus jawab 10 dari 16 tiap main.

func genTwoStepLinearHard(optionCount int) question {
	x := rand.IntN(21) - 10
	a := rand.IntN(8) + 2
	b := rand.IntN(31) - 15
	c := a*x + b
	correct := float64(x)
	return question{
		prompt:  fmt.Sprintf("%dx + %d = %d, x = ?", a, b, c),
		options: buildOptions(correct, []float64{correct - 1, correct + 1, correct - 2, correct + 2, correct + 3}, optionCount, true),
	}
}

func genFractionSumHard(optionCount int) question {
	denoms := []int{2, 3, 4, 5, 6, 8}
	bd := denoms[rand.IntN(len(denoms))]
	dd := denoms[rand.IntN(len(denoms))]
	bn := rand.IntN(bd-1) + 1
	dn := rand.IntN(dd-1) + 1
	num := bn*dd + dn*bd
	den := bd * dd
	g := gcd(num, den)
	if g == 0 {
		g = 1
	}
	simpNum, simpDen := num/g, den/g
	correct := float64(simpNum)
	return question{
		prompt:  fmt.Sprintf("%d/%d + %d/%d, hasilnya disederhanakan, pembilangnya jadi?", bn, bd, dn, dd),
		options: buildOptions(correct, []float64{correct - 1, correct + 1, correct + 2, float64(simpDen), float64(num)}, optionCount, false),
	}
}

func genPercentageChangeHard(optionCount int) question {
	percents := []int{10, 20, 25, 50}
	p := percents[rand.IntN(len(percents))]
	base := (rand.IntN(8) + 2) * 100
	up := rand.IntN(2) == 0
	var correct float64
	verb := "naik"
	if up {
		correct = float64(base) * float64(100+p) / 100
	} else {
		verb = "turun"
		correct = float64(base) * float64(100-p) / 100
	}
	return question{
		prompt:  fmt.Sprintf("Harga barang Rp%d %s %d%%. Harga sekarang jadi Rp?", base, verb, p),
		options: buildOptions(correct, []float64{correct - 100, correct + 100, correct - 50, correct + 50, float64(base)}, optionCount, false),
	}
}

func genRatioWordHard(optionCount int) question {
	a, b := rand.IntN(6)+1, rand.IntN(6)+1
	for gcd(a, b) != 1 {
		b = rand.IntN(6) + 1
	}
	parts := a + b
	k := rand.IntN(5) + 2
	total := parts * k * 10
	shareA := total * a / parts
	correct := float64(shareA)
	return question{
		prompt:  fmt.Sprintf("Perbandingan uang Andi dan Budi adalah %d : %d. Jika jumlah uang mereka Rp%d, uang Andi = Rp?", a, b, total),
		options: buildOptions(correct, []float64{correct - 10, correct + 10, correct - 20, correct + 20, float64(total - shareA)}, optionCount, false),
	}
}

func genIntegerOrderOpsHard(optionCount int) question {
	a := rand.IntN(15) + 1
	b := rand.IntN(9) + 2
	c := rand.IntN(20) + 1
	correct := float64(a - b*c)
	return question{
		prompt:  fmt.Sprintf("%d - %d × %d = ?", a, b, c),
		options: buildOptions(correct, []float64{correct - 1, correct + 1, correct + float64(b), correct - float64(b), -correct}, optionCount, true),
	}
}

func genSMPCampuranMixed(optionCount int) question {
	switch rand.IntN(5) {
	case 0:
		return genTwoStepLinearHard(optionCount)
	case 1:
		return genFractionSumHard(optionCount)
	case 2:
		return genPercentageChangeHard(optionCount)
	case 3:
		return genRatioWordHard(optionCount)
	default:
		return genIntegerOrderOpsHard(optionCount)
	}
}

func smpCampuranChallenges() []challengeSpec {
	challenges := make([]challengeSpec, 0, 11)
	for level := 1; level <= 10; level++ {
		name := fmt.Sprintf("Campuran Level %d", level)
		challenges = append(challenges, buildChallenge(name, false, 10, 16, 70, 4, 30, genSMPCampuranMixed))
	}
	challenges = append(challenges, buildChallenge("Ujian Soal Campuran SMP", true, 10, 16, 70, 4, 600, genSMPCampuranMixed))
	return challenges
}

// ---------- SMP - Campuran 3 Variabel (batch 3, 3 angka + 2 operasi, salah satunya ratusan) ----------
//
// Tiap soal ngoperasiin 3 angka sekaligus (2 operasi), salah satu angkanya
// dibikin di rentang ratusan (100-999). Level SMP: ngikutin urutan operasi baku,
// ada kurung, dan hasilnya boleh negatif (bilangan bulat). Separuh soal langsung
// (ekspresi), separuh soal cerita. Distraktor sengaja masukin jawaban "salah
// urutan operasi" (dihitung kiri-ke-kanan) biar kejebak kalau asal hitung.
// Struktur: 10 level (bank 23, wajib 15) + 1 ujian (bank 35, wajib 25).

func genAddMulThreeVar(optionCount int) question {
	a := rand.IntN(900) + 100 // 100-999
	b := rand.IntN(11) + 2    // 2-12
	c := rand.IntN(41) + 10   // 10-50
	correct := float64(a + b*c)
	wrongOrder := float64((a + b) * c)
	return question{
		prompt:  fmt.Sprintf("%d + %d × %d = ?", a, b, c),
		options: buildOptions(correct, []float64{wrongOrder, correct - 1, correct + 1, correct - 10, correct + 10, correct + float64(c)}, optionCount, false),
	}
}

func genSubMulThreeVar(optionCount int) question {
	a := rand.IntN(900) + 100   // 100-999
	b := rand.IntN(16) + 5      // 5-20
	c := rand.IntN(51) + 10     // 10-60
	correct := float64(a - b*c) // bisa negatif
	wrongOrder := float64((a - b) * c)
	return question{
		prompt:  fmt.Sprintf("%d - %d × %d = ?", a, b, c),
		options: buildOptions(correct, []float64{wrongOrder, -correct, correct - 1, correct + 1, correct - 10, correct + 10}, optionCount, true),
	}
}

func genDivSubThreeVar(optionCount int) question {
	b := rand.IntN(8) + 2 // 2-9
	q := rand.IntN(999/b-100/b) + 100/b + 1
	a := b * q                // a ÷ b = q (pas), a di rentang ratusan
	c := rand.IntN(q+50) + 10 // bisa lebih gede dari q -> hasil negatif
	correct := float64(q - c)
	return question{
		prompt:  fmt.Sprintf("%d ÷ %d - %d = ?", a, b, c),
		options: buildOptions(correct, []float64{-correct, correct - 1, correct + 1, correct - 2, correct + 2, float64(a - c)}, optionCount, true),
	}
}

func genParenSubMulThreeVar(optionCount int) question {
	a := rand.IntN(900) + 100 // 100-999
	b := rand.IntN(90) + 10   // 10-99
	c := rand.IntN(8) + 2     // 2-9
	correct := float64((a - b) * c)
	noParen := float64(a - b*c)
	return question{
		prompt:  fmt.Sprintf("(%d - %d) × %d = ?", a, b, c),
		options: buildOptions(correct, []float64{noParen, correct - float64(c), correct + float64(c), correct - 10, correct + 10}, optionCount, true),
	}
}

func genWordLibraryThreeVar(optionCount int) question {
	start := rand.IntN(900) + 100 // 100-999 buku awal
	boxes := rand.IntN(11) + 2    // 2-12 kardus
	perBox := rand.IntN(41) + 10  // 10-50 buku per kardus
	correct := float64(start + boxes*perBox)
	return question{
		prompt: fmt.Sprintf(
			"Perpustakaan sekolah punya %d buku. Datang kiriman %d kardus, tiap kardus isi %d buku. Jumlah buku di perpustakaan sekarang?",
			start, boxes, perBox,
		),
		options: buildOptions(correct, []float64{float64((start + boxes) * perBox), float64(start + boxes + perBox), correct - 10, correct + 10, correct - 1}, optionCount, false),
	}
}

func genWordSubmarineThreeVar(optionCount int) question {
	depth := rand.IntN(900) + 100             // 100-999 m di bawah permukaan laut
	rate := rand.IntN(8) + 2                  // 2-9 m per menit
	minutes := rand.IntN(26) + 5              // 5-30 menit
	correct := float64(-depth + rate*minutes) // bisa masih negatif (di bawah laut)
	return question{
		prompt: fmt.Sprintf(
			"Sebuah kapal selam berada di kedalaman %d meter di bawah permukaan laut. Kapal itu naik %d meter tiap menit selama %d menit. Posisi kapal selam sekarang (meter, negatif = di bawah permukaan laut)?",
			depth, rate, minutes,
		),
		options: buildOptions(correct, []float64{-correct, float64(-depth - rate*minutes), correct - float64(rate), correct + float64(rate), correct - 10}, optionCount, true),
	}
}

func genWordSnackThreeVar(optionCount int) question {
	classes := rand.IntN(8) + 2 // 2-9 kelas
	perClass := rand.IntN(999/classes-100/classes) + 100/classes + 1
	total := classes * perClass // total snack di rentang ratusan, kebagi pas
	given := rand.IntN(perClass-1) + 1
	correct := float64(perClass - given)
	return question{
		prompt: fmt.Sprintf(
			"Panitia punya %d kotak snack yang dibagi rata ke %d kelas. Setiap kelas lalu memberikan %d kotak ke wali kelasnya. Sisa kotak snack tiap kelas?",
			total, classes, given,
		),
		options: buildOptions(correct, []float64{float64(total - given), float64(perClass), correct - 1, correct + 1, correct + 2}, optionCount, false),
	}
}

func genSMPCampuranThreeVariableMixed(optionCount int) question {
	// 50:50 soal langsung vs soal cerita.
	if rand.IntN(2) == 0 {
		switch rand.IntN(4) {
		case 0:
			return genAddMulThreeVar(optionCount)
		case 1:
			return genSubMulThreeVar(optionCount)
		case 2:
			return genDivSubThreeVar(optionCount)
		default:
			return genParenSubMulThreeVar(optionCount)
		}
	}
	switch rand.IntN(3) {
	case 0:
		return genWordLibraryThreeVar(optionCount)
	case 1:
		return genWordSubmarineThreeVar(optionCount)
	default:
		return genWordSnackThreeVar(optionCount)
	}
}

func smpCampuranThreeVariableChallenges() []challengeSpec {
	challenges := make([]challengeSpec, 0, 11)
	for level := 1; level <= 10; level++ {
		name := fmt.Sprintf("Campuran 3 Variabel Level %d", level)
		challenges = append(challenges, buildChallenge(name, false, 15, 23, 70, 4, 30, genSMPCampuranThreeVariableMixed))
	}
	challenges = append(challenges, buildChallenge("Ujian Campuran 3 Variabel SMP", true, 25, 35, 70, 4, 900, genSMPCampuranThreeVariableMixed))
	return challenges
}

// ---------- SMP - Campuran 4 Variabel (batch 4, 4 angka + 3 operasi, salah satunya ratusan) ----------
//
// Naik satu tingkat dari batch 3: 4 angka dioperasikan sekaligus (3 operasi),
// minimal satu angka di rentang ratusan (100-999). Tetep gaya SMP: urutan operasi
// baku, ada kurung, hasil boleh negatif. 50:50 soal langsung vs soal cerita.
// Opsi jawaban naik jadi 5. Struktur: 10 level (bank 23, wajib 15) + 1 ujian
// (bank 35, wajib 25).

func genAddMulSubFourVarSMP(optionCount int) question {
	a := rand.IntN(900) + 100       // 100-999
	b := rand.IntN(11) + 2          // 2-12
	c := rand.IntN(41) + 10         // 10-50
	d := rand.IntN(900) + 100       // 100-999
	correct := float64(a + b*c - d) // bisa negatif
	wrongOrder := float64((a+b)*c - d)
	return question{
		prompt:  fmt.Sprintf("%d + %d × %d - %d = ?", a, b, c, d),
		options: buildOptions(correct, []float64{wrongOrder, -correct, correct - 1, correct + 1, correct - 10, correct + 10}, optionCount, true),
	}
}

func genSubMulDivFourVarSMP(optionCount int) question {
	a := rand.IntN(900) + 100   // 100-999
	b := rand.IntN(16) + 5      // 5-20
	d := rand.IntN(8) + 2       // 2-9
	k := rand.IntN(26) + 5      // 5-30
	c := d * k                  // b × c ÷ d = b × k (pas)
	correct := float64(a - b*k) // bisa negatif
	return question{
		prompt:  fmt.Sprintf("%d - %d × %d ÷ %d = ?", a, b, c, d),
		options: buildOptions(correct, []float64{float64(a - b*c), -correct, correct - float64(b), correct + float64(b), correct - 10, correct + 10}, optionCount, true),
	}
}

func genParenSubMulAddFourVarSMP(optionCount int) question {
	a := rand.IntN(900) + 100 // 100-999
	b := rand.IntN(90) + 10   // 10-99
	c := rand.IntN(8) + 2     // 2-9
	d := rand.IntN(90) + 10   // 10-99
	correct := float64((a-b)*c + d)
	noParen := float64(a - b*c + d)
	return question{
		prompt:  fmt.Sprintf("(%d - %d) × %d + %d = ?", a, b, c, d),
		options: buildOptions(correct, []float64{noParen, float64((a - b) * (c + d)), correct - float64(c), correct + float64(c), correct - 10, correct + 10}, optionCount, true),
	}
}

func genDivMulSubFourVarSMP(optionCount int) question {
	b := rand.IntN(8) + 2 // 2-9
	q := rand.IntN(999/b-100/b) + 100/b + 1
	a := b * q                  // a ÷ b = q (pas), a di rentang ratusan
	c := rand.IntN(8) + 2       // 2-9
	d := rand.IntN(900) + 100   // 100-999
	correct := float64(q*c - d) // bisa negatif
	return question{
		prompt:  fmt.Sprintf("%d ÷ %d × %d - %d = ?", a, b, c, d),
		options: buildOptions(correct, []float64{float64(q*c + d), -correct, correct - 1, correct + 1, correct - 10, correct + 10}, optionCount, true),
	}
}

func genMulSubMulFourVarSMP(optionCount int) question {
	a := rand.IntN(21) + 10       // 10-30
	b := rand.IntN(21) + 10       // 10-30
	c := rand.IntN(900) + 100     // 100-999
	d := rand.IntN(8) + 2         // 2-9
	correct := float64(a*b - c*d) // sering negatif
	wrongOrder := float64((a*b - c) * d)
	return question{
		prompt:  fmt.Sprintf("%d × %d - %d × %d = ?", a, b, c, d),
		options: buildOptions(correct, []float64{wrongOrder, -correct, correct - 1, correct + 1, correct - 10, correct + 10}, optionCount, true),
	}
}

func genWordSubmarineFourVarSMP(optionCount int) question {
	depth := rand.IntN(900) + 100 // 100-999 m di bawah permukaan laut
	rate := rand.IntN(8) + 2      // 2-9 m per menit
	minutes := rand.IntN(26) + 5  // 5-30 menit
	dive := rand.IntN(91) + 10    // 10-100 m
	correct := float64(-depth + rate*minutes - dive)
	return question{
		prompt: fmt.Sprintf(
			"Sebuah kapal selam berada di kedalaman %d meter di bawah permukaan laut. Kapal itu naik %d meter tiap menit selama %d menit, lalu menyelam lagi %d meter. Posisi kapal selam sekarang (meter, negatif = di bawah permukaan laut)?",
			depth, rate, minutes, dive,
		),
		options: buildOptions(correct, []float64{-correct, float64(-depth + rate*minutes + dive), float64(-depth - rate*minutes - dive), correct - float64(rate), correct + 10}, optionCount, true),
	}
}

func genWordShopFourVarSMP(optionCount int) question {
	start := rand.IntN(900) + 100 // 100-999 stok awal
	boxes := rand.IntN(11) + 2    // 2-12 kardus
	perBox := rand.IntN(41) + 10  // 10-50 per kardus
	sold := rand.IntN(start+boxes*perBox-1) + 1
	correct := float64(start + boxes*perBox - sold)
	return question{
		prompt: fmt.Sprintf(
			"Sebuah toko punya stok %d botol minuman. Toko itu menerima %d kardus baru, tiap kardus isi %d botol, lalu %d botol terjual. Sisa stok botol minuman sekarang?",
			start, boxes, perBox, sold,
		),
		options: buildOptions(correct, []float64{float64((start+boxes)*perBox - sold), float64(start + boxes + perBox - sold), correct - 1, correct + 1, correct - 10, correct + 10}, optionCount, false),
	}
}

func genWordCooperativeFourVarSMP(optionCount int) question {
	balance := rand.IntN(900) + 100                     // 100-999 ribu
	dozens := rand.IntN(11) + 5                         // 5-15 lusin
	price := rand.IntN(41) + 20                         // 20-60 ribu per lusin
	income := rand.IntN(91) + 10                        // 10-100 ribu
	correct := float64(balance - dozens*price + income) // bisa negatif (utang)
	return question{
		prompt: fmt.Sprintf(
			"Koperasi sekolah punya saldo %d ribu rupiah. Koperasi membeli %d lusin buku tulis seharga %d ribu rupiah per lusin, lalu mendapat pemasukan %d ribu rupiah. Saldo koperasi sekarang (ribu rupiah, negatif = utang)?",
			balance, dozens, price, income,
		),
		options: buildOptions(correct, []float64{float64(balance - dozens*price - income), float64((balance-dozens)*price + income), -correct, correct - 10, correct + 10}, optionCount, true),
	}
}

func genSMPCampuranFourVariableMixed(optionCount int) question {
	// 50:50 soal langsung vs soal cerita.
	if rand.IntN(2) == 0 {
		switch rand.IntN(5) {
		case 0:
			return genAddMulSubFourVarSMP(optionCount)
		case 1:
			return genSubMulDivFourVarSMP(optionCount)
		case 2:
			return genParenSubMulAddFourVarSMP(optionCount)
		case 3:
			return genDivMulSubFourVarSMP(optionCount)
		default:
			return genMulSubMulFourVarSMP(optionCount)
		}
	}
	switch rand.IntN(3) {
	case 0:
		return genWordSubmarineFourVarSMP(optionCount)
	case 1:
		return genWordShopFourVarSMP(optionCount)
	default:
		return genWordCooperativeFourVarSMP(optionCount)
	}
}

func smpCampuranFourVariableChallenges() []challengeSpec {
	challenges := make([]challengeSpec, 0, 11)
	for level := 1; level <= 10; level++ {
		name := fmt.Sprintf("Campuran 4 Variabel Level %d", level)
		challenges = append(challenges, buildChallenge(name, false, 15, 23, 70, 5, 40, genSMPCampuranFourVariableMixed))
	}
	challenges = append(challenges, buildChallenge("Ujian Campuran 4 Variabel SMP", true, 25, 35, 70, 5, 1000, genSMPCampuranFourVariableMixed))
	return challenges
}

// ---------- SMP - Campuran 5 Variabel (batch 5, 5 angka + 4 operasi, salah satunya ratusan) ----------
//
// Naik lagi dari batch 4: 5 angka dioperasikan sekaligus (4 operasi), minimal
// satu angka di rentang ratusan (100-999). Gaya SMP: urutan operasi baku, ada
// kurung, hasil boleh negatif. 50:50 soal langsung vs soal cerita. Opsi tetep 5.
// Struktur: 10 level (bank 23, wajib 15) + 1 ujian (bank 35, wajib 25).

func genAddMulSubDivFiveVarSMP(optionCount int) question {
	a := rand.IntN(900) + 100       // 100-999
	b := rand.IntN(11) + 2          // 2-12
	c := rand.IntN(41) + 10         // 10-50
	e := rand.IntN(8) + 2           // 2-9
	k := rand.IntN(141) + 10        // 10-150
	d := e * k                      // d ÷ e = k (pas)
	correct := float64(a + b*c - k) // bisa negatif kalau k gede
	return question{
		prompt:  fmt.Sprintf("%d + %d × %d - %d ÷ %d = ?", a, b, c, d, e),
		options: buildOptions(correct, []float64{float64(a + b*c - d), float64((a+b)*c - k), -correct, correct - 1, correct + 1, correct - 10}, optionCount, true),
	}
}

func genParenSubMulSubMulFiveVarSMP(optionCount int) question {
	a := rand.IntN(900) + 100         // 100-999
	b := rand.IntN(90) + 10           // 10-99
	c := rand.IntN(8) + 2             // 2-9
	d := rand.IntN(90) + 10           // 10-99
	e := rand.IntN(11) + 2            // 2-12
	correct := float64((a-b)*c - d*e) // bisa negatif
	noParen := float64(a - b*c - d*e)
	return question{
		prompt:  fmt.Sprintf("(%d - %d) × %d - %d × %d = ?", a, b, c, d, e),
		options: buildOptions(correct, []float64{noParen, float64(((a-b)*c - d) * e), -correct, correct - float64(e), correct + float64(e), correct + 10}, optionCount, true),
	}
}

func genDivMulAddSubFiveVarSMP(optionCount int) question {
	b := rand.IntN(8) + 2 // 2-9
	q := rand.IntN(999/b-100/b) + 100/b + 1
	a := b * q                      // a ÷ b = q (pas), a di rentang ratusan
	c := rand.IntN(8) + 2           // 2-9
	d := rand.IntN(90) + 10         // 10-99
	e := rand.IntN(900) + 100       // 100-999
	correct := float64(q*c + d - e) // bisa negatif
	return question{
		prompt:  fmt.Sprintf("%d ÷ %d × %d + %d - %d = ?", a, b, c, d, e),
		options: buildOptions(correct, []float64{float64(q*c + d + e), float64(q*c - d - e), -correct, correct - 1, correct + 1, correct + 10}, optionCount, true),
	}
}

func genMulSubMulAddFiveVarSMP(optionCount int) question {
	a := rand.IntN(21) + 10           // 10-30
	b := rand.IntN(21) + 10           // 10-30
	c := rand.IntN(900) + 100         // 100-999
	d := rand.IntN(8) + 2             // 2-9
	e := rand.IntN(90) + 10           // 10-99
	correct := float64(a*b - c*d + e) // sering negatif
	wrongOrder := float64((a*b-c)*d + e)
	return question{
		prompt:  fmt.Sprintf("%d × %d - %d × %d + %d = ?", a, b, c, d, e),
		options: buildOptions(correct, []float64{wrongOrder, float64(a*b - c*d - e), -correct, correct - 1, correct + 1, correct - 10}, optionCount, true),
	}
}

func genSubParenAddMulDivFiveVarSMP(optionCount int) question {
	a := rand.IntN(900) + 100       // 100-999
	b := rand.IntN(41) + 10         // 10-50
	c := rand.IntN(41) + 10         // 10-50
	e := rand.IntN(8) + 2           // 2-9
	k := rand.IntN(4) + 2           // 2-5 (biar d != e)
	d := e * k                      // (b + c) × d ÷ e = (b + c) × k (pas)
	correct := float64(a - (b+c)*k) // bisa negatif
	noParen := float64(a - b - c*k)
	return question{
		prompt:  fmt.Sprintf("%d - (%d + %d) × %d ÷ %d = ?", a, b, c, d, e),
		options: buildOptions(correct, []float64{noParen, float64(a - (b+c)*d), -correct, correct - float64(k), correct + float64(k), correct + 10}, optionCount, true),
	}
}

func genWordSubmarineFiveVarSMP(optionCount int) question {
	depth := rand.IntN(900) + 100 // 100-999 m di bawah permukaan laut
	rate := rand.IntN(8) + 2      // 2-9 m per menit
	minutes := rand.IntN(26) + 5  // 5-30 menit
	dive := rand.IntN(91) + 10    // 10-100 m
	up := rand.IntN(91) + 10      // 10-100 m
	correct := float64(-depth + rate*minutes - dive + up)
	return question{
		prompt: fmt.Sprintf(
			"Sebuah kapal selam berada di kedalaman %d meter di bawah permukaan laut. Kapal itu naik %d meter tiap menit selama %d menit, lalu menyelam lagi %d meter, kemudian naik %d meter. Posisi kapal selam sekarang (meter, negatif = di bawah permukaan laut)?",
			depth, rate, minutes, dive, up,
		),
		options: buildOptions(correct, []float64{-correct, float64(-depth + rate*minutes + dive - up), float64(-depth - rate*minutes - dive + up), correct - float64(rate), correct + 10}, optionCount, true),
	}
}

func genWordShopFiveVarSMP(optionCount int) question {
	start := rand.IntN(900) + 100 // 100-999 stok awal
	boxes := rand.IntN(11) + 2    // 2-12 kardus
	perBox := rand.IntN(41) + 10  // 10-50 per kardus
	sold := rand.IntN(start+boxes*perBox-1) + 1
	broken := rand.IntN(start + boxes*perBox - sold + 1) // 0..sisa
	correct := float64(start + boxes*perBox - sold - broken)
	return question{
		prompt: fmt.Sprintf(
			"Sebuah toko punya stok %d botol minuman. Toko itu menerima %d kardus baru, tiap kardus isi %d botol. Lalu %d botol terjual dan %d botol pecah. Sisa stok botol minuman sekarang?",
			start, boxes, perBox, sold, broken,
		),
		options: buildOptions(correct, []float64{float64((start+boxes)*perBox - sold - broken), float64(start + boxes*perBox - sold + broken), correct - 1, correct + 1, correct - 10, correct + 10}, optionCount, false),
	}
}

func genWordCooperativeFiveVarSMP(optionCount int) question {
	balance := rand.IntN(900) + 100                           // 100-999 ribu
	dozens := rand.IntN(11) + 5                               // 5-15 lusin
	price := rand.IntN(41) + 20                               // 20-60 ribu per lusin
	income := rand.IntN(91) + 10                              // 10-100 ribu
	fee := rand.IntN(46) + 5                                  // 5-50 ribu
	correct := float64(balance - dozens*price + income - fee) // bisa negatif (utang)
	return question{
		prompt: fmt.Sprintf(
			"Koperasi sekolah punya saldo %d ribu rupiah. Koperasi membeli %d lusin buku tulis seharga %d ribu rupiah per lusin, mendapat pemasukan %d ribu rupiah, lalu membayar ongkos kirim %d ribu rupiah. Saldo koperasi sekarang (ribu rupiah, negatif = utang)?",
			balance, dozens, price, income, fee,
		),
		options: buildOptions(correct, []float64{float64(balance - dozens*price - income - fee), float64(balance - dozens*price + income + fee), -correct, correct - 10, correct + 10}, optionCount, true),
	}
}

func genSMPCampuranFiveVariableMixed(optionCount int) question {
	// 50:50 soal langsung vs soal cerita.
	if rand.IntN(2) == 0 {
		switch rand.IntN(5) {
		case 0:
			return genAddMulSubDivFiveVarSMP(optionCount)
		case 1:
			return genParenSubMulSubMulFiveVarSMP(optionCount)
		case 2:
			return genDivMulAddSubFiveVarSMP(optionCount)
		case 3:
			return genMulSubMulAddFiveVarSMP(optionCount)
		default:
			return genSubParenAddMulDivFiveVarSMP(optionCount)
		}
	}
	switch rand.IntN(3) {
	case 0:
		return genWordSubmarineFiveVarSMP(optionCount)
	case 1:
		return genWordShopFiveVarSMP(optionCount)
	default:
		return genWordCooperativeFiveVarSMP(optionCount)
	}
}

func smpCampuranFiveVariableChallenges() []challengeSpec {
	challenges := make([]challengeSpec, 0, 11)
	for level := 1; level <= 10; level++ {
		name := fmt.Sprintf("Campuran 5 Variabel Level %d", level)
		challenges = append(challenges, buildChallenge(name, false, 15, 23, 70, 5, 45, genSMPCampuranFiveVariableMixed))
	}
	challenges = append(challenges, buildChallenge("Ujian Campuran 5 Variabel SMP", true, 25, 35, 70, 5, 1125, genSMPCampuranFiveVariableMixed))
	return challenges
}

// ---------- SMK / SMA ----------

func genLinearTwoStep(optionCount int) question {
	x := rand.IntN(11) - 5
	a := rand.IntN(5) + 2
	b := rand.IntN(20) - 10
	c := a*x + b
	correct := float64(x)
	return question{
		prompt:  fmt.Sprintf("%dx + %d = %d, x = ?", a, b, c),
		options: buildOptions(correct, []float64{correct - 1, correct + 1, correct - 2, correct + 2, correct + 3}, optionCount, true),
	}
}

func absInt(n int) int {
	if n < 0 {
		return -n
	}
	return n
}

func genQuadraticRoot(optionCount int) question {
	r1 := rand.IntN(7) - 3
	r2 := rand.IntN(7) - 3
	for r2 == r1 {
		r2 = rand.IntN(7) - 3
	}
	sum := -(r1 + r2)
	prod := r1 * r2
	correct := float64(r1)
	sign := "+"
	if sum < 0 {
		sign = "-"
	}
	prompt := fmt.Sprintf("Salah satu akar dari x² %s %dx + %d = 0 adalah?", sign, absInt(sum), prod)
	return question{
		prompt:  prompt,
		options: buildOptions(correct, []float64{float64(r2), correct + 1, correct - 1, correct + 2}, optionCount, true),
	}
}

type trigFact struct {
	label string
	value float64
}

var trigTable = []trigFact{
	{"sin(0°)", 0}, {"cos(0°)", 1}, {"sin(30°)", 0.5}, {"cos(60°)", 0.5},
	{"sin(90°)", 1}, {"cos(90°)", 0}, {"tan(45°)", 1}, {"tan(0°)", 0},
	{"sin(45°)", 0.71}, {"cos(45°)", 0.71}, {"cos(30°)", 0.87}, {"sin(60°)", 0.87},
}

func genTrig(optionCount int) question {
	fact := trigTable[rand.IntN(len(trigTable))]
	others := make([]float64, 0, len(trigTable))
	for _, f := range trigTable {
		if f.value != fact.value {
			others = append(others, f.value)
		}
	}
	return question{
		prompt:  fmt.Sprintf("%s = ?", fact.label),
		options: buildOptions(fact.value, others, optionCount, false),
	}
}

func genLogarithm(optionCount int) question {
	bases := []int{2, 3, 5, 10}
	base := bases[rand.IntN(len(bases))]
	k := rand.IntN(4) + 1
	value := 1
	for range k {
		value *= base
	}
	correct := float64(k)
	return question{
		prompt:  fmt.Sprintf("Log basis %d dari %d = ?", base, value),
		options: buildOptions(correct, []float64{correct - 1, correct + 1, correct + 2, correct - 2}, optionCount, false),
	}
}

func genStatisticsMean(optionCount int) question {
	const n = 5
	mean := rand.IntN(15) + 10
	values := make([]int, n)
	sum := 0
	for i := range n - 1 {
		v := mean + rand.IntN(7) - 3
		values[i] = v
		sum += v
	}
	values[n-1] = mean*n - sum
	strs := make([]string, n)
	for i, v := range values {
		strs[i] = fmt.Sprintf("%d", v)
	}
	correct := float64(mean)
	return question{
		prompt:  fmt.Sprintf("Rata-rata dari bilangan %s = ?", strings.Join(strs, ", ")),
		options: buildOptions(correct, []float64{correct - 1, correct + 1, correct - 2, correct + 2}, optionCount, true),
	}
}

func genSMKMixed(optionCount int) question {
	switch rand.IntN(5) {
	case 0:
		return genLinearTwoStep(optionCount)
	case 1:
		return genQuadraticRoot(optionCount)
	case 2:
		return genTrig(optionCount)
	case 3:
		return genLogarithm(optionCount)
	default:
		return genStatisticsMean(optionCount)
	}
}

func smkTierSpec() tierSpec {
	return tierSpec{
		code: "smk", name: "SMK / SMA",
		batches: []batchSpec{
			{
				name: "Batch 1 — Semester 1",
				challenges: []challengeSpec{
					buildChallenge("Aljabar Lanjut", false, 5, 8, 70, 5, 30, genLinearTwoStep),
					buildChallenge("Persamaan Kuadrat", false, 5, 8, 70, 5, 30, genQuadraticRoot),
					buildChallenge("Trigonometri Dasar", false, 5, 8, 70, 5, 30, genTrig),
					buildChallenge("Logaritma Dasar", false, 5, 8, 70, 5, 30, genLogarithm),
					buildChallenge("Statistika Dasar", false, 5, 8, 70, 5, 30, genStatisticsMean),
					buildChallenge("Ujian Semester 1 SMK/SMA", true, 8, 12, 70, 5, 600, genSMKMixed),
				},
			},
			{
				name:       "Batch 2 — Materi Kelas X–XI",
				challenges: smkMateriChallenges(smkMateriKelasXXI(), "Ujian Materi Kelas X–XI"),
			},
			{
				name:       "Batch 3 — Materi Kelas XII & SMK",
				challenges: smkMateriChallenges(smkMateriKelasXIIDanSMK(), "Ujian Materi Kelas XII & SMK"),
			},
			{
				name:       "Batch 4 — Campuran Semua Materi",
				challenges: smkCampuranSemuaMateriChallenges(),
			},
			{
				name:       "Batch 5 — Campuran Sulit",
				challenges: smkCampuranSulitChallenges(),
			},
		},
	}
}

// ---------- SMK/SMA - generator soal lanjutan ----------
//
// Dulu dipakai batch "Soal Campuran" (udah dihapus, diganti batch per-materi di
// smk_materi.go). Generatornya tetep dipake ulang di level materi yang cocok:
// SPLDV, jumlah/hasil kali akar kuadrat, kombinasi trig, kombinasi log, median.

func genLinearSystemHard(optionCount int) question {
	x := rand.IntN(11) - 5
	y := rand.IntN(11) - 5
	a, b := rand.IntN(4)+1, rand.IntN(4)+1
	d, e := rand.IntN(4)+1, rand.IntN(4)+1
	for d*b == e*a {
		d = rand.IntN(4) + 1
	}
	c := a*x + b*y
	f := d*x + e*y
	correct := float64(x)
	return question{
		prompt:  fmt.Sprintf("%s = %d dan %s = %d, nilai x = ?", formatLinearXY(a, b), c, formatLinearXY(d, e), f),
		options: buildOptions(correct, []float64{correct - 1, correct + 1, correct - 2, correct + 2, correct + 3}, optionCount, true),
	}
}

func genQuadraticSumProduct(optionCount int) question {
	r1 := rand.IntN(9) - 4
	r2 := rand.IntN(9) - 4
	for r2 == r1 {
		r2 = rand.IntN(9) - 4
	}
	a := rand.IntN(3) + 1
	b := -a * (r1 + r2)
	c := a * r1 * r2
	prompt := fmt.Sprintf("%s = 0 punya akar x1 dan x2.", formatPoly(a, b, c))

	var correct float64
	if rand.IntN(2) == 0 {
		correct = float64(r1 + r2)
		prompt += " Nilai x1 + x2 = ?"
	} else {
		correct = float64(r1 * r2)
		prompt += " Nilai x1 × x2 = ?"
	}
	return question{
		prompt:  prompt,
		options: buildOptions(correct, []float64{correct - 1, correct + 1, correct - 2, correct + 2, -correct}, optionCount, true),
	}
}

func genTrigCombo(optionCount int) question {
	f1 := trigTable[rand.IntN(len(trigTable))]
	f2 := trigTable[rand.IntN(len(trigTable))]
	for f2.label == f1.label {
		f2 = trigTable[rand.IntN(len(trigTable))]
	}
	op := "+"
	correct := f1.value + f2.value
	if rand.IntN(2) == 0 {
		op = "-"
		correct = f1.value - f2.value
	}
	return question{
		prompt:  fmt.Sprintf("%s %s %s = ?", f1.label, op, f2.label),
		options: buildOptions(correct, []float64{correct - 0.5, correct + 0.5, correct - 0.29, correct + 0.29, -correct}, optionCount, true),
	}
}

func genLogCombo(optionCount int) question {
	bases := []int{2, 3, 5}
	b1 := bases[rand.IntN(len(bases))]
	b2 := bases[rand.IntN(len(bases))]
	k1 := rand.IntN(4) + 1
	k2 := rand.IntN(4) + 1
	v1, v2 := 1, 1
	for range k1 {
		v1 *= b1
	}
	for range k2 {
		v2 *= b2
	}
	correct := float64(k1 + k2)
	return question{
		prompt:  fmt.Sprintf("Log basis %d dari %d + Log basis %d dari %d = ?", b1, v1, b2, v2),
		options: buildOptions(correct, []float64{correct - 1, correct + 1, correct + 2, correct - 2}, optionCount, false),
	}
}

func genStatsMedianHard(optionCount int) question {
	const n = 5
	values := make([]int, n)
	for i := range values {
		values[i] = rand.IntN(41) - 10
	}
	sorted := append([]int{}, values...)
	sort.Ints(sorted)
	median := sorted[n/2]
	strs := make([]string, n)
	for i, v := range values {
		strs[i] = fmt.Sprintf("%d", v)
	}
	correct := float64(median)
	return question{
		prompt:  fmt.Sprintf("Median dari data %s adalah?", strings.Join(strs, ", ")),
		options: buildOptions(correct, []float64{correct - 1, correct + 1, correct - 2, correct + 2, float64(sorted[0])}, optionCount, true),
	}
}
