package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port         string
	DatabaseURL  string
	JWTSecret    string
	JWTExpiresIn string
	FrontendURL  string
	Env          string
}

func Load() (*Config, error) {
	// Загружаем .env (не падаем, если файла нет — на проде переменные из окружения)
	_ = godotenv.Load()

	cfg := &Config{
		Port:         getEnv("PORT", "4000"),
		DatabaseURL:  getEnv("DATABASE_URL", ""),
		JWTSecret:    getEnv("JWT_SECRET", ""),
		JWTExpiresIn: getEnv("JWT_EXPIRES_IN", "168h"),
		FrontendURL:  getEnv("FRONTEND_URL", "http://localhost:3000"),
		Env:          getEnv("NODE_ENV", "development"),
	}

	if cfg.DatabaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL не задан")
	}
	if cfg.JWTSecret == "" {
		return nil, fmt.Errorf("JWT_SECRET не задан")
	}

	return cfg, nil
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
