package config

import (
	"os"
	"strconv"
	"time"
)

// Config is loaded from environment variables. Every value has a working
// development default so `go run ./cmd/web` just works.
type Config struct {
	AppName string
	Env     string // "development" or "production"
	Addr    string

	DBDriver string // "sqlite" or "postgres"
	DBDSN    string

	SessionTTL    time.Duration
	SecureCookies bool

	SeedAdminEmail    string
	SeedAdminPassword string
}

func (c *Config) IsDev() bool { return c.Env != "production" }

func Load() Config {
	return Config{
		AppName: env("BARA_APP_NAME", "Bara"),
		Env:     env("BARA_ENV", "development"),
		Addr:    env("BARA_ADDR", ":8080"),

		DBDriver: env("BARA_DB_DRIVER", "sqlite"),
		DBDSN:    env("BARA_DB_DSN", "data/bara.db"),

		SessionTTL:    time.Duration(envInt("BARA_SESSION_TTL_HOURS", 24*7)) * time.Hour,
		SecureCookies: env("BARA_COOKIE_SECURE", "false") == "true",

		SeedAdminEmail:    env("BARA_SEED_ADMIN_EMAIL", "admin@bara.dev"),
		SeedAdminPassword: env("BARA_SEED_ADMIN_PASSWORD", "Admin1234"),
	}
}

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func envInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return fallback
}
