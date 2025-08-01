package config

import (
	"os"
	"strconv"
	"time"
)

// DBConfig holds the Postgres connection settings
type DBConfig struct {
	URL         string
	MaxOpenConn int
	ConnMaxIdle time.Duration
}

// LoadDBConfig loads configuration from environment
func LoadDBConfig() DBConfig {
	maxOpen, _ := strconv.Atoi(os.Getenv("NOTIF_DB_MAX_OPEN"))
	idle, _ := time.ParseDuration(os.Getenv("NOTIF_DB_CONN_IDLE"))
	return DBConfig{
		URL:         os.Getenv("NOTIF_DB_URL"),
		MaxOpenConn: maxOpen,
		ConnMaxIdle: idle,
	}
}
