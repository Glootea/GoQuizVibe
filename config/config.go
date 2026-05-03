package config

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
)

type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
}

func (d *DatabaseConfig) DSN() string {
	return fmt.Sprintf(
		"host=127.0.0.1 user=%s password=%s dbname=%s port=%s sslmode=disable",
		d.User, d.Password, d.DBName, d.Port,
	)
}

type MinioConfig struct {
	Endpoint  string
	AccessKey string
	SecretKey string
	Bucket    string
}

type Config struct {
	ServerPort string
	JWTSecret  string
	Database   DatabaseConfig
	Minio      MinioConfig
}

func Load() *Config {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "goquizvibe-secret-key-change-in-production"
	}

	dbHost := os.Getenv("DB_HOST")
	if dbHost == "" {
		dbHost = "localhost"
	}

	dbPort := os.Getenv("DB_PORT")
	if dbPort == "" {
		dbPort = "5432"
	}

	dbUser := os.Getenv("POSTGRES_USER")
	if dbUser == "" {
		panic("Failed to get env for db")
	}

	dbPassword := os.Getenv("POSTGRES_PASSWORD")
	if dbPassword == "" {
		panic("Failed to get env for db")
	}

	dbName := os.Getenv("POSTGRES_DB")
	if dbName == "" {
		panic("Failed to get env for db")
	}

	return &Config{
		ServerPort: port,
		JWTSecret:  secret,
		Database: DatabaseConfig{
			Host:     dbHost,
			Port:     dbPort,
			User:     dbUser,
			Password: dbPassword,
			DBName:   dbName,
		},
		Minio: MinioConfig{
			Endpoint:  os.Getenv("MINIO_ENDPOINT"),
			AccessKey: os.Getenv("MINIO_ROOT_USER"),
			SecretKey: os.Getenv("MINIO_ROOT_PASSWORD"),
			Bucket:    os.Getenv("MINIO_BUCKET"),
		},
	}
}
