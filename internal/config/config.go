package config

import (
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/wxvn/golang-messenger/internal/logger"
	"github.com/wxvn/golang-messenger/internal/postgres"
)

type Config struct {
	Addr            string
	ShutdownTimeout time.Duration
	JWTSecret       string

	Postgres postgres.Config
	Logger   logger.Config
}

func Load() Config {
	_ = godotenv.Load()

	timeout, err := time.ParseDuration(
		getEnv("HTTP_SHUTDOWN_TIMEOUT", "5s"),
	)
	if err != nil {
		panic(err)
	}

	return Config{
		Addr:            getEnv("HTTP_ADDR", "0.0.0.0:8080"),
		ShutdownTimeout: timeout,
		Postgres:        loadPostgresConfig(),

		JWTSecret: mustEnv("JWT_SECRET"),
		Logger: logger.Config{
			Level:  getEnv("LOG_LEVEL", "debug"),
			Folder: getEnv("LOG_FOLDER", "./logs"),
		},
	}
}

func loadPostgresConfig() postgres.Config {
	return postgres.Config{
		Host:     mustEnv("DB_HOST"),
		Port:     mustEnv("DB_PORT"),
		User:     mustEnv("DB_USER"),
		Password: mustEnv("DB_PASSWORD"),
		DB:       mustEnv("DB_DB"),

		Timeout: getEnvDuration("DB_TIMEOUT", "5s"),
	}
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}

	return fallback
}

func mustEnv(key string) string {
	value, ok := os.LookupEnv(key)
	if !ok {
		panic("missing env: " + key)
	}

	return value
}

func getEnvDuration(key, fallback string) time.Duration {
	val := getEnv(key, fallback)

	d, err := time.ParseDuration(val)
	if err != nil {
		panic("invalid duration for " + key + ": " + err.Error())
	}

	return d
}
