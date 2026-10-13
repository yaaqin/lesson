package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"lesson/api/internal/config"
	"lesson/api/internal/coursesvc"
)

// Seed Course Coding Dasar (migrations/0016). Beda dari seed-curriculum, di
// sini semua konten ditulis manual & selalu ditimpa tiap dijalanin (module &
// level di-upsert by code, soal di module itu dihapus lalu diinsert ulang) --
// aman karena attempt nyimpen snapshot soalnya sendiri, dan konten course
// gak diedit dari dashboard.
//
// Module tanpa level (Looping, Asynchronous) tampil "Segera hadir" di
// apps/web & otomatis kekunci sampai module sebelumnya lulus.

type moduleSpec struct {
	code, title, description string
	examQuestionCount        int
	examTimeLimitSeconds     int
	examPassPercent          int
	levels                   []levelSpec
	examOnly                 []questionSpec // soal khusus final test
}

type levelSpec struct {
	code, title, summary string
	questionCount        int // soal per latihan, diambil acak dari bank
	material             []coursesvc.MaterialBlock
	questions            []questionSpec
}

// payload: sama bentuknya dengan coursesvc.payload (kolom JSONB).
type payload struct {
	Options     []string          `json:"options,omitempty"`
	Correct     int               `json:"correct"`
	Accepted    []string          `json:"accepted,omitempty"`
	Schema      *coursesvc.Schema `json:"schema,omitempty"`
	StrictJSON  bool              `json:"strictJson,omitempty"`
	Placeholder string            `json:"placeholder,omitempty"`
	Sample      string            `json:"sample,omitempty"`
}

type questionSpec struct {
	kind        string
	prompt      string
	code        string
	explanation string
	payload     payload
}

func main() {
	cfg := config.Load()
	ctx := context.Background()

	pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("connect db: %v", err)
	}
	defer pool.Close()

	for order, m := range modules() {
		if err := seedModule(ctx, pool, order, m); err != nil {
			log.Fatalf("module %s: %v", m.code, err)
		}
	}
	fmt.Println("\nSeed course selesai.")
}

func seedModule(ctx context.Context, pool *pgxpool.Pool, order int, m moduleSpec) error {
	tx, err := pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	var moduleID string
	err = tx.QueryRow(ctx, `
		INSERT INTO course_modules (code, title, description, order_index,
			exam_question_count, exam_time_limit_seconds, exam_pass_percent)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (code) DO UPDATE SET
			title = EXCLUDED.title, description = EXCLUDED.description,
			order_index = EXCLUDED.order_index,
			exam_question_count = EXCLUDED.exam_question_count,
			exam_time_limit_seconds = EXCLUDED.exam_time_limit_seconds,
			exam_pass_percent = EXCLUDED.exam_pass_percent,
			updated_at = now()
		RETURNING id
	`, m.code, m.title, m.description, order, m.examQuestionCount, m.examTimeLimitSeconds, m.examPassPercent).Scan(&moduleID)
	if err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `DELETE FROM course_questions WHERE module_id = $1`, moduleID); err != nil {
		return err
	}

	fmt.Printf("%s:\n", m.title)
	total := 0
	for i, l := range m.levels {
		material, err := json.Marshal(l.material)
		if err != nil {
			return err
		}
		var levelID string
		err = tx.QueryRow(ctx, `
			INSERT INTO course_levels (module_id, code, title, summary, order_index, material, question_count)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
			ON CONFLICT (module_id, code) DO UPDATE SET
				title = EXCLUDED.title, summary = EXCLUDED.summary,
				order_index = EXCLUDED.order_index, material = EXCLUDED.material,
				question_count = EXCLUDED.question_count, updated_at = now()
			RETURNING id
		`, moduleID, l.code, l.title, l.summary, i, material, l.questionCount).Scan(&levelID)
		if err != nil {
			return err
		}
		if err := insertQuestions(ctx, tx, moduleID, &levelID, l.questions); err != nil {
			return err
		}
		total += len(l.questions)
		fmt.Printf("  Level %d — %s: %d blok materi, %d soal di bank (%d per latihan)\n",
			i+1, l.title, len(l.material), len(l.questions), l.questionCount)
	}
	if err := insertQuestions(ctx, tx, moduleID, nil, m.examOnly); err != nil {
		return err
	}
	total += len(m.examOnly)
	if len(m.levels) == 0 {
		fmt.Println("  (segera hadir)")
	} else {
		fmt.Printf("  Final test: %d soal dari bank %d, %ds, lulus %d%%\n",
			m.examQuestionCount, total, m.examTimeLimitSeconds, m.examPassPercent)
	}
	return tx.Commit(ctx)
}

