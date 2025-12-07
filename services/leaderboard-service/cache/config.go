package cache

import (
	"time"

	config "github.com/burakmert236/goodswipe-common/config"
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
	envLoader := config.NewEnvLoader("REDIS")

	return &RedisConfig{
		Address:      envLoader.GetString("ADDR", "localhost:6379"),
		Password:     envLoader.GetString("PASSWORD", ""),
		DB:           envLoader.GetInt("DB", 0),
		MaxRetries:   envLoader.GetInt("MAX_RETRIES", 3),
		DialTimeout:  envLoader.GetDuration("DIAL_TIMEOUT", 5*time.Second),
		ReadTimeout:  envLoader.GetDuration("READ_TIMEOUT", 3*time.Second),
		WriteTimeout: envLoader.GetDuration("WRITE_TIMEOUT", 3*time.Second),
		PoolSize:     envLoader.GetInt("POOL_SIZE", 10),
		MinIdleConns: envLoader.GetInt("MIN_IDLE_CONNS", 2),
	}
}
