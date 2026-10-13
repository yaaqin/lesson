-- Course Coding Dasar: mode belajar materi, kepisah total dari kuis (tiers /
-- challenges / lives). Gak ada nyawa di sini, boleh ngulang sepuasnya.
--
-- Struktur: module (Tipe Variabel -> Looping -> Asynchronous, berurutan) ->
-- level (materi + latihan, lulus = 100% bener) -> final test per module
-- (speedrun: soal diacak dari bank seluruh module, ada batas waktu total).
-- Module berikutnya kebuka kalau final test module sebelumnya lulus; level
-- berikutnya kebuka kalau level sebelumnya lulus; final test kebuka kalau
-- semua level di module itu lulus (lihat internal/coursesvc/unlock.go).

CREATE TABLE course_modules (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code VARCHAR(50) UNIQUE NOT NULL,
    title VARCHAR(150) NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    order_index INT NOT NULL DEFAULT 0,
    exam_question_count INT NOT NULL DEFAULT 15,
    exam_time_limit_seconds INT NOT NULL DEFAULT 300,
    exam_pass_percent INT NOT NULL DEFAULT 80,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- material: array blok konten (heading / paragraph / code / example / tip /
-- table), dirender apa adanya sama apps/web -- bentuk tiap blok lihat
-- MaterialBlock di internal/coursesvc/types.go.
CREATE TABLE course_levels (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    module_id UUID NOT NULL REFERENCES course_modules(id) ON DELETE CASCADE,
    code VARCHAR(50) NOT NULL,
    title VARCHAR(150) NOT NULL,
    summary TEXT NOT NULL DEFAULT '',
    order_index INT NOT NULL DEFAULT 0,
    material JSONB NOT NULL DEFAULT '[]',
    -- soal yang diambil acak dari bank level ini tiap latihan
    question_count INT NOT NULL DEFAULT 5,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (module_id, code)
);

CREATE INDEX idx_course_levels_module_id ON course_levels(module_id);

-- kind:
--   multiple_choice -> payload {"options": [...], "correct": <indeks>}
--   type_answer     -> user ngetik nama tipe; payload {"accepted": ["string", ...]}
--   write_value     -> user nulis nilai/data JS sendiri, dicek pakai skema
--                      (bukan jawaban pasti); payload {"schema": {...},
--                      "strictJson": bool, "placeholder": "...", "sample": "..."}
-- level_id NULL = soal khusus final test. in_exam = soal level yang ikut
-- masuk bank final test module-nya.
CREATE TABLE course_questions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    module_id UUID NOT NULL REFERENCES course_modules(id) ON DELETE CASCADE,
    level_id UUID REFERENCES course_levels(id) ON DELETE CASCADE,
    kind VARCHAR(30) NOT NULL CHECK (kind IN ('multiple_choice', 'type_answer', 'write_value')),
    prompt TEXT NOT NULL,
    code TEXT,
    payload JSONB NOT NULL,
    explanation TEXT NOT NULL DEFAULT '',
    in_exam BOOLEAN NOT NULL DEFAULT true,
    order_index INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_course_questions_module_id ON course_questions(module_id);
CREATE INDEX idx_course_questions_level_id ON course_questions(level_id);

-- level_id NULL = attempt final test. questions_snapshot nyimpen soal + kunci
-- persis kayak waktu dimulai (bank bisa di-seed ulang tanpa ngerusak attempt
-- yang lagi jalan). duration_ms = waktu speedrun (dihitung server).
CREATE TABLE course_attempts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    module_id UUID NOT NULL REFERENCES course_modules(id) ON DELETE CASCADE,
    level_id UUID REFERENCES course_levels(id) ON DELETE CASCADE,
    questions_snapshot JSONB NOT NULL,
    answers JSONB,
    correct_count INT,
    total_questions INT NOT NULL,
    score_percent INT,
    passed BOOLEAN,
    duration_ms INT,
    started_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    completed_at TIMESTAMPTZ
);

CREATE INDEX idx_course_attempts_user_module ON course_attempts(user_id, module_id);
