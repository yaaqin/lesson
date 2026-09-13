package main

import (
	"context"
	"log"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"

	"lesson/api/internal/authsvc"
	"lesson/api/internal/config"
	"lesson/api/internal/httpserver"
)

func main() {
	cfg := config.Load()

	db, err := pgxpool.New(context.Background(), cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("failed to create db pool: %v", err)
	}
	defer db.Close()

	auth := authsvc.NewService(db, cfg.JWTSecret, cfg.AccessTokenTTL, cfg.RefreshTokenTTL)
	srv := httpserver.New(db, auth, cfg.AllowedOrigins)

	log.Printf("api listening on :%s", cfg.Port)
	if err := http.ListenAndServe(":"+cfg.Port, srv.Handler()); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
