-- Leaderboard per jenjang (SD/SMP/SMK/Kampus): poin = jumlah jawaban benar.
-- Sebelumnya attempt cuma nyimpen score_percent, jadi jumlah benarnya
-- dicatat juga -- lihat curriculumsvc/tier_leaderboard.go.
ALTER TABLE challenge_attempts ADD COLUMN correct_count INT;

-- Backfill attempt yang udah selesai: score_percent = round(benar / total x
-- 100) dan total soal <= 100, jadi jumlah benarnya bisa dihitung balik persis.
UPDATE challenge_attempts
SET correct_count = round(score_percent * jsonb_array_length(questions_snapshot) / 100.0)::int
WHERE status IN ('passed', 'failed') AND score_percent IS NOT NULL;

CREATE INDEX idx_challenge_attempts_leaderboard
    ON challenge_attempts(challenge_id, user_id)
    WHERE status IN ('passed', 'failed');
