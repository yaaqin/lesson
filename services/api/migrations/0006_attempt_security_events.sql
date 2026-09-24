-- Anti-cheating: log sinyal mencurigakan per attempt (tab di-switch, window
-- kehilangan fokus, keluar fullscreen, nyoba copy soal). Dikirim batch dari
-- apps/web, dievaluasi jadi risk score pas attempt disubmit. Durasi jawab
-- sengaja GAK dijadiin sinyal (false positive buat user yang ngerjain tier di
-- bawah kemampuannya).

CREATE TABLE attempt_security_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    attempt_id UUID NOT NULL REFERENCES challenge_attempts(id) ON DELETE CASCADE,
    event_type TEXT NOT NULL CHECK (event_type IN ('tab_hidden', 'window_blur', 'fullscreen_exit', 'copy_blocked')),
    question_index INT,
    started_at TIMESTAMPTZ NOT NULL,
    duration_ms INT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_attempt_security_events_attempt_id ON attempt_security_events(attempt_id);

-- Cache hasil scoring pas submit (biar dashboard/leaderboard gak perlu
-- aggregate ulang). risk_level: normal | low_confidence | review.
ALTER TABLE challenge_attempts
    ADD COLUMN risk_score INT,
    ADD COLUMN risk_level TEXT;
