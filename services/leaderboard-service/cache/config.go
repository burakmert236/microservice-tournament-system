package cache

import (
	"time"

	utils "github.com/burakmert236/goodswipe/internal/utils"
)

type RedisConfig struct {
	Address      string
	Password     string
	DB           int
	MaxRetries   int
	DialTimeout  time.Duration
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	PoolSize     int
	MinIdleConns int
}

func Load() *RedisConfig {
	return &RedisConfig{
		Address:      utils.GetEnv("REDIS_ADDR", "localhost:6379"),
		Password:     utils.GetEnv("REDIS_PASSWORD", ""),
		DB:           utils.GetEnvAsInt("REDIS_DB", 0),
		MaxRetries:   utils.GetEnvAsInt("REDIS_MAX_RETRIES", 3),
		DialTimeout:  utils.GetEnvAsDuration("REDIS_DIAL_TIMEOUT", 5*time.Second),
		ReadTimeout:  utils.GetEnvAsDuration("REDIS_READ_TIMEOUT", 3*time.Second),
		WriteTimeout: utils.GetEnvAsDuration("REDIS_WRITE_TIMEOUT", 3*time.Second),
		PoolSize:     utils.GetEnvAsInt("REDIS_POOL_SIZE", 10),
		MinIdleConns: utils.GetEnvAsInt("REDIS_MIN_IDLE_CONNS", 2),
	}
}