func insertQuestions(ctx context.Context, tx pgx.Tx, moduleID string, levelID *string, qs []questionSpec) error {
	for i, q := range qs {
		raw, err := json.Marshal(q.payload)
		if err != nil {
			return err
		}
		var code *string
		if q.code != "" {
			code = &q.code
		}
		if _, err := tx.Exec(ctx, `
			INSERT INTO course_questions (module_id, level_id, kind, prompt, code, payload, explanation, order_index)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		`, moduleID, levelID, q.kind, q.prompt, code, raw, q.explanation, i); err != nil {
			return err
		}
	}
	return nil
}

// ---------- helper bikin soal ----------

const askOutput = "Apa output kode berikut?"
const askType = "Nilai ini tipe datanya apa?"

// mc: correct ditaruh di indeks 0, server yang ngacak urutan opsi tiap mulai.
func mc(prompt, code, correct string, wrong []string, explanation string) questionSpec {
	return questionSpec{
		kind: coursesvc.KindMultipleChoice, prompt: prompt, code: code, explanation: explanation,
		payload: payload{Options: append([]string{correct}, wrong...), Correct: 0},
	}
}

// typeQ: user ngetik nama tipe. accepted[0] = jawaban yang ditampilin sebagai kunci.
func typeQ(code string, accepted []string, explanation string) questionSpec {
	return questionSpec{
		kind: coursesvc.KindTypeAnswer, prompt: askType, code: code, explanation: explanation,
		payload: payload{Accepted: accepted},
	}
}

// write: user nulis nilai/data sendiri, dicek pakai schema. sample = salah
// satu contoh jawaban yang bener (ditampilin abis submit).
func write(prompt, placeholder string, schema coursesvc.Schema, sample, explanation string) questionSpec {
	return questionSpec{
		kind: coursesvc.KindWriteValue, prompt: prompt, explanation: explanation,
		payload: payload{Schema: &schema, Placeholder: placeholder, Sample: sample},
	}
}

func writeJSON(prompt string, schema coursesvc.Schema, sample, explanation string) questionSpec {
	q := write(prompt, "", schema, sample, explanation)
	q.payload.StrictJSON = true
	return q
}

// --- helper skema ---

func str() coursesvc.Schema     { return coursesvc.Schema{Type: "string", NotEmpty: true} }
func num() coursesvc.Schema     { return coursesvc.Schema{Type: "number"} }
func boolean() coursesvc.Schema { return coursesvc.Schema{Type: "boolean"} }
func null() coursesvc.Schema    { return coursesvc.Schema{Type: "null"} }

func obj(fields map[string]coursesvc.Schema) coursesvc.Schema {
	return coursesvc.Schema{Type: "object", Required: fields}
}

func arr(minItems int, items coursesvc.Schema) coursesvc.Schema {
	return coursesvc.Schema{Type: "array", MinItems: minItems, Items: &items}
}

func oneOf(values ...string) coursesvc.Schema {
	return coursesvc.Schema{Type: "string", OneOf: values}
}

func minNum(min float64) coursesvc.Schema {
	return coursesvc.Schema{Type: "number", Min: &min}
}

// --- helper materi ---

func heading(text string) coursesvc.MaterialBlock {
	return coursesvc.MaterialBlock{Type: "heading", Text: text}
}

func para(text string) coursesvc.MaterialBlock {
	return coursesvc.MaterialBlock{Type: "paragraph", Text: text}
}

func code(title, src, output string) coursesvc.MaterialBlock {
	return coursesvc.MaterialBlock{Type: "code", Title: title, Code: src, Output: output}
}

func example(title, text, src, output string) coursesvc.MaterialBlock {
	return coursesvc.MaterialBlock{Type: "example", Title: title, Text: text, Code: src, Output: output}
}

func tip(text string) coursesvc.MaterialBlock {
	return coursesvc.MaterialBlock{Type: "tip", Text: text}
}

func list(items ...string) coursesvc.MaterialBlock {
	return coursesvc.MaterialBlock{Type: "list", Items: items}
}

func table(headers []string, rows ...[]string) coursesvc.MaterialBlock {
	return coursesvc.MaterialBlock{Type: "table", Headers: headers, Rows: rows}
}

func modules() []moduleSpec {
	return []moduleSpec{
		variabelModule(),
		{
			code: "looping", title: "Looping",
			description:       "Ngulang perintah pakai for, while, dan ngolah array of object.",
			examQuestionCount: 15, examTimeLimitSeconds: 420, examPassPercent: 80,
		},
		{
			code: "async", title: "Asynchronous",
			description:       "Callback, Promise, async/await, dan ngambil data dari API.",
			examQuestionCount: 15, examTimeLimitSeconds: 420, examPassPercent: 80,
		},
	}
}
