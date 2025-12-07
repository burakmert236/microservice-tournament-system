package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/burakmert236/goodswipe-common/config"
	"github.com/burakmert236/goodswipe-common/database"
	proto "github.com/burakmert236/goodswipe-common/generated/v1/grpc"
	utils "github.com/burakmert236/goodswipe-common/utils"
	"github.com/burakmert236/goodswipe-user-service/internal/handler"
	"github.com/burakmert236/goodswipe-user-service/internal/repository"
	"github.com/burakmert236/goodswipe-user-service/internal/service"
)

func main() {
	cfg, err := config.Load("../config")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	log.Printf("Starting user service in %s mode", cfg.Server.Environment)
	log.Printf("Using DynamoDB table: %s", cfg.DynamoDB.TableName)

	dynamoClient, err := database.NewDynamoDBClient(cfg)
	if err != nil {
		log.Fatalf("Failed to create DynamoDB client: %v", err)
	}

	userRepo := repository.NewUserRepository(dynamoClient)

	userService := service.NewUserService(userRepo)

	userHandler := handler.NewUserHandler(userService)

	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(utils.LoggingInterceptor),
	)

	proto.RegisterUserServiceServer(grpcServer, userHandler)
	reflection.Register(grpcServer)

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.Server.GRPCPort))
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	log.Printf("User service listening on port %d", cfg.Server.GRPCPort)

	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("Failed to serve: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")
	grpcServer.GracefulStop()
	log.Println("Server stopped")
}
