package utils

import (
	"context"
	"log"
	"os"
	"os/signal"

	"google.golang.org/grpc"
)

func LoggingInterceptor(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (interface{}, error) {
	log.Printf("gRPC call: %s", info.FullMethod)
	resp, err := handler(ctx, req)
	if err != nil {
		log.Printf("gRPC error: %v", err)
	}
	return resp, err
}

func WaitForGracefulShutdown() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c

	log.Println("Shutting down...")
}
