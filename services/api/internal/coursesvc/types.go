package coursesvc

import (
	"encoding/json"
	"errors"
)

var (
	ErrNotFound        = errors.New("not found")
	ErrLocked          = errors.New("masih kekunci")
	ErrEmptyBank       = errors.New("bank soal kosong")
	ErrAttemptNotFound = errors.New("attempt not found")
	ErrAttemptFinished = errors.New("attempt already submitted")
)

const (
	KindMultipleChoice = "multiple_choice"
	KindTypeAnswer     = "type_answer"
	KindWriteValue     = "write_value"
)

// examGraceSeconds: toleransi telat submit final test (latency jaringan) --
// lewat dari batas waktu + ini, final test dianggap gak lulus.
const examGraceSeconds = 10

// MaterialBlock: satu blok materi (kolom course_levels.material). Type:
//   - heading   -> Text
//   - paragraph -> Text (boleh pakai `kode` inline)
//   - code      -> Code, Output opsional (hasil console.log), Title opsional
//   - example   -> contoh kasus nyata: Title, Text, Code, Output opsional
//   - tip       -> Text (catatan / jebakan yang sering kejadian)
//   - list      -> Items
//   - table     -> Headers, Rows
type MaterialBlock struct {
	Type    string     `json:"type"`
	Title   string     `json:"title,omitempty"`
	Text    string     `json:"text,omitempty"`
	Code    string     `json:"code,omitempty"`
	Output  string     `json:"output,omitempty"`
	Items   []string   `json:"items,omitempty"`
	Headers []string   `json:"headers,omitempty"`
	Rows    [][]string `json:"rows,omitempty"`
}

// --- Overview (halaman daftar module) ---

type ModuleItem struct {
	Code        string      `json:"code"`
	Title       string      `json:"title"`
	Description string      `json:"description"`
	Locked      bool        `json:"locked"`
	Completed   bool        `json:"completed"`
	ComingSoon  bool        `json:"comingSoon"` // belum ada level sama sekali
	Levels      []LevelItem `json:"levels"`
	Exam        ExamInfo    `json:"exam"`
}

type LevelItem struct {
	ID            string `json:"id"`
	Code          string `json:"code"`
	Title         string `json:"title"`
	Summary       string `json:"summary"`
	QuestionCount int    `json:"questionCount"`
	Locked        bool   `json:"locked"`
	Passed        bool   `json:"passed"`
}

type ExamInfo struct {
	Locked           bool `json:"locked"`
	Passed           bool `json:"passed"`
	QuestionCount    int  `json:"questionCount"`
	BankSize         int  `json:"bankSize"`
	TimeLimitSeconds int  `json:"timeLimitSeconds"`
	PassPercent      int  `json:"passPercent"`
	BestScore        *int `json:"bestScore,omitempty"`
	BestTimeMs       *int `json:"bestTimeMs,omitempty"` // waktu tercepat dari final test yang lulus
}

// --- Level (materi) ---

type LevelDetail struct {
	ID            string          `json:"id"`
	Code          string          `json:"code"`
	Title         string          `json:"title"`
	Summary       string          `json:"summary"`
	ModuleCode    string          `json:"moduleCode"`
	ModuleTitle   string          `json:"moduleTitle"`
	LevelNumber   int             `json:"levelNumber"`
	QuestionCount int             `json:"questionCount"`
	Passed        bool            `json:"passed"`
	Material      json.RawMessage `json:"material"`
}

// --- Attempt ---

// Question: bentuk publik soal -- kunci jawaban gak pernah ikut.
type Question struct {
	ID          string   `json:"id"`
	Kind        string   `json:"kind"`
	Prompt      string   `json:"prompt"`
	Code        string   `json:"code,omitempty"`
	Options     []string `json:"options,omitempty"`
	Placeholder string   `json:"placeholder,omitempty"`
	StrictJSON  bool     `json:"strictJson,omitempty"`
}

type StartResult struct {
	AttemptID        string     `json:"attemptId"`
	IsExam           bool       `json:"isExam"`
	Title            string     `json:"title"`
	ModuleCode       string     `json:"moduleCode"`
	PassPercent      int        `json:"passPercent"`
	TimeLimitSeconds int        `json:"timeLimitSeconds,omitempty"` // cuma final test
	Questions        []Question `json:"questions"`
}

// payload: isi kolom course_questions.payload, beda field kepake per kind.
type payload struct {
	Options     []string `json:"options,omitempty"`
	Correct     int      `json:"correct"`
	Accepted    []string `json:"accepted,omitempty"`
	Schema      *Schema  `json:"schema,omitempty"`
	StrictJSON  bool     `json:"strictJson,omitempty"`
	Placeholder string   `json:"placeholder,omitempty"`
	Sample      string   `json:"sample,omitempty"`
}

// snapshotQuestion: soal + kunci (opsi PG udah diacak) yang disimpen di
// course_attempts.questions_snapshot.
type snapshotQuestion struct {
	ID          string  `json:"id"`
	Kind        string  `json:"kind"`
	Prompt      string  `json:"prompt"`
	Code        string  `json:"code,omitempty"`
	Explanation string  `json:"explanation,omitempty"`
	Payload     payload `json:"payload"`
}

type Answer struct {
	QuestionID string `json:"questionId"`
	Choice     *int   `json:"choice,omitempty"` // multiple_choice: indeks opsi
	Text       string `json:"text,omitempty"`   // type_answer / write_value
}

type QuestionResult struct {
	QuestionID  string `json:"questionId"`
	Correct     bool   `json:"correct"`
	Feedback    string `json:"feedback,omitempty"` // alasan salah (write_value)
	Expected    string `json:"expected,omitempty"` // kunci / contoh jawaban yang bener
	Explanation string `json:"explanation,omitempty"`
}

type SubmitResult struct {
	Passed       bool             `json:"passed"`
	ScorePercent int              `json:"scorePercent"`
	CorrectCount int              `json:"correctCount"`
	Total        int              `json:"total"`
	PassPercent  int              `json:"passPercent"`
	DurationMs   int              `json:"durationMs"`
	TimedOut     bool             `json:"timedOut,omitempty"`
	Results      []QuestionResult `json:"results"`
	// Review: cara klien nampilin pembahasan (diatur admin, lihat config.go).
	Review ReviewSetting `json:"review"`
	// Navigasi setelah lulus: level berikutnya di module yang sama, atau
	// module berikutnya yang baru kebuka (abis final test).
	NextLevelID        string `json:"nextLevelId,omitempty"`
	ExamUnlocked       bool   `json:"examUnlocked,omitempty"`
	NextModuleUnlocked string `json:"nextModuleUnlocked,omitempty"`
}
