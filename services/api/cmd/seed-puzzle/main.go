package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/rand/v2"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"lesson/api/internal/config"
)

// Seed tier "umum": tier tanpa batch (uses_batch = false), challenge-nya
// nempel ke category, bukan batch. Dua kategori awal: Addition Puzzle Boxes
// (soal grid_puzzle kind "addition_grid", level = ukuran kotak NxN) &
// Cryptarithm (grid_puzzle kind "cryptarithm", level = jumlah huruf unik).
// Idempotent per-challenge sama kayak seed-curriculum: kalau challenge udah
// punya soal, dilewatin.

type additionGridPayload struct {
	Size         int      `json:"size"`
	SolutionGrid [][]int  `json:"solutionGrid"`
	GivenMask    [][]bool `json:"givenMask"`
}

type cryptarithmPayload struct {
	Words    []string       `json:"words"`
	Solution map[string]int `json:"solution"`
}

type puzzleChallengeSpec struct {
	name                string
	timeLimitSeconds    int
	additionGridSize    int // > 0 kalau kind == addition_grid
	cryptarithmPuzzles  []cryptarithmSpec
	additionPuzzleCount int
}

type cryptarithmSpec struct {
	words  []string
	result string
}

func main() {
	cfg := config.Load()
	ctx := context.Background()

	pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("connect db: %v", err)
	}
	defer pool.Close()

	tierID, err := getOrCreateTier(ctx, pool, "umum", "Umum", false, 3)
	if err != nil {
		log.Fatalf("tier umum: %v", err)
	}

	// --- Kategori 1: Addition Puzzle Boxes ---
	additionCatID, err := getOrCreateCategory(ctx, pool, tierID, "addition-puzzle-boxes", "Addition Puzzle Boxes", 0)
	if err != nil {
		log.Fatalf("category addition-puzzle-boxes: %v", err)
	}
	additionLevels := []puzzleChallengeSpec{
		{name: "Level 1 — Kotak 3x3", timeLimitSeconds: 120, additionGridSize: 3, additionPuzzleCount: 3},
		{name: "Level 2 — Kotak 4x4", timeLimitSeconds: 180, additionGridSize: 4, additionPuzzleCount: 3},
		{name: "Level 3 — Kotak 5x5", timeLimitSeconds: 300, additionGridSize: 5, additionPuzzleCount: 3},
	}
	fmt.Println("Addition Puzzle Boxes:")
	for order, level := range additionLevels {
		if err := seedAdditionLevel(ctx, pool, additionCatID, level, order); err != nil {
			log.Fatalf("addition level %s: %v", level.name, err)
		}
	}

	// --- Kategori 2: Cryptarithm ---
	cryptoCatID, err := getOrCreateCategory(ctx, pool, tierID, "cryptarithm", "Cryptarithm", 1)
	if err != nil {
		log.Fatalf("category cryptarithm: %v", err)
	}
	cryptoLevels := []puzzleChallengeSpec{
		{
			name:             "Level 1 — Pemula",
			timeLimitSeconds: 180,
			cryptarithmPuzzles: []cryptarithmSpec{
				{words: []string{"BIG", "BAG"}, result: "LOAD"},
				{words: []string{"FUN", "RUN"}, result: "RACE"},
				{words: []string{"BASE", "BALL"}, result: "GAMES"},
			},
		},
		{
			name:             "Level 2 — Menengah",
			timeLimitSeconds: 240,
			cryptarithmPuzzles: []cryptarithmSpec{
				{words: []string{"SEND", "MORE"}, result: "MONEY"},
				{words: []string{"CAT", "DOG"}, result: "PETS"},
				{words: []string{"CROSS", "ROADS"}, result: "DANGER"},
			},
		},
	}
	fmt.Println("Cryptarithm:")
	for order, level := range cryptoLevels {
		if err := seedCryptarithmLevel(ctx, pool, cryptoCatID, level, order); err != nil {
			log.Fatalf("cryptarithm level %s: %v", level.name, err)
		}
	}

	fmt.Println("\nSeed puzzle (tier umum) selesai.")
}

