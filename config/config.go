package config

import (
	"os"
)

type Config struct {
	ServerPort string
	JWTSecret string
}

func Load() *Config {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "goquizvibe-secret-key-change-in-production"
	}

	return &Config{
		ServerPort: port,
		JWTSecret:  secret,
	}
}
