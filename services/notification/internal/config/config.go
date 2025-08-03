package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// Config holds the application settings loaded from environment variables.
type Config struct {
	WorkerLimit    int
	WorkerInterval time.Duration
	ConsumerConfig ConsumerConfig
	DBConfig       DBConfig
	AppCfg         AppConfig
}

type AppConfig struct {
	Port string
}

// ConsumerConfig holds all the Kafka consumer settings.
type ConsumerConfig struct {
	KafkaBrokers       []string
	KafkaTopic         string
	KafkaConsumerGroup string
}

// DBConfig holds the Postgres connection settings.
type DBConfig struct {
	URL         string
	MaxOpenConn int
	ConnMaxIdle time.Duration
}

// LoadConfig reads environment variables and returns a Config or an error.
func LoadConfig() (*Config, error) {
	var err error
	cfg := &Config{}

	// Helper closures
	getInt := func(key string, def int) (int, error) {
		if v := os.Getenv(key); v != "" {
			i, e := strconv.Atoi(v)
			if e != nil {
				return 0, fmt.Errorf("invalid %s: %w", key, e)
			}
			return i, nil
		}
		return def, nil
	}

	getDuration := func(key string, def time.Duration) (time.Duration, error) {
		if v := os.Getenv(key); v != "" {
			d, e := time.ParseDuration(v)
			if e != nil {
				return 0, fmt.Errorf("invalid %s: %w", key, e)
			}
			return d, nil
		}
		return def, nil
	}

	getString := func(key, def string) string {
		if v := os.Getenv(key); v != "" {
			return v
		}
		return def
	}

	// Worker settings
	if cfg.WorkerLimit, err = getInt("WORKER_LIMIT", 10); err != nil {
		return nil, err
	}
	if cfg.WorkerInterval, err = getDuration("WORKER_INTERVAL", 30*time.Second); err != nil {
		return nil, err
	}

	// Kafka settings
	cfg.ConsumerConfig.KafkaBrokers = strings.Split(getString("KAFKA_BROKERS", "localhost:9092"), ",")
	for i, b := range cfg.ConsumerConfig.KafkaBrokers {
		cfg.ConsumerConfig.KafkaBrokers[i] = strings.TrimSpace(b)
	}
	cfg.ConsumerConfig.KafkaTopic = getString("KAFKA_TOPIC", "notifications")
	cfg.ConsumerConfig.KafkaConsumerGroup = getString("KAFKA_CONSUMER_GROUP", "notification-workers")

	// DB settings
	cfg.DBConfig.URL = os.Getenv("DB_URL")
	if cfg.DBConfig.URL == "" {
		return nil, fmt.Errorf("DB_URL is required")
	}
	if cfg.DBConfig.MaxOpenConn, err = getInt("DB_MAX_OPEN_CONN", 10); err != nil {
		return nil, err
	}
	if cfg.DBConfig.ConnMaxIdle, err = getDuration("DB_CONN_MAX_IDLE", 5*time.Minute); err != nil {
		return nil, err
	}

	port, err := getInt("PORT", 8083)

	cfg.AppCfg.Port = strconv.Itoa(port)

	return cfg, nil
}
