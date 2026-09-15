package curriculumsvc

import (
	"context"
	"encoding/json"
	"fmt"
)

// AdminListPuzzleQuestions: semua soal grid_puzzle 1 challenge, LENGKAP sama
// solusinya (beda dari AdminListQuestions yang cuma balikin {id,type,prompt,status}
// generik buat semua tipe soal) -- dipakai halaman CRUD khusus puzzle di dashboard.
func (s *Service) AdminListPuzzleQuestions(ctx context.Context, challengeID string) ([]AdminPuzzleQuestion, error) {
	rows, err := s.db.Query(ctx, `
		SELECT q.id, q.prompt, q.status, pq.kind, pq.payload
		FROM questions q
		JOIN puzzle_questions pq ON pq.question_id = q.id
		WHERE q.challenge_id = $1
		ORDER BY q.created_at
	`, challengeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	questions := []AdminPuzzleQuestion{}
	for rows.Next() {
		var (
			id, prompt, status, kind string
			payload                  []byte
		)
		if err := rows.Scan(&id, &prompt, &status, &kind, &payload); err != nil {
			return nil, err
		}
		q, err := unmarshalAdminPuzzleQuestion(id, prompt, status, kind, payload)
		if err != nil {
			return nil, err
		}
		questions = append(questions, *q)
	}
	return questions, rows.Err()
}

func unmarshalAdminPuzzleQuestion(id, prompt, status, kind string, payload []byte) (*AdminPuzzleQuestion, error) {
	q := &AdminPuzzleQuestion{ID: id, Kind: kind, Prompt: prompt, Status: status}
	switch kind {
	case PuzzleKindAdditionGrid:
		var raw additionGridRaw
		if err := json.Unmarshal(payload, &raw); err != nil {
			return nil, err
		}
		q.Size = raw.Size
		q.SolutionGrid = raw.SolutionGrid
		q.GivenMask = raw.GivenMask
	case PuzzleKindCryptarithm:
		var raw cryptarithmRaw
		if err := json.Unmarshal(payload, &raw); err != nil {
			return nil, err
		}
		q.Words = raw.Words
		q.Solution = raw.Solution
	default:
		return nil, ErrInvalidPuzzleKind
	}
	return q, nil
}

// resolvePuzzleUpsert: validasi req & bentuk AdminPuzzleQuestion + payload JSONB
// yang siap disimpen -- dipakai bareng sama AdminCreatePuzzleQuestion &
// AdminUpdatePuzzleQuestion biar validasinya satu tempat aja.
func resolvePuzzleUpsert(req AdminPuzzleUpsert) (*AdminPuzzleQuestion, []byte, *int, error) {
	status := req.Status
	if status == "" {
		status = "published"
	}

	switch req.Kind {
	case PuzzleKindAdditionGrid:
		if err := validateAdditionGrid(req.Size, req.SolutionGrid, req.GivenMask); err != nil {
			return nil, nil, nil, err
		}
		prompt := req.Prompt
		if prompt == "" {
			prompt = fmt.Sprintf(
				"Isi kotak kosong dengan angka 1–%d (masing-masing dipakai tepat sekali) supaya tiap baris & kolom sesuai jumlah yang tertera.",
				req.Size*req.Size,
			)
		}
		payload, err := json.Marshal(additionGridRaw{Size: req.Size, SolutionGrid: req.SolutionGrid, GivenMask: req.GivenMask})
		if err != nil {
			return nil, nil, nil, err
		}
		size := req.Size
		return &AdminPuzzleQuestion{
			Kind: req.Kind, Prompt: prompt, Status: status,
			Size: req.Size, SolutionGrid: req.SolutionGrid, GivenMask: req.GivenMask,
		}, payload, &size, nil

	case PuzzleKindCryptarithm:
		if len(req.Words) == 0 || req.Result == "" {
			return nil, nil, nil, ErrInvalidPuzzleGrid
		}
		solution, err := SolveCryptarithm(req.Words, req.Result)
		if err != nil {
			return nil, nil, nil, fmt.Errorf("puzzle gak solvable: %w", err)
		}
		words := append(append([]string{}, req.Words...), req.Result)
		prompt := req.Prompt
		if prompt == "" {
			prompt = fmt.Sprintf(
				"Setiap huruf mewakili satu angka (0-9), gak boleh ada dua huruf beda yang sama angkanya. Cari nilai tiap huruf supaya %s = %s benar.",
				joinPlus(req.Words), req.Result,
			)
		}
		payload, err := json.Marshal(cryptarithmRaw{Words: words, Solution: solution})
		if err != nil {
			return nil, nil, nil, err
		}
		return &AdminPuzzleQuestion{
			Kind: req.Kind, Prompt: prompt, Status: status,
			Words: words, Solution: solution,
		}, payload, nil, nil

	default:
		return nil, nil, nil, ErrInvalidPuzzleKind
	}
}

func (s *Service) AdminCreatePuzzleQuestion(ctx context.Context, challengeID string, req AdminPuzzleUpsert) (*AdminPuzzleQuestion, error) {
	resp, payload, gridSize, err := resolvePuzzleUpsert(req)
	if err != nil {
		return nil, err
	}

	var questionID string
	if err := s.db.QueryRow(ctx, `
		INSERT INTO questions (challenge_id, question_type, prompt, status)
		VALUES ($1, $2, $3, $4)
		RETURNING id
	`, challengeID, QuestionTypeGridPuzzle, resp.Prompt, resp.Status).Scan(&questionID); err != nil {
		return nil, err
	}
	if _, err := s.db.Exec(ctx, `
		INSERT INTO puzzle_questions (question_id, kind, grid_size, payload)
		VALUES ($1, $2, $3, $4)
	`, questionID, resp.Kind, gridSize, payload); err != nil {
		return nil, err
	}

	resp.ID = questionID
	return resp, nil
}

func (s *Service) AdminUpdatePuzzleQuestion(ctx context.Context, questionID string, req AdminPuzzleUpsert) (*AdminPuzzleQuestion, error) {
	resp, payload, gridSize, err := resolvePuzzleUpsert(req)
	if err != nil {
		return nil, err
	}

	if _, err := s.db.Exec(ctx, `
		UPDATE questions SET prompt = $1, status = $2 WHERE id = $3
	`, resp.Prompt, resp.Status, questionID); err != nil {
		return nil, err
	}
	if _, err := s.db.Exec(ctx, `
		UPDATE puzzle_questions SET kind = $1, grid_size = $2, payload = $3 WHERE question_id = $4
	`, resp.Kind, gridSize, payload, questionID); err != nil {
		return nil, err
	}

	resp.ID = questionID
	return resp, nil
}

// validateAdditionGrid: size wajar (3-6), dimensi solutionGrid/givenMask harus
// cocok size×size, isi solutionGrid harus PERSIS permutasi 1..size² (gak boleh
// dobel/di luar range -- ini exact bug yang bikin user salah masukin jawaban di
// userApp, lihat AdditionGridAnswer di apps/web), dan minimal 1 kotak blank
// (kalau semua "given" gak ada yang perlu diisi).
func validateAdditionGrid(size int, solutionGrid [][]int, givenMask [][]bool) error {
	if size < 3 || size > 6 {
		return ErrInvalidPuzzleGrid
	}
	if len(solutionGrid) != size || len(givenMask) != size {
		return ErrInvalidPuzzleGrid
	}

	seen := make([]bool, size*size+1)
	blankCount := 0
	for i := 0; i < size; i++ {
		if len(solutionGrid[i]) != size || len(givenMask[i]) != size {
			return ErrInvalidPuzzleGrid
		}
		for j := 0; j < size; j++ {
			v := solutionGrid[i][j]
			if v < 1 || v > size*size || seen[v] {
				return ErrInvalidPuzzleGrid
			}
			seen[v] = true
			if !givenMask[i][j] {
				blankCount++
			}
		}
	}
	if blankCount == 0 {
		return ErrInvalidPuzzleGrid
	}
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
