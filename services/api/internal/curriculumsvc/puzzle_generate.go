package curriculumsvc

import (
	"errors"
	"math/rand/v2"
)

// GenerateAdditionGrid: acak permutasi 1..size*size ke kotak NxN, lalu pilih
// max(2,size) sel acak buat jadi "given" (kesisanya kosong buat diisi user).
// rowSums/colSums gak disimpen di sini -- dihitung on-the-fly dari
// SolutionGrid di derivePuzzle (server-side, satu sumber kebenaran). Dipakai
// cmd/seed-puzzle & admin_puzzle_questions.go (endpoint dashboard) -- sengaja
// satu implementasi biar gak ada 2 versi yang bisa ketinggalan salah satu
// pas diubah.
func GenerateAdditionGrid(size int) (solution [][]int, givenMask [][]bool) {
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

// SolveCryptarithm: brute-force nyari 1 solusi valid (huruf -> digit unik
// 0-9, huruf awal kata gak boleh 0) buat words[0]+words[1]+...=result. Dipake
// buat MEMVALIDASI puzzle beneran solvable & benar secara aritmatika sebelum
// disimpen -- bukan cuma percaya hasil hitungan manual (dipakai cmd/seed-puzzle
// & endpoint admin create/update cryptarithm).
func SolveCryptarithm(words []string, result string) (map[string]int, error) {
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
