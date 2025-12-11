package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/burakmert236/goodswipe-common/config"
	"github.com/burakmert236/goodswipe-user-service/app"
)

func main() {
	cfg, err := config.Load("../config")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	application, appErr := app.New(ctx, cfg)
	if appErr != nil {
		log.Fatalf("Failed to initialize application: %v", appErr)
	}

	if appErr := application.Start(); appErr != nil {
		log.Fatalf("Failed to start application: %v", appErr)
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down gracefully...")

	if appErr := application.Stop(); appErr != nil {
		log.Printf("Error during shutdown: %v", appErr)
	}
}
