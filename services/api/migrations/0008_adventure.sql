-- Adventure: mode "jalan terus" lintas jenjang (SD -> Kampus), soalnya diambil
-- dari bank soal kurikulum yang udah ada. 4000 soal = 40 checkpoint x 100 soal.
-- CP 1-10 urut per blok jenjang, CP 11+ gacha dengan komposisi yang sama per
-- CP (lihat curriculumsvc/adventure.go). Tiap CP punya jatah salah 5x -- salah
-- ke-6 = CP gagal, ulang dari soal pertama CP itu.

-- Posisi user sekarang. current_checkpoint = CP yang lagi/akan dikerjain
-- (1-based); lewat dari total CP = adventure tamat.
CREATE TABLE adventure_progress (
    user_id UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    current_checkpoint INT NOT NULL DEFAULT 1,
    completed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TYPE adventure_attempt_status_enum AS ENUM ('in_progress', 'passed', 'failed', 'abandoned');

-- Satu baris = satu kali main 1 checkpoint. Jawaban dicek per soal di server
-- (kunci jawaban cuma ada di questions_snapshot, gak pernah dikirim ke klien).
-- Keluar di tengah CP -> attempt-nya jadi 'abandoned', lanjutnya dari awal CP.
-- rolled_back_at keisi kalau user balik ke CP sebelumnya: attempt 'passed' di
-- CP >= target tetap disimpen buat history tapi gak dianggap progres lagi.
CREATE TABLE adventure_checkpoint_attempts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    checkpoint_no INT NOT NULL,
    status adventure_attempt_status_enum NOT NULL DEFAULT 'in_progress',
    questions_snapshot JSONB NOT NULL,
    current_index INT NOT NULL DEFAULT 0,
    correct_count INT NOT NULL DEFAULT 0,
    fails_used INT NOT NULL DEFAULT 0,
    max_fails INT NOT NULL,
    question_started_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    started_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    completed_at TIMESTAMPTZ,
    rolled_back_at TIMESTAMPTZ
);

CREATE INDEX idx_adventure_attempts_user_checkpoint
    ON adventure_checkpoint_attempts(user_id, checkpoint_no);

-- Maksimal 1 attempt in_progress per user.
CREATE UNIQUE INDEX uq_adventure_attempts_one_in_progress
    ON adventure_checkpoint_attempts(user_id) WHERE status = 'in_progress';
