package config

import (
	"os"
	"strings"
	"time"
)

type Config struct {
	Port               string
	DatabaseURL        string
	JWTSecret          string
	AccessTokenTTL     time.Duration
	RefreshTokenTTL    time.Duration
	AllowedOrigins     []string
	GoogleClientID     string
	GoogleClientSecret string
	GoogleRedirectURL  string
	WebAppURL          string
}

func Load() Config {
	return Config{
		Port:               getEnv("PORT", "9801"),
		DatabaseURL:        getEnv("DATABASE_URL", "postgres://lesson:lesson@localhost:9800/lesson?sslmode=disable"),
		JWTSecret:          getEnv("JWT_SECRET", "dev-secret-change-me"),
		AccessTokenTTL:     getEnvDuration("ACCESS_TOKEN_TTL", 15*time.Minute),
		RefreshTokenTTL:    getEnvDuration("REFRESH_TOKEN_TTL", 7*24*time.Hour),
		AllowedOrigins:     getEnvList("ALLOWED_ORIGINS", []string{"http://localhost:9802", "http://localhost:9803"}),
		GoogleClientID:     getEnv("GOOGLE_CLIENT_ID", ""),
		GoogleClientSecret: getEnv("GOOGLE_CLIENT_SECRET", ""),
		GoogleRedirectURL:  getEnv("GOOGLE_REDIRECT_URL", "http://localhost:9801/api/v1/auth/google/callback"),
		WebAppURL:          getEnv("WEB_APP_URL", "http://localhost:9803"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvDuration(key string, fallback time.Duration) time.Duration {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	d, err := time.ParseDuration(v)
	if err != nil {
		return fallback
	}
	return d
}

func getEnvList(key string, fallback []string) []string {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	parts := strings.Split(v, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if trimmed := strings.TrimSpace(p); trimmed != "" {
			out = append(out, trimmed)
		}
	}
	return out
}
