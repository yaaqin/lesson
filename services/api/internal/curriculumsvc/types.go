package curriculumsvc

// --- Read (userApp) ---

type Tier struct {
	ID   string `json:"id"`
	Code string `json:"code"`
	Name string `json:"name"`
}

type Batch struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type ChallengeListItem struct {
	ID                    string `json:"id"`
	Name                  string `json:"name"`
	IsExam                bool   `json:"isExam"`
	QuestionCountRequired int    `json:"questionCountRequired"`
	PassThresholdPercent  int    `json:"passThresholdPercent"`
	TimeLimitSeconds      int    `json:"timeLimitSeconds"`
	Completed             bool   `json:"completed"`
	HasEssay              bool   `json:"hasEssay"`
}

// --- Gameplay ---

const (
	QuestionTypeMultipleChoice = "multiple_choice"
	QuestionTypeEssayNumeric   = "essay_numeric"
)

type QuestionOption struct {
	Value     float64 `json:"value"`
	IsCorrect bool    `json:"isCorrect"`
}

// SessionQuestion: buat multiple_choice, Options keisi & CorrectAnswerValue
// nil. Buat essay_numeric, Options kosong & CorrectAnswerValue keisi -- klien
// yang nampilin keypad angka & ngecek sendiri kebenarannya (simplifikasi yang
// sama kayak MC yang juga nampilin options[].isCorrect ke klien).
type SessionQuestion struct {
	ID                 string           `json:"id"`
	Type               string           `json:"type"`
	Prompt             string           `json:"prompt"`
	Options            []QuestionOption `json:"options,omitempty"`
	CorrectAnswerValue *float64         `json:"correctAnswerValue,omitempty"`
}

// snapshotQuestion sama isinya dengan SessionQuestion, dipisah tipe biar
// perubahan format JSON respons API gak otomatis mengubah format snapshot
// yang sudah tersimpan di kolom questions_snapshot milik attempt lama.
type snapshotQuestion struct {
	ID                 string           `json:"id"`
	Type               string           `json:"type"`
	Prompt             string           `json:"prompt"`
	Options            []QuestionOption `json:"options,omitempty"`
	CorrectAnswerValue *float64         `json:"correctAnswerValue,omitempty"`
}

type StartResult struct {
	AttemptID            string            `json:"attemptId"`
	ChallengeName        string            `json:"challengeName"`
	IsExam               bool              `json:"isExam"`
	TimeLimitSeconds     int               `json:"timeLimitSeconds"`
	PassThresholdPercent int               `json:"passThresholdPercent"`
	Questions            []SessionQuestion `json:"questions"`
}

type SubmitAnswer struct {
	QuestionID    string   `json:"questionId"`
	SelectedValue *float64 `json:"selectedValue"`
}

type SubmitResult struct {
	ScorePercent   int  `json:"scorePercent"`
	Passed         bool `json:"passed"`
	CorrectCount   int  `json:"correctCount"`
	TotalQuestions int  `json:"totalQuestions"`
	LivesRemaining int  `json:"livesRemaining"`
	CurrentStreak  int  `json:"currentStreak"`
}

type MeInfo struct {
	ID             string `json:"id"`
	Email          string `json:"email"`
	DisplayName    string `json:"displayName"`
	Role           string `json:"role"`
	CurrentStreak  int    `json:"currentStreak"`
	LongestStreak  int    `json:"longestStreak"`
	LivesRemaining int    `json:"livesRemaining"`
}

// --- Admin (dashboard) ---

type AdminChallenge struct {
	ID                    string `json:"id"`
	Name                  string `json:"name"`
	IsExam                bool   `json:"isExam"`
	TimeLimitSeconds      int    `json:"timeLimitSeconds"`
	QuestionCountRequired int    `json:"questionCountRequired"`
	PassThresholdPercent  int    `json:"passThresholdPercent"`
	OptionCount           int    `json:"optionCount"`
	QuestionBankSize      int    `json:"questionBankSize"`
}

type AdminBatch struct {
	ID         string           `json:"id"`
	Name       string           `json:"name"`
	Challenges []AdminChallenge `json:"challenges"`
}

type AdminTier struct {
	ID      string       `json:"id"`
	Code    string       `json:"code"`
	Name    string       `json:"name"`
	Batches []AdminBatch `json:"batches"`
}

type AdminQuestionOption struct {
	Value     float64 `json:"value"`
	IsCorrect bool    `json:"isCorrect"`
}

type AdminQuestion struct {
	ID                 string                `json:"id"`
	Type               string                `json:"type"`
	Prompt             string                `json:"prompt"`
	Status             string                `json:"status"`
	Options            []AdminQuestionOption `json:"options,omitempty"`
	CorrectAnswerValue *float64              `json:"correctAnswerValue,omitempty"`
}
