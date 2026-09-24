package curriculumsvc

import "time"

// --- Read (userApp) ---

type Tier struct {
	ID        string `json:"id"`
	Code      string `json:"code"`
	Name      string `json:"name"`
	UsesBatch bool   `json:"usesBatch"`
}

type Batch struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// Category: ganti batch buat tier yang uses_batch = false (mis. "umum"),
// ngelompokin challenge yang strukturnya beda-beda kayak puzzle (FSD.md 3.3).
type Category struct {
	ID   string `json:"id"`
	Code string `json:"code"`
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
	QuestionTypeGridPuzzle     = "grid_puzzle"
)

const (
	PuzzleKindAdditionGrid = "addition_grid"
	PuzzleKindCryptarithm  = "cryptarithm"
)

type QuestionOption struct {
	Value     float64 `json:"value"`
	IsCorrect bool    `json:"isCorrect"`
}

// PuzzlePayload: bentuk publik puzzle (grid_puzzle) yang dikirim ke klien.
// Kind nentuin gimana klien nge-render ("addition_grid" -> kotak NxN angka,
// "cryptarithm" -> persamaan huruf). Given = sel/huruf yang udah kekasih tau
// nilainya (key "r{i}c{j}" buat grid, huruf itu sendiri buat cryptarithm).
// BlankKeys = key yang harus diisi user. RowSums/ColSums cuma kepake buat
// addition_grid. SENGAJA gak ada field solusi di sini -- solusinya cuma ada di
// snapshotQuestion (server-side), dicek pas SubmitAttempt.
type PuzzlePayload struct {
	Kind      string             `json:"kind"`
	Size      int                `json:"size,omitempty"`
	Words     []string           `json:"words,omitempty"`
	Given     map[string]float64 `json:"given,omitempty"`
	BlankKeys []string           `json:"blankKeys"`
	RowSums   []float64          `json:"rowSums,omitempty"`
	ColSums   []float64          `json:"colSums,omitempty"`
}

// SessionQuestion: buat multiple_choice, Options keisi & CorrectAnswerValue
// nil. Buat essay_numeric, Options kosong & CorrectAnswerValue keisi -- klien
// yang nampilin keypad angka & ngecek sendiri kebenarannya (simplifikasi yang
// sama kayak MC yang juga nampilin options[].isCorrect ke klien). Buat
// grid_puzzle, Puzzle keisi (tanpa solusi -- lihat PuzzlePayload).
type SessionQuestion struct {
	ID                 string           `json:"id"`
	Type               string           `json:"type"`
	Prompt             string           `json:"prompt"`
	Options            []QuestionOption `json:"options,omitempty"`
	CorrectAnswerValue *float64         `json:"correctAnswerValue,omitempty"`
	Puzzle             *PuzzlePayload   `json:"puzzle,omitempty"`
}

// snapshotQuestion sama isinya dengan SessionQuestion, dipisah tipe biar
// perubahan format JSON respons API gak otomatis mengubah format snapshot
// yang sudah tersimpan di kolom questions_snapshot milik attempt lama. Beda
// dari SessionQuestion, di sini ADA PuzzleSolution -- solusi grid_puzzle,
// dipakai SubmitAttempt buat nyocokin jawaban, gak pernah ikut ke-marshal ke
// respons StartChallenge karena SessionQuestion gak punya field ini.
type snapshotQuestion struct {
	ID                 string             `json:"id"`
	Type               string             `json:"type"`
	Prompt             string             `json:"prompt"`
	Options            []QuestionOption   `json:"options,omitempty"`
	CorrectAnswerValue *float64           `json:"correctAnswerValue,omitempty"`
	Puzzle             *PuzzlePayload     `json:"puzzle,omitempty"`
	PuzzleSolution     map[string]float64 `json:"puzzleSolution,omitempty"`
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
	QuestionID    string             `json:"questionId"`
	SelectedValue *float64           `json:"selectedValue"`
	SelectedGrid  map[string]float64 `json:"selectedGrid,omitempty"`
}

type SubmitResult struct {
	ScorePercent   int  `json:"scorePercent"`
	Passed         bool `json:"passed"`
	CorrectCount   int  `json:"correctCount"`
	TotalQuestions int  `json:"totalQuestions"`
	LivesRemaining int  `json:"livesRemaining"`
	CurrentStreak  int  `json:"currentStreak"`
	// RiskLevel: hasil scoring anti-cheating (normal | low_confidence | review).
	// Skor mentahnya sengaja gak dikirim ke klien, cuma disimpen di DB.
	RiskLevel string `json:"riskLevel"`
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

type AdminCategory struct {
	ID         string           `json:"id"`
	Code       string           `json:"code"`
	Name       string           `json:"name"`
	Challenges []AdminChallenge `json:"challenges"`
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

// AdminPuzzleQuestion: view admin-only 1 soal grid_puzzle -- BEDA dari
// PuzzlePayload (yang dikirim ke userApp lewat StartChallenge, solusinya
// sengaja gak ada di situ). Di sini solusinya SENGAJA ikut biar admin bisa
// lihat & edit lewat dashboard.
type AdminPuzzleQuestion struct {
	ID     string `json:"id"`
	Kind   string `json:"kind"`
	Prompt string `json:"prompt"`
	Status string `json:"status"`
	// addition_grid
	Size         int      `json:"size,omitempty"`
	SolutionGrid [][]int  `json:"solutionGrid,omitempty"`
	GivenMask    [][]bool `json:"givenMask,omitempty"`
	// cryptarithm -- Words termasuk kata hasil di elemen terakhir (konsisten
	// sama shape cryptarithmRaw di puzzle.go), Solution diisi server lewat
	// SolveCryptarithm, gak pernah diketik manual.
	Words    []string       `json:"words,omitempty"`
	Solution map[string]int `json:"solution,omitempty"`
}

// AdminPuzzleUpsert: body request create/update, Kind nentuin field mana yang
// dipakai (sama pola diskriminasi kayak upsertQuestionRequest di routes_admin.go).
type AdminPuzzleUpsert struct {
	Kind         string
	Prompt       string
	Status       string
	Size         int
	SolutionGrid [][]int
	GivenMask    [][]bool
	Words        []string // addend, TANPA kata hasil
	Result       string
}

type AdminUserListItem struct {
	ID             string    `json:"id"`
	Email          string    `json:"email"`
	DisplayName    string    `json:"displayName"`
	CurrentStreak  int       `json:"currentStreak"`
	LongestStreak  int       `json:"longestStreak"`
	LivesRemaining int       `json:"livesRemaining"`
	CreatedAt      time.Time `json:"createdAt"`
}

type AdminUserListResult struct {
	Items    []AdminUserListItem `json:"items"`
	Total    int                 `json:"total"`
	Page     int                 `json:"page"`
	PageSize int                 `json:"pageSize"`
}

type AdminUserDetail struct {
	ID               string     `json:"id"`
	Email            string     `json:"email"`
	DisplayName      string     `json:"displayName"`
	Role             string     `json:"role"`
	CurrentStreak    int        `json:"currentStreak"`
	LongestStreak    int        `json:"longestStreak"`
	LastActiveDate   *time.Time `json:"lastActiveDate,omitempty"`
	LivesRemaining   int        `json:"livesRemaining"`
	LivesLastResetAt time.Time  `json:"livesLastResetAt"`
	TotalAttempts    int        `json:"totalAttempts"`
	PassedAttempts   int        `json:"passedAttempts"`
	CreatedAt        time.Time  `json:"createdAt"`
}
