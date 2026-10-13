-- Pelajaran (subject) di atas tier: sebelumnya semua tier = matematika. Tier
-- baru "coding" (Coding Dasar, JavaScript) punya subject 'coding' -- dipisah
-- biar picker jenjang matematika di userApp, Adventure, & multiplayer gak
-- kecampur soal coding (lihat ListTiers & ListMultiplayerSources).
ALTER TABLE tiers ADD COLUMN subject VARCHAR(30) NOT NULL DEFAULT 'matematika';

-- Opsi PG berupa teks (mis. potongan kode / output console.log) buat soal
-- coding. option_value tetep dipakai sebagai identitas opsi yang dikirim balik
-- pas submit (diisi indeks 0..n-1), option_label yang ditampilin ke user.
-- NULL = opsi angka biasa, klien nampilin option_value kayak sebelumnya.
ALTER TABLE question_options ADD COLUMN option_label TEXT;
