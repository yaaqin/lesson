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

// Seed akun admin platform pertama. Idempotent — aman dijalankan berkali-kali
// (ON CONFLICT (email) akan update display_name/password_hash-nya).
func main() {
	cfg := config.Load()
	ctx := context.Background()

	email := getEnv("SEED_ADMIN_EMAIL", "admin@mathquest.dev")
	password := getEnv("SEED_ADMIN_PASSWORD", "admin12345")
	displayName := getEnv("SEED_ADMIN_NAME", "Admin Platform")

	pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("connect db: %v", err)
	}
	defer pool.Close()

	passwordHash, err := authsvc.HashPassword(password)
	if err != nil {
		log.Fatalf("hash password: %v", err)
	}

	var userID string
	err = pool.QueryRow(ctx, `
		INSERT INTO users (email, display_name, role, password_hash)
		VALUES ($1, $2, 'superadmin', $3)
		ON CONFLICT (email) DO UPDATE
		SET display_name = EXCLUDED.display_name,
		    password_hash = EXCLUDED.password_hash,
		    role = 'superadmin',
		    updated_at = now()
		RETURNING id
	`, email, displayName, passwordHash).Scan(&userID)
	if err != nil {
		log.Fatalf("upsert admin user: %v", err)
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
		log.Fatalf("check auth_provider: %v", err)
	}
	if !authProviderExists {
		if _, err := pool.Exec(ctx, `
			INSERT INTO auth_providers (user_id, provider, provider_uid)
			VALUES ($1, 'password', NULL)
		`, userID); err != nil {
			log.Fatalf("insert auth_provider: %v", err)
		}
	}

	fmt.Println("Admin seeded:")
	fmt.Printf("  email:    %s\n", email)
	fmt.Printf("  password: %s\n", password)
	fmt.Println("(ganti password ini setelah login pertama kali)")
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
