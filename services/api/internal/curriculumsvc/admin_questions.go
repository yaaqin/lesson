package curriculumsvc

import (
	"context"
	"errors"
)

var (
	ErrTooFewOptions        = errors.New("minimal 2 opsi jawaban")
	ErrNoCorrectOption      = errors.New("harus ada 1 opsi yang ditandai benar")
	ErrMultipleCorrect      = errors.New("cuma boleh 1 opsi yang ditandai benar")
	ErrDuplicateOption      = errors.New("nilai opsi jangan ada yang sama")
	ErrMissingCorrectAnswer = errors.New("jawaban benar wajib diisi buat soal essay")
	ErrInvalidQuestionType  = errors.New("tipe soal gak dikenal")
)

// AdminListQuestions: semua soal 1 challenge (termasuk draft/archived, beda
// dengan loadPublishedBank di gameplay.go yang cuma ambil status published).
func (s *Service) AdminListQuestions(ctx context.Context, challengeID string) ([]AdminQuestion, error) {
	rows, err := s.db.Query(ctx, `
		SELECT q.id, q.question_type, q.prompt, q.status, q.correct_answer_value, qo.option_value, qo.is_correct
		FROM questions q
		LEFT JOIN question_options qo ON qo.question_id = q.id
		WHERE q.challenge_id = $1
		ORDER BY q.created_at, qo.option_value
	`, challengeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	byID := map[string]*AdminQuestion{}
	order := []string{}
	for rows.Next() {
		var (
			id, qType, prompt, status string
			correctAnswer             *float64
			optValue                  *float64
			isCorrect                 *bool
		)
		if err := rows.Scan(&id, &qType, &prompt, &status, &correctAnswer, &optValue, &isCorrect); err != nil {
			return nil, err
		}
		q, ok := byID[id]
		if !ok {
			q = &AdminQuestion{ID: id, Type: qType, Prompt: prompt, Status: status}
			if qType == QuestionTypeEssayNumeric {
				q.CorrectAnswerValue = correctAnswer
			} else {
				q.Options = []AdminQuestionOption{}
			}
			byID[id] = q
			order = append(order, id)
		}
		if optValue != nil && isCorrect != nil {
			q.Options = append(q.Options, AdminQuestionOption{Value: *optValue, IsCorrect: *isCorrect})
		}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	questions := make([]AdminQuestion, 0, len(order))
	for _, id := range order {
		questions = append(questions, *byID[id])
	}
	return questions, nil
}

func validateOptions(options []AdminQuestionOption) (float64, error) {
	if len(options) < 2 {
		return 0, ErrTooFewOptions
	}
	seen := make(map[float64]bool, len(options))
	var correct *float64
	for _, o := range options {
		if seen[o.Value] {
			return 0, ErrDuplicateOption
		}
		seen[o.Value] = true
		if o.IsCorrect {
			if correct != nil {
				return 0, ErrMultipleCorrect
			}
			v := o.Value
			correct = &v
		}
	}
	if correct == nil {
		return 0, ErrNoCorrectOption
	}
	return *correct, nil
}

// resolveCorrectAnswer: buat multiple_choice, jawaban benar diturunkan dari
// options[].isCorrect. Buat essay_numeric, gak ada options sama sekali --
// jawaban benar diinput langsung lewat correctAnswerValue.
func resolveCorrectAnswer(questionType string, options []AdminQuestionOption, correctAnswerValue *float64) (float64, error) {
	switch questionType {
	case QuestionTypeEssayNumeric:
		if correctAnswerValue == nil {
			return 0, ErrMissingCorrectAnswer
		}
		return *correctAnswerValue, nil
	case QuestionTypeMultipleChoice:
		return validateOptions(options)
	default:
		return 0, ErrInvalidQuestionType
	}
}

func (s *Service) AdminCreateQuestion(
	ctx context.Context, challengeID, questionType, prompt, status string,
	options []AdminQuestionOption, correctAnswerValue *float64,
) (string, error) {
	correctValue, err := resolveCorrectAnswer(questionType, options, correctAnswerValue)
	if err != nil {
		return "", err
	}

	var questionID string
	if err := s.db.QueryRow(ctx, `
		INSERT INTO questions (challenge_id, question_type, prompt, correct_answer_value, status)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`, challengeID, questionType, prompt, correctValue, status).Scan(&questionID); err != nil {
		return "", err
	}

	if questionType == QuestionTypeMultipleChoice {
		if err := s.insertOptions(ctx, questionID, options); err != nil {
			return "", err
		}
	}
	return questionID, nil
}

func (s *Service) AdminUpdateQuestion(
	ctx context.Context, questionID, questionType, prompt, status string,
	options []AdminQuestionOption, correctAnswerValue *float64,
) error {
	correctValue, err := resolveCorrectAnswer(questionType, options, correctAnswerValue)
	if err != nil {
		return err
	}

	if _, err := s.db.Exec(ctx, `
		UPDATE questions SET question_type = $1, prompt = $2, correct_answer_value = $3, status = $4 WHERE id = $5
	`, questionType, prompt, correctValue, status, questionID); err != nil {
		return err
	}

	if _, err := s.db.Exec(ctx, `DELETE FROM question_options WHERE question_id = $1`, questionID); err != nil {
		return err
	}
	if questionType == QuestionTypeMultipleChoice {
		return s.insertOptions(ctx, questionID, options)
	}
	return nil
}

func (s *Service) insertOptions(ctx context.Context, questionID string, options []AdminQuestionOption) error {
	for _, opt := range options {
		if _, err := s.db.Exec(ctx, `
			INSERT INTO question_options (question_id, option_value, is_correct)
			VALUES ($1, $2, $3)
		`, questionID, opt.Value, opt.IsCorrect); err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) AdminDeleteQuestion(ctx context.Context, questionID string) error {
	_, err := s.db.Exec(ctx, `DELETE FROM questions WHERE id = $1`, questionID)
	return err
}
