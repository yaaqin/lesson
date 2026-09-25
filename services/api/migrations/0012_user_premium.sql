-- Premium: penanda user yang boleh bikin room multiplayer (lihat
-- internal/multiplayer). Diset manual oleh admin platform dari dashboard
-- (PUT /admin/users/{id}/premium) -- belum ada alur pembayaran.
ALTER TABLE users ADD COLUMN is_premium BOOLEAN NOT NULL DEFAULT false;
