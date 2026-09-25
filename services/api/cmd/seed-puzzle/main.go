package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"lesson/api/internal/config"
	"lesson/api/internal/curriculumsvc"
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

	tierID, err := getOrCreateTier(ctx, pool, "umum", "Umum", false, 4)
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
		solution, givenMask := curriculumsvc.GenerateAdditionGrid(spec.additionGridSize)
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
		solution, err := curriculumsvc.SolveCryptarithm(puzzle.words, puzzle.result)
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
