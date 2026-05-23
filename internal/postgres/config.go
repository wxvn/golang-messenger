package postgres

import (
	"time"
)

type Config struct {
	Host     string
	Port     string
	User     string
	Password string
	DB       string

	Timeout time.Duration
}
