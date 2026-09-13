package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"

	"github.com/jackc/pgx/v5/pgxpool"

	"lesson/api/internal/config"
)

// Migration runner sederhana: jalanin file .sql di migrations/ secara urut,
// dicatat di tabel schema_migrations biar idempotent (aman dijalanin berkali-kali).
func main() {
	cfg := config.Load()
	ctx := context.Background()

	pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("connect db: %v", err)
	}
	defer pool.Close()

	if _, err := pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version TEXT PRIMARY KEY,
			applied_at TIMESTAMPTZ NOT NULL DEFAULT now()
		)
	`); err != nil {
		log.Fatalf("create schema_migrations: %v", err)
	}

	files, err := filepath.Glob("migrations/*.sql")
	if err != nil {
		log.Fatalf("glob migrations: %v", err)
	}
	sort.Strings(files)

	applied := 0
	for _, file := range files {
		version := filepath.Base(file)

		var alreadyApplied bool
		if err := pool.QueryRow(ctx,
			`SELECT EXISTS(SELECT 1 FROM schema_migrations WHERE version = $1)`, version,
		).Scan(&alreadyApplied); err != nil {
			log.Fatalf("check %s: %v", version, err)
		}
		if alreadyApplied {
			continue
		}

		sqlBytes, err := os.ReadFile(file)
		if err != nil {
			log.Fatalf("read %s: %v", file, err)
		}

		tx, err := pool.Begin(ctx)
		if err != nil {
			log.Fatalf("begin tx for %s: %v", version, err)
		}
		if _, err := tx.Exec(ctx, string(sqlBytes)); err != nil {
			tx.Rollback(ctx)
			log.Fatalf("apply %s: %v", version, err)
		}
		if _, err := tx.Exec(ctx, `INSERT INTO schema_migrations (version) VALUES ($1)`, version); err != nil {
			tx.Rollback(ctx)
			log.Fatalf("record %s: %v", version, err)
		}
		if err := tx.Commit(ctx); err != nil {
			log.Fatalf("commit %s: %v", version, err)
		}

		fmt.Printf("applied %s\n", version)
		applied++
	}

	if applied == 0 {
		fmt.Println("no pending migrations, database already up to date")
	}
}
