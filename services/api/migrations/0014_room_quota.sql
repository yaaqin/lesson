-- Kuota bikin room multiplayer. User biasa punya jatah (room_quota) yang
-- berkurang 1 tiap bikin room; user premium (diset admin) gak kepotong.
-- Jatah awal user baru diambil dari multiplayer_config.initial_room_quota
-- pas daftar (authsvc/google.go), jadi ngubah setelan cuma ngaruh ke user
-- yang daftar setelahnya. User bisa minta tambahan jatah ke admin lewat
-- room_quota_requests (1 permintaan pending per user).

ALTER TABLE multiplayer_config ADD COLUMN initial_room_quota INT NOT NULL DEFAULT 3;

ALTER TABLE users ADD COLUMN room_quota INT NOT NULL DEFAULT 0 CHECK (room_quota >= 0);
-- Murid yang udah terdaftar dapet jatah awal yang sama kayak user baru.
UPDATE users SET room_quota = 3 WHERE role = 'student';

CREATE TABLE room_quota_requests (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    message TEXT NOT NULL,
    status VARCHAR(16) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'approved', 'rejected')),
    granted_quota INT,
    decided_by UUID REFERENCES users(id) ON DELETE SET NULL,
    decided_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE UNIQUE INDEX room_quota_requests_one_pending ON room_quota_requests (user_id) WHERE status = 'pending';
CREATE INDEX room_quota_requests_status_created ON room_quota_requests (status, created_at);
