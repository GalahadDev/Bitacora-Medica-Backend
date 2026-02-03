package config

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DBUrl        string
	SupabaseURL  string
	SupabaseKey  string
	JwtSecret    string
	Port         string
	SMTPHost     string
	SMTPPort     string
	SMTPEmail    string
	SMTPPassword string
}

func LoadConfig() *Config {
	if err := godotenv.Load(); err != nil {
		slog.Warn("No .env file found or failed to load, relying on system env vars")
	}

	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=require",
		getEnv("DB_HOST", ""),
		getEnv("DB_USER", ""),
		getEnv("DB_PASSWORD", ""),
		getEnv("DB_NAME", "postgres"),
		getEnv("DB_PORT", "5432"),
	)

	cfg := &Config{
		DBUrl:        dsn,
		SupabaseURL:  getEnv("SUPABASE_URL", ""),
		SupabaseKey:  getEnv("SUPABASE_SERVICE_ROLE_KEY", ""),
		JwtSecret:    getEnv("JWT_SECRET", ""),
		Port:         getEnv("PORT", "8080"),
		SMTPHost:     getEnv("SMTP_HOST", "smtp.gmail.com"),
		SMTPPort:     getEnv("SMTP_PORT", "587"),
		SMTPEmail:    getEnv("SMTP_EMAIL", ""),
		SMTPPassword: getEnv("SMTP_PASSWORD", ""),
	}

	if cfg.JwtSecret == "" {
		slog.Warn("JWT_SECRET is missing. Auth verification might fail if not using JWKS.")
	}

	return cfg
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
