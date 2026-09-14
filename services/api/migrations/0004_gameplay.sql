-- Attempt & nyawa — FSD.md bagian 4.4 (weekly cap 15/minggu disederhanakan:
-- reset harian ke 3 tiap hari, belum implementasi cap mingguan).

CREATE TYPE attempt_status_enum AS ENUM ('in_progress', 'passed', 'failed');

CREATE TABLE challenge_attempts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    challenge_id UUID NOT NULL REFERENCES challenges(id) ON DELETE CASCADE,
    status attempt_status_enum NOT NULL DEFAULT 'in_progress',
    score_percent INT,
    questions_snapshot JSONB NOT NULL,
    answers_submitted JSONB,
    started_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    completed_at TIMESTAMPTZ
);

CREATE TABLE lives (
    user_id UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    lives_remaining INT NOT NULL DEFAULT 3,
    last_daily_reset_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_challenge_attempts_user_id ON challenge_attempts(user_id);
CREATE INDEX idx_challenge_attempts_challenge_id ON challenge_attempts(challenge_id);
