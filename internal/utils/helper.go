package utils

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"time"
)

func GetEnvRequired(key string) string {
	var value string

	if value = os.Getenv(key); value != "" {
		return value
	}

	if value == "" {
		log.Fatalf("%s env variable is required", key)
	}

	return ""
}

func GetEnv(key string, defaultValue string) string {
	var value string

	if value = os.Getenv(key); value != "" {
		return value
	}

	return defaultValue
}

func GetEnvAsInt(key string, defaultValue int) int {
	valueStr := GetEnv(key, "")
	if valueStr == "" {
		return defaultValue
	}

	value, err := strconv.Atoi(valueStr)
	if err != nil {
		fmt.Printf("Warning: Invalid integer for %s, using default %d\n", key, defaultValue)
		return defaultValue
	}

	return value
}

func GetEnvAsDuration(key string, defaultValue time.Duration) time.Duration {
	valueStr := GetEnv(key, "")
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

func WaitForGracefulShutdown() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c

	log.Println("Shutting down...")
}
