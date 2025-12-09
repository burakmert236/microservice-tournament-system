package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/burakmert236/goodswipe-common/config"
	"github.com/burakmert236/goodswipe-leaderboard-service/app"
)

func main() {
	cfg, err := config.Load("../config")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	application, err := app.New(ctx, cfg)
	if err != nil {
		log.Fatalf("Failed to initialize application: %v", err)
	}

	if err := application.Start(); err != nil {
		log.Fatalf("Failed to start application: %v", err)
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down gracefully...")

	if err := application.Stop(); err != nil {
		log.Printf("Error during shutdown: %v", err)
	}
}
