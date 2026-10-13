-- Pengaturan pembahasan Course (diatur admin dari dashboard), dipisah buat
-- latihan level & final test. review_mode:
--   full        -> jawabanmu, kunci, & penjelasan tampil terus
--   timed       -> tampil review_seconds detik, abis itu disembunyiin (sisa
--                  benar/salah per soal)
--   result_only -> cuma benar/salah per soal; kunci & penjelasan gak dikirim
--                  server sama sekali (lihat coursesvc.Submit)
-- Satu baris aja (id selalu true).
CREATE TABLE course_config (
    id BOOLEAN PRIMARY KEY DEFAULT true CHECK (id),
    level_review_mode VARCHAR(20) NOT NULL DEFAULT 'timed'
        CHECK (level_review_mode IN ('full', 'timed', 'result_only')),
    level_review_seconds INT NOT NULL DEFAULT 5,
    exam_review_mode VARCHAR(20) NOT NULL DEFAULT 'result_only'
        CHECK (exam_review_mode IN ('full', 'timed', 'result_only')),
    exam_review_seconds INT NOT NULL DEFAULT 5,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
INSERT INTO course_config DEFAULT VALUES;
