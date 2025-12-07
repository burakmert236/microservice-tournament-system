package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/burakmert236/goodswipe-common/config"
	"github.com/burakmert236/goodswipe-common/database"

	// "github.com/burakmert236/goodswipe-tournament-service/internal/handler"
	"github.com/burakmert236/goodswipe-tournament-service/internal/repository"
	"github.com/burakmert236/goodswipe-tournament-service/internal/scheduler"
	"github.com/burakmert236/goodswipe-tournament-service/internal/service"
)

func main() {
	cfg, err := config.Load("../config")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	log.Printf("Starting tournament service in %s mode", cfg.Server.Environment)
	log.Printf("Using DynamoDB table: %s", cfg.DynamoDB.TableName)

	dynamoClient, err := database.NewDynamoDBClient(cfg)
	if err != nil {
		log.Fatalf("Failed to create DynamoDB client: %v", err)
	}

	tournamentRepo := repository.NewTournamentRepository(dynamoClient)
	participationRepo := repository.NewParticipationRRepository(dynamoClient)
	groupRepo := repository.NewGroupRepository(dynamoClient)

	tournamentService := service.NewTournamentService(tournamentRepo, participationRepo, groupRepo)

	tournamentSchedular := scheduler.NewTournamentScheduler(tournamentService)
	taskScheduler := scheduler.NewScheduler(tournamentSchedular)
	taskScheduler.Start()

	// tournamentHandler := handler.NewHandler(tournamentService)

	// grpcServer := grpc.NewServer(
	// 	grpc.UnaryInterceptor(utils.LoggingInterceptor),
	// )

	// proto.RegisterUserServiceServer(grpcServer, tournamentHandler)
	// reflection.Register(grpcServer)

	// lis, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.Server.GRPCPort))
	// if err != nil {
	// 	log.Fatalf("Failed to listen: %v", err)
	// }

	// log.Printf("User service listening on port %d", cfg.Server.GRPCPort)

	// go func() {
	// 	if err := grpcServer.Serve(lis); err != nil {
	// 		log.Fatalf("Failed to serve: %v", err)
	// 	}
	// }()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	// grpcServer.GracefulStop()
	taskScheduler.Stop()

	log.Println("Server stopped")
}
