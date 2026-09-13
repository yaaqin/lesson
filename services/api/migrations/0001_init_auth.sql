-- Autentikasi & user dasar — FSD.md bagian 4.1

CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TYPE user_role AS ENUM ('student', 'admin', 'superadmin');
CREATE TYPE theme_preference_type AS ENUM ('light', 'dark', 'system');
CREATE TYPE auth_provider_type AS ENUM ('google', 'password');

CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) UNIQUE NOT NULL,
    username VARCHAR(255) UNIQUE,
    password_hash VARCHAR(255),
    display_name VARCHAR(255) NOT NULL,
    role user_role NOT NULL DEFAULT 'student',
    theme_preference theme_preference_type NOT NULL DEFAULT 'system',
    locale_preference VARCHAR(10) NOT NULL DEFAULT 'id',
    current_streak INT NOT NULL DEFAULT 0,
    longest_streak INT NOT NULL DEFAULT 0,
    last_active_date DATE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE auth_providers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    provider auth_provider_type NOT NULL,
    provider_uid VARCHAR(255),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (provider, provider_uid)
);