func seedAdditionLevel(ctx context.Context, pool *pgxpool.Pool, categoryID string, spec puzzleChallengeSpec, orderIndex int) error {
	challengeID, existed, err := getOrCreateChallengeByCategory(ctx, pool, categoryID, spec.name, spec.timeLimitSeconds, orderIndex)
	if err != nil {
		return err
	}
	if existed {
		fmt.Printf("  - %s (sudah ada, dilewati)\n", spec.name)
		return nil
	}

	for i := 0; i < spec.additionPuzzleCount; i++ {
		solution, givenMask := generateAdditionGrid(spec.additionGridSize)
		payload := additionGridPayload{Size: spec.additionGridSize, SolutionGrid: solution, GivenMask: givenMask}
		payloadJSON, err := json.Marshal(payload)
		if err != nil {
			return err
		}

		prompt := fmt.Sprintf(
			"Isi kotak kosong dengan angka 1–%d (masing-masing dipakai tepat sekali) supaya tiap baris & kolom sesuai jumlah yang tertera.",
			spec.additionGridSize*spec.additionGridSize,
		)

		var questionID string
		if err := pool.QueryRow(ctx, `
			INSERT INTO questions (challenge_id, question_type, prompt, status)
			VALUES ($1, 'grid_puzzle', $2, 'published')
			RETURNING id
		`, challengeID, prompt).Scan(&questionID); err != nil {
			return err
		}
		if _, err := pool.Exec(ctx, `
			INSERT INTO puzzle_questions (question_id, kind, grid_size, payload)
			VALUES ($1, 'addition_grid', $2, $3)
		`, questionID, spec.additionGridSize, payloadJSON); err != nil {
			return err
		}
	}
	fmt.Printf("  - %s: %d puzzle di bank\n", spec.name, spec.additionPuzzleCount)
	return nil
}

func seedCryptarithmLevel(ctx context.Context, pool *pgxpool.Pool, categoryID string, spec puzzleChallengeSpec, orderIndex int) error {
	challengeID, existed, err := getOrCreateChallengeByCategory(ctx, pool, categoryID, spec.name, spec.timeLimitSeconds, orderIndex)
	if err != nil {
		return err
	}
	if existed {
		fmt.Printf("  - %s (sudah ada, dilewati)\n", spec.name)
		return nil
	}

	for _, puzzle := range spec.cryptarithmPuzzles {
		solution, err := solveCryptarithm(puzzle.words, puzzle.result)
		if err != nil {
			return fmt.Errorf("%s+%s: %w", puzzle.words, puzzle.result, err)
		}

		words := append([]string{}, puzzle.words...)
		words = append(words, puzzle.result)
		payload := cryptarithmPayload{Words: words, Solution: solution}
		payloadJSON, err := json.Marshal(payload)
		if err != nil {
			return err
		}

		prompt := fmt.Sprintf(
			"Setiap huruf mewakili satu angka (0-9), gak boleh ada dua huruf beda yang sama angkanya. Cari nilai tiap huruf supaya %s = %s benar.",
			joinPlus(puzzle.words), puzzle.result,
		)

		var questionID string
		if err := pool.QueryRow(ctx, `
			INSERT INTO questions (challenge_id, question_type, prompt, status)
			VALUES ($1, 'grid_puzzle', $2, 'published')
			RETURNING id
		`, challengeID, prompt).Scan(&questionID); err != nil {
			return err
		}
		if _, err := pool.Exec(ctx, `
			INSERT INTO puzzle_questions (question_id, kind, grid_size, payload)
			VALUES ($1, 'cryptarithm', NULL, $2)
		`, questionID, payloadJSON); err != nil {
			return err
		}
	}
	fmt.Printf("  - %s: %d puzzle di bank\n", spec.name, len(spec.cryptarithmPuzzles))
	return nil
}

func joinPlus(words []string) string {
	out := ""
	for i, w := range words {
		if i > 0 {
			out += " + "
		}
		out += w
	}
	return out
}

// generateAdditionGrid: acak permutasi 1..size*size ke kotak NxN, lalu pilih
// max(2,size) sel acak buat jadi "given" (kesisanya kosong buat diisi user).
// rowSums/colSums gak disimpen di sini -- dihitung on-the-fly dari
// SolutionGrid di curriculumsvc.derivePuzzle (server-side, satu sumber kebenaran).
func generateAdditionGrid(size int) (solution [][]int, givenMask [][]bool) {
	values := make([]int, size*size)
	for i := range values {
		values[i] = i + 1
	}
	rand.Shuffle(len(values), func(i, j int) { values[i], values[j] = values[j], values[i] })

	solution = make([][]int, size)
	for i := 0; i < size; i++ {
		solution[i] = append([]int{}, values[i*size:(i+1)*size]...)
	}

	givenMask = make([][]bool, size)
	for i := range givenMask {
		givenMask[i] = make([]bool, size)
	}
	givenCount := size
	if givenCount < 2 {
		givenCount = 2
	}
	cells := rand.Perm(size * size)
	for i := 0; i < givenCount; i++ {
		r, c := cells[i]/size, cells[i]%size
		givenMask[r][c] = true
	}
	return solution, givenMask
}

