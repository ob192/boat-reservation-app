package config

import (
	"fmt"
	"os"
	"strings"
)

// Config holds all runtime configuration for the application.
// Populated from environment variables at startup.
type Config struct {
	// Database
	DatabaseURL string

	// DirectDatabaseURL is a non-pooled connection used only for AutoMigrate.
	// Neon's pooled ("-pooler") endpoint can hand a schema-changing statement and a
	// later query to different physical connections sharing a stale server-side
	// cached plan, which Postgres rejects with "cached plan must not change result
	// type" (SQLSTATE 0A000) once a migration adds/changes columns mid-session.
	// Defaults to DatabaseURL with "-pooler" stripped from the host; override via
	// DIRECT_DATABASE_URL if the DB isn't Neon or uses a different naming scheme.
	DirectDatabaseURL string

	// Supabase
	SupabaseURL              string
	SupabaseProjectReference string
	SupabaseAnonKey          string
	SupabaseServiceRoleKey   string

	// Frontend / CORS
	FrontendURL string

	// Email
	EmailFrom string

	// Payment gateway
	PaymentGateway   string // "stub", "liqpay", etc.
	LiqpayPublicKey  string
	LiqpayPrivateKey string

	// Server
	Port    string
	GinMode string

	// Poster CRM
	PosterAPIToken string
	PosterEnabled  bool

	// Backend self-URL (for webhook callback construction, success URLs, etc.)
	BackendURL string
}

// Load reads environment variables and returns a populated Config.
// Returns an error if any required variable is missing or malformed.
func Load() (*Config, error) {
	databaseURL := getenv("DATABASE_URL", "")
	cfg := &Config{
		DatabaseURL:              databaseURL,
		DirectDatabaseURL:        getenv("DIRECT_DATABASE_URL", deriveDirectDatabaseURL(databaseURL)),
		SupabaseURL:              getenv("SUPABASE_URL", ""),
		SupabaseProjectReference: getenv("SUPABASE_PROJECT_REFERENCE", ""),
		SupabaseAnonKey:          getenv("SUPABASE_ANON_KEY", ""),
		SupabaseServiceRoleKey:   getenv("SUPABASE_SERVICE_ROLE_KEY", ""),
		FrontendURL:              getenv("FRONTEND_URL", "http://localhost:3000"),
		EmailFrom:                getenv("EMAIL_FROM", ""),
		PaymentGateway:           getenv("PAYMENT_GATEWAY", "liqpay"),
		LiqpayPublicKey:          getenv("LIQPAY_PUBLIC_KEY", "sandbox_i47824146433"),
		LiqpayPrivateKey:         getenv("LIQPAY_PRIVATE_KEY", "sandbox_1DUnv7mMzCsASTV52Bf9mf9DOg2ogVNB2dkhT7Ox"),
		Port:                     getenv("PORT", "8080"),
		GinMode:                  getenv("GIN_MODE", "debug"),
		BackendURL:               getenv("BACKEND_URL", "http://localhost:8080"),
		PosterAPIToken:           getenv("POSTER_API_TOKEN", ""),
		PosterEnabled:            getenv("POSTER_API_TOKEN", "") != "",
	}

	if err := cfg.validate(); err != nil {
		return nil, err
	}
	return cfg, nil
}

func (c *Config) validate() error {
	required := map[string]string{
		"DATABASE_URL":               c.DatabaseURL,
		"SUPABASE_URL":               c.SupabaseURL,
		"SUPABASE_PROJECT_REFERENCE": c.SupabaseProjectReference,
		"SUPABASE_ANON_KEY":          c.SupabaseAnonKey,
		"SUPABASE_SERVICE_ROLE_KEY":  c.SupabaseServiceRoleKey,
	}
	for k, v := range required {
		if v == "" {
			return fmt.Errorf("missing required environment variable: %s", k)
		}
	}
	return nil
}

func getenv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return fallback
}

// deriveDirectDatabaseURL strips Neon's "-pooler" suffix from the host portion of
// dsn, giving the direct (non-PgBouncer) endpoint. If dsn isn't a Neon pooled URL,
// it's returned unchanged.
func deriveDirectDatabaseURL(dsn string) string {
	return strings.Replace(dsn, "-pooler.", ".", 1)
}
