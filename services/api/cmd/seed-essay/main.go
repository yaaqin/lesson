package main

import (
	"context"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"

	"lesson/api/internal/config"
)

// Seed contoh soal essay_numeric (isian angka, bukan pilihan ganda) buat SMP
// & SMK -- nambahin ke bank soal challenge yang UDAH ADA (bukan bikin
// challenge baru). Idempotent per challenge: kalau challenge itu udah punya
// soal essay, dilewatin (gak dobel kalau rerun).
type essayQuestion struct {
	prompt        string
	correctAnswer float64
}

type target struct {
	tierCode      string
	challengeName string
	questions     []essayQuestion
}

var targets = []target{
	{
		tierCode:      "smp",
		challengeName: "Aljabar Dasar",
		questions: []essayQuestion{
			{"Jika 2x = 18, berapa nilai x?", 9},
			{"Sebuah kelas punya 8 baris kursi, masing-masing baris berisi 4 kursi. Berapa total kursi di kelas itu?", 32},
			{"Umur Andi 3 tahun lebih tua dari Budi. Kalau umur Budi 12 tahun, berapa umur Andi?", 15},
			{"Berapa hasil dari 144 ÷ 12?", 12},
		},
	},
	{
		tierCode:      "smk",
		challengeName: "Aljabar Lanjut",
		questions: []essayQuestion{
			{"Jika 3x - 5 = 16, berapa nilai x?", 7},
			{"Keliling sebuah persegi adalah 48 cm. Berapa panjang sisinya (cm)?", 12},
			{"Sebuah mobil menempuh 240 km dalam 4 jam. Berapa kecepatan rata-ratanya (km/jam)?", 60},
			{"Jika log basis 2 dari x sama dengan 5, berapa nilai x?", 32},
		},
	},
	{
		tierCode:      "smp",
		challengeName: "Ujian Semester 1 SMP",
		questions: []essayQuestion{
			{"Jika 4x = 28, berapa nilai x?", 7},
			{"Sebuah toko menjual 6 lusin pensil. Berapa jumlah pensil semuanya?", 72},
			{"Berapa hasil dari 15 + 27?", 42},
		},
	},
	{
		tierCode:      "smk",
		challengeName: "Ujian Semester 1 SMK/SMA",
		questions: []essayQuestion{
			{"Jika 5x + 10 = 60, berapa nilai x?", 10},
			{"Sebuah persegi panjang punya panjang 15 cm dan lebar 8 cm. Berapa luasnya (cm²)?", 120},
			{"Jika log basis 10 dari x sama dengan 3, berapa nilai x?", 1000},
		},
	},
}

func main() {
	cfg := config.Load()
	ctx := context.Background()

	pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("connect db: %v", err)
	}
	defer pool.Close()

	for _, t := range targets {
		var challengeID string
		err := pool.QueryRow(ctx, `
			SELECT c.id
			FROM challenges c
			JOIN batches b ON b.id = c.batch_id
			JOIN tiers t ON t.id = b.tier_id
			WHERE t.code = $1 AND c.name = $2
		`, t.tierCode, t.challengeName).Scan(&challengeID)
		if err != nil {
			log.Fatalf("challenge %s/%s not found: %v", t.tierCode, t.challengeName, err)
		}

		var alreadyHasEssay bool
		if err := pool.QueryRow(ctx, `
			SELECT EXISTS(
				SELECT 1 FROM questions WHERE challenge_id = $1 AND question_type = 'essay_numeric'
			)
		`, challengeID).Scan(&alreadyHasEssay); err != nil {
			log.Fatalf("check existing essay for %s: %v", t.challengeName, err)
		}
		if alreadyHasEssay {
			fmt.Printf("%s (%s): sudah ada soal essay, dilewati\n", t.challengeName, t.tierCode)
			continue
		}

		for _, q := range t.questions {
			if _, err := pool.Exec(ctx, `
				INSERT INTO questions (challenge_id, question_type, prompt, correct_answer_value, status)
				VALUES ($1, 'essay_numeric', $2, $3, 'published')
			`, challengeID, q.prompt, q.correctAnswer); err != nil {
				log.Fatalf("insert essay question for %s: %v", t.challengeName, err)
			}
		}
		fmt.Printf("%s (%s): %d soal essay ditambahkan\n", t.challengeName, t.tierCode, len(t.questions))
	}

	fmt.Println("\nSeed soal essay selesai.")
}
