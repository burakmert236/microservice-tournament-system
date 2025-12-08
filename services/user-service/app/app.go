package app

import (
	"context"
	"fmt"
	"net"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/burakmert236/goodswipe-common/config"
	"github.com/burakmert236/goodswipe-common/database"
	protogrpc "github.com/burakmert236/goodswipe-common/generated/v1/grpc"
	"github.com/burakmert236/goodswipe-common/logger"
	"github.com/burakmert236/goodswipe-common/natsjetstream"
	"github.com/burakmert236/goodswipe-user-service/internal/events"
	"github.com/burakmert236/goodswipe-user-service/internal/handler"
	"github.com/burakmert236/goodswipe-user-service/internal/repository"
	"github.com/burakmert236/goodswipe-user-service/internal/service"
)

type App struct {
	cfg            *config.Config
	grpcServer     *grpc.Server
	db             *database.DynamoDBClient
	natsClient     *natsjetstream.Client
	logger         *logger.Logger
	eventPublisher *events.EventPublisher

	cleanup []func() error
}

func New(ctx context.Context, cfg *config.Config) (*App, error) {
	app := &App{
		cfg:     cfg,
		cleanup: make([]func() error, 0),
	}

	if err := app.initLogger(); err != nil {
		return nil, fmt.Errorf("failed to init logger: %w", err)
	}

	if err := app.initDatabase(); err != nil {
		return nil, fmt.Errorf("failed to init database: %w", err)
	}

	if err := app.initNATS(); err != nil {
		return nil, fmt.Errorf("failed to init NATS: %w", err)
	}

	if err := app.initMessaging(); err != nil {
		return nil, fmt.Errorf("failed to init messaging: %w", err)
	}

	if err := app.initGRPC(); err != nil {
		return nil, fmt.Errorf("failed to init gRPC: %w", err)
	}

	return app, nil
}

func (a *App) initLogger() error {
	a.logger = logger.Development("user-service")
	return nil
}

func (a *App) initDatabase() error {
	dynamoClient, err := database.NewDynamoDBClient(a.cfg)
	if err != nil {
		a.logger.Fatal("Failed to create DynamoDB client: %v", err)
	}

	a.db = dynamoClient
	return nil
}

func (a *App) initNATS() error {
	natsClient, err := natsjetstream.NewClient(&natsjetstream.Config{
		URL:           a.cfg.NATS.URL,
		MaxReconnect:  a.cfg.NATS.MaxReconnect,
		ReconnectWait: time.Duration(a.cfg.NATS.ReconnectWaitSeconds) * time.Second,
		Timeout:       time.Duration(a.cfg.NATS.TimeoutSeconds) * time.Second,
	})
	if err != nil {
		return err
	}

	a.natsClient = natsClient
	a.cleanup = append(a.cleanup, natsClient.Close)

	return nil
}

func (a *App) initMessaging() error {
	a.eventPublisher = events.NewEventPublisher(a.natsClient, a.logger)

	return nil
}

func (a *App) initGRPC() error {
	userRepo := repository.NewUserRepository(a.db)

	userService := service.NewUserService(
		userRepo,
		a.eventPublisher,
		a.logger,
	)

	userHandler := handler.NewUserHandler(userService, a.logger)

	a.grpcServer = grpc.NewServer(
		grpc.UnaryInterceptor(a.loggingInterceptor),
	)

	protogrpc.RegisterUserServiceServer(a.grpcServer, userHandler)
	reflection.Register(a.grpcServer)

	return nil
}

func (a *App) Start() error {
	go func() {
		lis, err := net.Listen("tcp", fmt.Sprintf(":%d", a.cfg.Server.GRPCPort))
		if err != nil {
			a.logger.Fatal("Failed to listen: %v", err)
		}

		a.logger.Info(fmt.Sprintf("gRPC server listening on %d", a.cfg.Server.GRPCPort))
		if err := a.grpcServer.Serve(lis); err != nil {
			a.logger.Fatal("Failed to serve: %v", err)
		}
	}()

	a.logger.Info("Application started successfully")

	return nil
}

func (a *App) Stop() error {
	a.logger.Info("Stopping application...")

	if a.grpcServer != nil {
		a.grpcServer.GracefulStop()
	}

	for _, cleanup := range a.cleanup {
		if err := cleanup(); err != nil {
			a.logger.Error(fmt.Sprintf("Cleanup error: %v", err))
		}
	}

	a.logger.Info("Application stopped")
	return nil
}

func (a *App) loggingInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	start := time.Now()
	resp, err := handler(ctx, req)
	a.logger.Info(fmt.Sprintf("Method: %s, Duration: %v", info.FullMethod, time.Since(start)))
	return resp, err
}
