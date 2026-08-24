package redis

import "time"

type Config struct {
	Host     string
	Port     string
	Password string
	DB       int

	Timeout time.Duration
}
