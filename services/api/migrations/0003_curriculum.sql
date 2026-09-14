-- Kurikulum & bank soal — FSD.md bagian 4.2/4.3.
-- time_limit_seconds ditambahkan di sesi ini (belum ada di FSD.md):
--   - is_exam = false -> detik PER SOAL (timer reset tiap pindah soal)
--   - is_exam = true  -> total detik buat SELURUH sesi (satu timer, semua soal)

CREATE TYPE question_type_enum AS ENUM ('multiple_choice', 'essay_numeric');
CREATE TYPE question_status_enum AS ENUM ('draft', 'published', 'archived');

CREATE TABLE tiers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code VARCHAR(20) UNIQUE NOT NULL,
    name VARCHAR(100) NOT NULL,
    uses_batch BOOLEAN NOT NULL DEFAULT true,
    order_index INT NOT NULL DEFAULT 0
);

CREATE TABLE batches (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tier_id UUID NOT NULL REFERENCES tiers(id) ON DELETE CASCADE,
    name VARCHAR(150) NOT NULL,
    order_index INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE challenges (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    batch_id UUID NOT NULL REFERENCES batches(id) ON DELETE CASCADE,
    name VARCHAR(150) NOT NULL,
    order_index INT NOT NULL DEFAULT 0,
    is_exam BOOLEAN NOT NULL DEFAULT false,
    question_count_required INT NOT NULL DEFAULT 5,
    pass_threshold_percent INT NOT NULL DEFAULT 70,
    option_count INT NOT NULL DEFAULT 4,
    time_limit_seconds INT NOT NULL DEFAULT 15,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE questions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    challenge_id UUID NOT NULL REFERENCES challenges(id) ON DELETE CASCADE,
    question_type question_type_enum NOT NULL DEFAULT 'multiple_choice',
    prompt TEXT NOT NULL,
    correct_answer_value NUMERIC NOT NULL,
    status question_status_enum NOT NULL DEFAULT 'published',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE question_options (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    question_id UUID NOT NULL REFERENCES questions(id) ON DELETE CASCADE,
    option_value NUMERIC NOT NULL,
    is_correct BOOLEAN NOT NULL DEFAULT false
);

CREATE INDEX idx_batches_tier_id ON batches(tier_id);
CREATE INDEX idx_challenges_batch_id ON challenges(batch_id);
CREATE INDEX idx_questions_challenge_id ON questions(challenge_id);
CREATE INDEX idx_question_options_question_id ON question_options(question_id);
