package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"

	"lesson/api/internal/authsvc"
	"lesson/api/internal/config"
)

// Seed akun dasar: admin platform + 1 murid demo (buat coba gameplay tanpa
// perlu bikin akun manual). Idempotent — aman dijalankan berkali-kali.
func main() {
	cfg := config.Load()
	ctx := context.Background()

	pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("connect db: %v", err)
	}
	defer pool.Close()

	seedAccount(ctx, pool,
		getEnv("SEED_ADMIN_EMAIL", "admin@mathquest.dev"),
		getEnv("SEED_ADMIN_PASSWORD", "admin12345"),
		getEnv("SEED_ADMIN_NAME", "Admin Platform"),
		"superadmin",
	)

	seedAccount(ctx, pool,
		getEnv("SEED_STUDENT_EMAIL", "siswa@mathquest.dev"),
		getEnv("SEED_STUDENT_PASSWORD", "siswa12345"),
		getEnv("SEED_STUDENT_NAME", "Siswa Demo"),
		"student",
	)
}

func seedAccount(ctx context.Context, pool *pgxpool.Pool, email, password, displayName, role string) {
	passwordHash, err := authsvc.HashPassword(password)
	if err != nil {
		log.Fatalf("hash password (%s): %v", email, err)
	}

	var userID string
	err = pool.QueryRow(ctx, `
		INSERT INTO users (email, display_name, role, password_hash)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (email) DO UPDATE
		SET display_name = EXCLUDED.display_name,
		    password_hash = EXCLUDED.password_hash,
		    role = EXCLUDED.role,
		    updated_at = now()
		RETURNING id
	`, email, displayName, role, passwordHash).Scan(&userID)
	if err != nil {
		log.Fatalf("upsert user (%s): %v", email, err)
	}

	// provider_uid NULL bikin UNIQUE (provider, provider_uid) gak efektif buat ON CONFLICT
	// (NULL selalu dianggap beda di Postgres), jadi dicek manual dulu biar idempotent.
	var authProviderExists bool
	err = pool.QueryRow(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM auth_providers WHERE user_id = $1 AND provider = 'password'
		)
	`, userID).Scan(&authProviderExists)
	if err != nil {
		log.Fatalf("check auth_provider (%s): %v", email, err)
	}
	if !authProviderExists {
		if _, err := pool.Exec(ctx, `
			INSERT INTO auth_providers (user_id, provider, provider_uid)
			VALUES ($1, 'password', NULL)
		`, userID); err != nil {
			log.Fatalf("insert auth_provider (%s): %v", email, err)
		}
	}

	fmt.Printf("%s seeded: %s / %s\n", role, email, password)
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
