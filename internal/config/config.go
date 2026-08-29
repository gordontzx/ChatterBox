package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port		string
	DatabaseURL string
	JWTSecret	string
}

func Load() Config {
	_ = godotenv.Load()

	cfg := Config{
		Port:		 getEnv("PORT", "8080"),
		DatabaseURL: mustGetEnv("DATABASE_URL"),
		JWTSecret:	 mustGetEnv("JWT_SECRET"),
	}
	return cfg
}

func getEnv(key string, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func mustGetEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		log.Fatalf("missing required env var: %s", key)
	}
	return v
}
