-- Profil user: nickname + avatar.
--
-- Nickname = kolom users.username yang udah ada (UNIQUE, jadi sekaligus key
-- kedua selain email -- login password juga bisa pakai ini). Format: 6-20
-- karakter, cuma a-z, 0-9, "_" dan ".". Nullable: user baru (mis. login Google)
-- belum punya, apps/web ngarahin ke onboarding buat bikin dulu.
ALTER TABLE users
    ADD CONSTRAINT chk_users_username_format
    CHECK (username IS NULL OR username ~ '^[a-z0-9_.]{6,20}$');

-- avatar_type 'character' -> avatar_key = id karakter kartun (daftar ada di
-- curriculumsvc/profile.go & apps/web lib/avatars.ts). 'google' -> pakai foto
-- akun Google (google_avatar_url, di-update tiap login Google).
ALTER TABLE users
    ADD COLUMN avatar_type TEXT NOT NULL DEFAULT 'character'
        CHECK (avatar_type IN ('character', 'google')),
    ADD COLUMN avatar_key TEXT NOT NULL DEFAULT 'fox',
    ADD COLUMN google_avatar_url TEXT;
