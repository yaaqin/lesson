package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math/rand/v2"
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

type tierSpec struct {
	code       string
	name       string
	batchName  string
	challenges []challengeSpec
}

func main() {
	cfg := config.Load()
	ctx := context.Background()

	pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("connect db: %v", err)
	}
	defer pool.Close()

	tiers := []tierSpec{sdTierSpec(), smpTierSpec(), smkTierSpec()}

	for tierOrder, tier := range tiers {
		tierID, err := getOrCreateTier(ctx, pool, tier.code, tier.name, tierOrder)
		if err != nil {
			log.Fatalf("tier %s: %v", tier.code, err)
		}

		batchID, err := getOrCreateBatch(ctx, pool, tierID, tier.batchName, 0)
		if err != nil {
			log.Fatalf("batch %s: %v", tier.batchName, err)
		}

		fmt.Printf("%s (%s):\n", tier.name, tier.batchName)
		for order, spec := range tier.challenges {
			challengeID, existed, err := getOrCreateChallenge(ctx, pool, batchID, spec, order)
			if err != nil {
				log.Fatalf("challenge %s: %v", spec.name, err)
			}
			if existed {
				fmt.Printf("  - %s (sudah ada, dilewati)\n", spec.name)
				continue
			}
			if err := insertQuestions(ctx, pool, challengeID, spec.questions); err != nil {
				log.Fatalf("insert questions for %s: %v", spec.name, err)
			}
			kind := "challenge"
			if spec.isExam {
				kind = "ujian"
			}
			fmt.Printf("  - %s [%s]: %d soal di bank\n", spec.name, kind, len(spec.questions))
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
	options = append(options, option{value: correct, isCorrect: true})
	for _, d := range distinct {
		options = append(options, option{value: d, isCorrect: false})
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
		code: "sd", name: "SD", batchName: "Batch 1 — Semester 1",
		challenges: []challengeSpec{
			buildChallenge("Penjumlahan", false, 5, 8, 70, 3, 20, genAddition),
			buildChallenge("Pengurangan", false, 5, 8, 70, 3, 20, genSubtraction),
			buildChallenge("Perkalian", false, 5, 8, 70, 3, 20, genMultiplication),
			buildChallenge("Pembagian", false, 5, 8, 70, 3, 20, genDivision),
			buildChallenge("Campuran Dasar", false, 5, 8, 70, 3, 20, genSDMixed),
			buildChallenge("Ujian Semester 1 SD", true, 8, 12, 70, 3, 300, genSDMixed),
		},
	}
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
		code: "smp", name: "SMP", batchName: "Batch 1 — Semester 1",
		challenges: []challengeSpec{
			buildChallenge("Aljabar Dasar", false, 5, 8, 70, 4, 25, genLinearEquation),
			buildChallenge("Pecahan", false, 5, 8, 70, 4, 25, genFractionSimplify),
			buildChallenge("Persentase", false, 5, 8, 70, 4, 25, genPercentage),
			buildChallenge("Perbandingan", false, 5, 8, 70, 4, 25, genRatio),
			buildChallenge("Bilangan Bulat", false, 5, 8, 70, 4, 25, genIntegerOps),
			buildChallenge("Ujian Semester 1 SMP", true, 8, 12, 70, 4, 480, genSMPMixed),
		},
	}
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
		code: "smk", name: "SMK / SMA", batchName: "Batch 1 — Semester 1",
		challenges: []challengeSpec{
			buildChallenge("Aljabar Lanjut", false, 5, 8, 70, 5, 30, genLinearTwoStep),
			buildChallenge("Persamaan Kuadrat", false, 5, 8, 70, 5, 30, genQuadraticRoot),
			buildChallenge("Trigonometri Dasar", false, 5, 8, 70, 5, 30, genTrig),
			buildChallenge("Logaritma Dasar", false, 5, 8, 70, 5, 30, genLogarithm),
			buildChallenge("Statistika Dasar", false, 5, 8, 70, 5, 30, genStatisticsMean),
			buildChallenge("Ujian Semester 1 SMK/SMA", true, 8, 12, 70, 5, 600, genSMKMixed),
		},
	}
}
