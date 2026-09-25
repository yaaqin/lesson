-- Multiplayer: pengaturan jeda antar soal (diatur admin platform dari
-- dashboard, dibaca pas room dibikin -- room yang lagi jalan gak ikut berubah)
-- + hasil pertandingan minimal buat kartu share (siapa aja yang main & rank-nya,
-- tanpa skor). Lihat internal/multiplayer.

-- Satu baris aja (id selalu true).
CREATE TABLE multiplayer_config (
    id BOOLEAN PRIMARY KEY DEFAULT true CHECK (id),
    classic_cooldown_seconds INT NOT NULL DEFAULT 3,
    race_cooldown_seconds INT NOT NULL DEFAULT 3,
    race_winner_reveal_seconds INT NOT NULL DEFAULT 1,
    results_countdown_seconds INT NOT NULL DEFAULT 5,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
INSERT INTO multiplayer_config DEFAULT VALUES;

CREATE TABLE multiplayer_matches (
    id UUID PRIMARY KEY,
    mode VARCHAR(16) NOT NULL,
    question_count INT NOT NULL,
    finished_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE multiplayer_match_players (
    match_id UUID NOT NULL REFERENCES multiplayer_matches(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    rank INT NOT NULL,
    PRIMARY KEY (match_id, user_id)
);
