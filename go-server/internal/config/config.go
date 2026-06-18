package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port             string
	DatabaseURL      string
	BaseURL          string
	DefaultEncoding  string
}

func Load() *Config {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	return &Config{
		Port:            os.Getenv("PORT"),
		DatabaseURL:     os.Getenv("DATABASE_URL"),
		BaseURL:         os.Getenv("BASE_URL"),
		DefaultEncoding: os.Getenv("DEFAULT_ENCODING"),
	}
}