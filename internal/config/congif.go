package config

import (
	"os"
	"time"
)

type Config struct {
	ServerPort  string
	DatabaseURL string
	MigrateURL  string
	JWTSecret   string
	AccessTTL   time.Duration
	RefreshTTL  time.Duration
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func Load() *Config {
	jwtSecret := getEnv("JWT_SECRET", "")
	if jwtSecret == "" {
		panic("JWT_SECRET os required")
	}
	return &Config{
		ServerPort:  getEnv("SERVER_PORT", "8080"),
		DatabaseURL: getEnv("POSTGRES_URL", "postgres://postgres:secret@localhost:5432/mixfood?sslmode=disable"),
		MigrateURL:  getEnv("MIGRATE_URL", "file://migrations"),
		JWTSecret:   jwtSecret,
		AccessTTL:   time.Minute * 15,
		RefreshTTL:  time.Hour * 24 * 30,
	}
}
