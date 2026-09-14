package curriculumsvc

import "errors"

var (
	ErrNotFound        = errors.New("not found")
	ErrNoLives         = errors.New("no lives remaining")
	ErrNotEnoughBank   = errors.New("not enough published questions in bank")
	ErrAttemptNotFound = errors.New("attempt not found")
	ErrAttemptFinished = errors.New("attempt already submitted")
)
