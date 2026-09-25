-- Leaderboard Adventure: tiap attempt nyatet total waktu yang disediain (jumlah
-- time limit semua soal di CP itu), biar waktu yang dipakai bisa dijadiin
-- persentase. Skor waktu = (waktu dipakai - sisa jatah salah x 5 detik) /
-- waktu disediain -- lihat curriculumsvc/adventure_leaderboard.go.
ALTER TABLE adventure_checkpoint_attempts
    ADD COLUMN time_allowed_seconds INT NOT NULL DEFAULT 0;

UPDATE adventure_checkpoint_attempts a
SET time_allowed_seconds = COALESCE((
    SELECT sum((elem->>'timeLimitSeconds')::int)
    FROM jsonb_array_elements(a.questions_snapshot) elem
), 0);

-- Buat query leaderboard (attempt lulus yang masih berlaku).
CREATE INDEX idx_adventure_attempts_valid_passed
    ON adventure_checkpoint_attempts(user_id, checkpoint_no)
    WHERE status = 'passed' AND rolled_back_at IS NULL;
