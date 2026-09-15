-- Kategori: dipakai tier yang gak pake batch (mis. "umum") sebagai ganti batch
-- buat ngelompokin challenge (FSD.md 3.3/4.3 -- sebelumnya cuma ada route stub
-- notImplemented buat ini). Sebuah challenge kudu punya PERSIS salah satu dari
-- batch_id / category_id, gak boleh dua-duanya atau gak ada sama sekali.
CREATE TABLE categories (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tier_id UUID NOT NULL REFERENCES tiers(id) ON DELETE CASCADE,
    code VARCHAR(50) NOT NULL,
    name VARCHAR(150) NOT NULL,
    order_index INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (tier_id, code)
);

CREATE INDEX idx_categories_tier_id ON categories(tier_id);

ALTER TABLE challenges
    ALTER COLUMN batch_id DROP NOT NULL,
    ADD COLUMN category_id UUID REFERENCES categories(id) ON DELETE CASCADE,
    ADD CONSTRAINT chk_challenges_batch_xor_category CHECK (
        (batch_id IS NOT NULL AND category_id IS NULL) OR
        (batch_id IS NULL AND category_id IS NOT NULL)
    );

CREATE INDEX idx_challenges_category_id ON challenges(category_id);

-- Tipe soal baru: grid_puzzle -- soal berupa satu puzzle utuh (kotak isian
-- angka, cryptarithm, dst), bukan satu jawaban skalar kayak MC/essay. Struktur
-- tiap "kind" beda-beda makanya disimpen di tabel terpisah puzzle_questions
-- (JSONB payload), bukan dipaksain ke question_options. correct_answer_value
-- jadi nullable karena grid_puzzle gak punya satu correct_answer_value.
ALTER TYPE question_type_enum ADD VALUE 'grid_puzzle';
ALTER TABLE questions ALTER COLUMN correct_answer_value DROP NOT NULL;

-- payload terstruktur per "kind" (addition_grid, cryptarithm, dst nanti bisa
-- nambah lagi tanpa migration baru -- kind cuma nentuin gimana klien nge-render
-- & bagian mana yang harus diisi user). Solusi lengkap CUMA ada di sini
-- (server-side) & di questions_snapshot punya attempt -- gak pernah dikirim ke
-- klien lewat StartChallenge (beda dari question_options.is_correct yang emang
-- udah dikirim ke klien sebagai simplifikasi yang disengaja).
CREATE TABLE puzzle_questions (
    question_id UUID PRIMARY KEY REFERENCES questions(id) ON DELETE CASCADE,
    kind VARCHAR(50) NOT NULL,
    grid_size INT,
    payload JSONB NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
