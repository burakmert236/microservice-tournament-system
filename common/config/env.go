package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type EnvLoader struct {
	prefix string
}

func NewEnvLoader(prefix string) *EnvLoader {
	return &EnvLoader{prefix: prefix}
}

// GetString retrieves a string value from environment variable
// Returns defaultValue if not found
func (e *EnvLoader) GetString(key, defaultValue string) string {
	envKey := e.buildKey(key)
	if value := os.Getenv(envKey); value != "" {
		return value
	}
	return defaultValue
}

// GetStringRequired retrieves a required string value from environment variable
// Returns error if not found
func (e *EnvLoader) GetStringRequired(key string) (string, error) {
	envKey := e.buildKey(key)
	value := os.Getenv(envKey)
	if value == "" {
		return "", fmt.Errorf("required environment variable %s is not set", envKey)
	}
	return value, nil
}

// GetInt retrieves an integer value from environment variable
// Returns defaultValue if not found or invalid
func (e *EnvLoader) GetInt(key string, defaultValue int) int {
	envKey := e.buildKey(key)
	value := os.Getenv(envKey)
	if value == "" {
		return defaultValue
	}

	intValue, err := strconv.Atoi(value)
	if err != nil {
		return defaultValue
	}
	return intValue
}

// GetInt64 retrieves an int64 value from environment variable
func (e *EnvLoader) GetInt64(key string, defaultValue int64) int64 {
	envKey := e.buildKey(key)
	value := os.Getenv(envKey)
	if value == "" {
		return defaultValue
	}

	intValue, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return defaultValue
	}
	return intValue
}

// GetBool retrieves a boolean value from environment variable
// Accepts: "true", "1", "yes", "on" for true
// Accepts: "false", "0", "no", "off" for false
func (e *EnvLoader) GetBool(key string, defaultValue bool) bool {
	envKey := e.buildKey(key)
	value := strings.ToLower(os.Getenv(envKey))
	if value == "" {
		return defaultValue
	}

	switch value {
	case "true", "1", "yes", "on":
		return true
	case "false", "0", "no", "off":
		return false
	default:
		return defaultValue
	}
}

// GetBool retrieves a duration value from environment variable
func (e *EnvLoader) GetDuration(key string, defaultValue time.Duration) time.Duration {
	envKey := e.buildKey(key)
	valueStr := strings.ToLower(os.Getenv(envKey))
	if valueStr == "" {
		return defaultValue
	}

	value, err := time.ParseDuration(valueStr)
	if err != nil {
		fmt.Printf("Warning: Invalid duration for %s, using default %v\n", key, defaultValue)
		return defaultValue
	}

	return value
}

// buildKey constructs the full environment variable key with prefix
// Example: prefix="GOODSWIPE", key="AWS_REGION" -> "GOODSWIPE_AWS_REGION"
func (e *EnvLoader) buildKey(key string) string {
	if e.prefix == "" {
		return key
	}
	return fmt.Sprintf("%s_%s", e.prefix, key)
}
