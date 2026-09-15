package curriculumsvc

import (
	"encoding/json"
	"fmt"
	"sort"
)

// Raw JSONB shapes disimpen di puzzle_questions.payload, per "kind".

type additionGridRaw struct {
	Size         int      `json:"size"`
	SolutionGrid [][]int  `json:"solutionGrid"`
	GivenMask    [][]bool `json:"givenMask"`
}

type cryptarithmRaw struct {
	Words    []string       `json:"words"`
	Solution map[string]int `json:"solution"`
}

// derivePuzzle ngubah payload mentah (JSONB, per kind) jadi PuzzlePayload
// publik (buat dikirim ke klien) + solution map (server-side aja, dipakai
// SubmitAttempt buat nyocokin jawaban). Kunci di kedua map konsisten: "r{i}c{j}"
// buat addition_grid, huruf itu sendiri buat cryptarithm -- jadi scoring-nya
// generik, gak perlu tau kind sama sekali (lihat SubmitAttempt).
func derivePuzzle(kind string, raw []byte) (*PuzzlePayload, map[string]float64, error) {
	switch kind {
	case "addition_grid":
		var p additionGridRaw
		if err := json.Unmarshal(raw, &p); err != nil {
			return nil, nil, err
		}
		given := map[string]float64{}
		solution := map[string]float64{}
		blankKeys := []string{}
		rowSums := make([]float64, p.Size)
		colSums := make([]float64, p.Size)
		for i := 0; i < p.Size; i++ {
			for j := 0; j < p.Size; j++ {
				key := fmt.Sprintf("r%dc%d", i, j)
				val := float64(p.SolutionGrid[i][j])
				rowSums[i] += val
				colSums[j] += val
				if p.GivenMask[i][j] {
					given[key] = val
				} else {
					solution[key] = val
					blankKeys = append(blankKeys, key)
				}
			}
		}
		sort.Strings(blankKeys)
		return &PuzzlePayload{
			Kind:      kind,
			Size:      p.Size,
			Given:     given,
			BlankKeys: blankKeys,
			RowSums:   rowSums,
			ColSums:   colSums,
		}, solution, nil

	case "cryptarithm":
		var p cryptarithmRaw
		if err := json.Unmarshal(raw, &p); err != nil {
			return nil, nil, err
		}
		solution := map[string]float64{}
		blankKeys := make([]string, 0, len(p.Solution))
		for letter, digit := range p.Solution {
			solution[letter] = float64(digit)
			blankKeys = append(blankKeys, letter)
		}
		sort.Strings(blankKeys)
		return &PuzzlePayload{
			Kind:      kind,
			Words:     p.Words,
			BlankKeys: blankKeys,
		}, solution, nil

	default:
		return nil, nil, fmt.Errorf("unknown puzzle kind: %s", kind)
	}
}
