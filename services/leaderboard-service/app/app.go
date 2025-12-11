package app

import (
	"context"
	"fmt"
	"net"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/burakmert236/goodswipe-common/cache"
	"github.com/burakmert236/goodswipe-common/config"
	apperrors "github.com/burakmert236/goodswipe-common/errors"
	protogrpc "github.com/burakmert236/goodswipe-common/generated/v1/grpc"
	"github.com/burakmert236/goodswipe-common/logger"
	"github.com/burakmert236/goodswipe-common/natsjetstream"
	"github.com/burakmert236/goodswipe-leaderboard-service/internal/events"
	"github.com/burakmert236/goodswipe-leaderboard-service/internal/handler"
	"github.com/burakmert236/goodswipe-leaderboard-service/internal/repository"
	"github.com/burakmert236/goodswipe-leaderboard-service/internal/service"
)

type App struct {
	cfg                *config.Config
	redisClient        *cache.RedisClient
	grpcServer         *grpc.Server
	natsClient         *natsjetstream.Client
	logger             *logger.Logger
	leaderboardService service.LeaderboardService
	eventSubscriber    *events.EventSubscriber

	cleanup []func() error
}

func New(ctx context.Context, cfg *config.Config) (*App, *apperrors.AppError) {
	app := &App{
		cfg:     cfg,
		cleanup: make([]func() error, 0),
	}

	if err := app.initLogger(); err != nil {
		return nil, err
	}

	if err := app.initRedis(); err != nil {
		return nil, err
	}

	if err := app.initNATS(); err != nil {
		return nil, err
	}

	if err := app.initGRPC(); err != nil {
		return nil, err
	}

	if err := app.initMessaging(ctx); err != nil {
		return nil, err
	}

	return app, nil
}

func (a *App) initLogger() *apperrors.AppError {
	a.logger = logger.Development("leaderboard-service")
	return nil
}

func (a *App) initRedis() *apperrors.AppError {
	redisClient, err := cache.NewRedisClient(a.cfg.Redis)
	if err != nil {
		a.logger.Fatal(fmt.Sprintf("Redis client could not started: %s", err))
	}

	a.redisClient = redisClient

	return nil
}

func (a *App) initNATS() *apperrors.AppError {
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

func (a *App) initMessaging(ctx context.Context) *apperrors.AppError {
	a.eventSubscriber = events.NewEventSubscriber(a.natsClient, a.leaderboardService, a.logger)
	return a.eventSubscriber.Start(ctx)
}

func (a *App) initGRPC() *apperrors.AppError {
	leaderboardRepo := repository.NewLeaderboardRepository(a.redisClient, a.logger)

	a.leaderboardService = service.NewLeaderboardService(
		*leaderboardRepo,
		a.logger,
	)

	leaderboardHandler := handler.NewLeaderboardHandler(a.leaderboardService, a.logger)

	a.grpcServer = grpc.NewServer(
		grpc.UnaryInterceptor(a.loggingInterceptor),
	)

	protogrpc.RegisterLeaderboardServiceServer(a.grpcServer, leaderboardHandler)
	reflection.Register(a.grpcServer)

	return nil
}

func (a *App) Start() *apperrors.AppError {
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

func (a *App) Stop() *apperrors.AppError {
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
	start := time.Now().UTC()
	resp, err := handler(ctx, req)
	a.logger.Info(fmt.Sprintf("Method: %s, Duration: %v", info.FullMethod, time.Since(start)))
	return resp, err
}