// solveCryptarithm: brute-force nyari 1 solusi valid (huruf -> digit unik
// 0-9, huruf awal kata gak boleh 0) buat words[0]+words[1]+...=result. Dipake
// buat MEMVALIDASI puzzle yang di-seed beneran solvable & benar secara
// aritmatika sebelum disimpen -- bukan cuma percaya hasil hitungan manual.
func solveCryptarithm(words []string, result string) (map[string]int, error) {
	all := append(append([]string{}, words...), result)
	letterSet := map[rune]bool{}
	for _, w := range all {
		for _, ch := range w {
			letterSet[ch] = true
		}
	}
	letters := make([]rune, 0, len(letterSet))
	for ch := range letterSet {
		letters = append(letters, ch)
	}
	if len(letters) > 10 {
		return nil, errors.New("lebih dari 10 huruf unik, gak mungkin dipetakan ke digit 0-9")
	}

	leading := map[rune]bool{}
	for _, w := range all {
		if len(w) > 1 {
			leading[rune(w[0])] = true
		}
	}

	digits := []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
	assignment := map[rune]int{}
	used := make([]bool, 10)

	var wordValue func(w string, m map[rune]int) int
	wordValue = func(w string, m map[rune]int) int {
		n := 0
		for _, ch := range w {
			n = n*10 + m[ch]
		}
		return n
	}

	var backtrack func(idx int) bool
	backtrack = func(idx int) bool {
		if idx == len(letters) {
			sum := 0
			for _, w := range words {
				sum += wordValue(w, assignment)
			}
			return sum == wordValue(result, assignment)
		}
		ch := letters[idx]
		for _, d := range digits {
			if used[d] {
				continue
			}
			if d == 0 && leading[ch] {
				continue
			}
			used[d] = true
			assignment[ch] = d
			if backtrack(idx + 1) {
				return true
			}
			used[d] = false
			delete(assignment, ch)
		}
		return false
	}

	if !backtrack(0) {
		return nil, errors.New("gak solvable")
	}

	out := make(map[string]int, len(assignment))
	for ch, d := range assignment {
		out[string(ch)] = d
	}
	return out, nil
}

func getOrCreateTier(ctx context.Context, pool *pgxpool.Pool, code, name string, usesBatch bool, orderIndex int) (string, error) {
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
		VALUES ($1, $2, $3, $4)
		RETURNING id
	`, code, name, usesBatch, orderIndex).Scan(&id)
	return id, err
}

func getOrCreateCategory(ctx context.Context, pool *pgxpool.Pool, tierID, code, name string, orderIndex int) (string, error) {
	var id string
	err := pool.QueryRow(ctx, `SELECT id FROM categories WHERE tier_id = $1 AND code = $2`, tierID, code).Scan(&id)
	if err == nil {
		return id, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return "", err
	}
	err = pool.QueryRow(ctx, `
		INSERT INTO categories (tier_id, code, name, order_index)
		VALUES ($1, $2, $3, $4)
		RETURNING id
	`, tierID, code, name, orderIndex).Scan(&id)
	return id, err
}

func getOrCreateChallengeByCategory(ctx context.Context, pool *pgxpool.Pool, categoryID, name string, timeLimitSeconds, orderIndex int) (id string, existed bool, err error) {
	err = pool.QueryRow(ctx, `
		SELECT id FROM challenges WHERE category_id = $1 AND name = $2
	`, categoryID, name).Scan(&id)
	if err == nil {
		return id, true, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return "", false, err
	}
	err = pool.QueryRow(ctx, `
		INSERT INTO challenges (
			category_id, name, order_index, is_exam,
			question_count_required, pass_threshold_percent, time_limit_seconds
		)
		VALUES ($1, $2, $3, false, 1, 100, $4)
		RETURNING id
	`, categoryID, name, orderIndex, timeLimitSeconds).Scan(&id)
	return id, false, err
}
