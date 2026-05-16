package config

import (
	"fmt"
	"log"
	"os"
	"time"

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

type RedisConfig struct {
	Password          string
	CacheTTL          time.Duration
	TimerCronInterval time.Duration
}

type Config struct {
	ServerPort string
	JWTSecret  string
	Database   DatabaseConfig
	Minio      MinioConfig
	Redis      RedisConfig
}

func Load() *Config {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	port := os.Getenv("PORT")
	if port == "" {
		port = "7890"
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
	redis_pass := os.Getenv("REDIS_PASSWORD")
	if redis_pass == "" {
		panic("Failed to get env for redis")
	}

	redisCacheTTL := 5 * time.Minute
	if ttl := os.Getenv("REDIS_CACHE_TTL"); ttl != "" {
		if parsed, err := time.ParseDuration(ttl); err == nil {
			redisCacheTTL = parsed
		} else {
			panic("Failed to parse REDIS_CACHE_TTL")
		}
	}

	var timerCronInterval time.Duration
	if interval := os.Getenv("QUIZ_TIMER_CRON_INTERVAL"); interval != "" {
		if parsed, err := time.ParseDuration(interval); err == nil {
			timerCronInterval = parsed
		} else {
			panic("Failed to parse QUIZ_TIMER_CRON_INTERVAL")
		}
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
		Redis: RedisConfig{Password: redis_pass, CacheTTL: redisCacheTTL, TimerCronInterval: timerCronInterval},
	}
}
