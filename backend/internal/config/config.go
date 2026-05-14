package config

import (
	"fmt"
	"os"
)

// Config holds all runtime configuration for the application.
// Populated from environment variables at startup.
type Config struct {
	// Database
	DatabaseURL string

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

	// Backend self-URL (for webhook callback construction, success URLs, etc.)
	BackendURL string
}

// Load reads environment variables and returns a populated Config.
// Returns an error if any required variable is missing or malformed.
func Load() (*Config, error) {
	cfg := &Config{
		DatabaseURL:              getenv("DATABASE_URL", ""),
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
