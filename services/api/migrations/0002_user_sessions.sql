-- Sesi refresh token. user_id sebagai PRIMARY KEY memastikan maksimal 1 baris
-- per user, sehingga login baru selalu menggantikan (invalidate) sesi lama —
-- satu user tidak bisa punya lebih dari satu refresh token aktif sekaligus.

CREATE TABLE user_sessions (
    user_id UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    refresh_token_hash VARCHAR(64) NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
