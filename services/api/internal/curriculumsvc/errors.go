package curriculumsvc

import "errors"

var (
	ErrNotFound          = errors.New("not found")
	ErrNoLives           = errors.New("no lives remaining")
	ErrNotEnoughBank     = errors.New("not enough published questions in bank")
	ErrAttemptNotFound   = errors.New("attempt not found")
	ErrAttemptFinished   = errors.New("attempt already submitted")
	ErrInvalidPuzzleKind = errors.New("kind puzzle gak dikenal (harus addition_grid atau cryptarithm)")
	ErrInvalidPuzzleGrid = errors.New("grid gak valid -- ukuran gak cocok atau isinya bukan permutasi 1..size² yang unik, minimal harus ada 1 kotak kosong")
)
