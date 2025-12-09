package app

import (
	"context"
	"fmt"
	"net"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"

	"github.com/burakmert236/goodswipe-common/config"
	"github.com/burakmert236/goodswipe-common/database"
	commonevents "github.com/burakmert236/goodswipe-common/events"
	protogrpc "github.com/burakmert236/goodswipe-common/generated/v1/grpc"
	"github.com/burakmert236/goodswipe-common/logger"
	"github.com/burakmert236/goodswipe-common/natsjetstream"
	publisher "github.com/burakmert236/goodswipe-tournament-service/internal/events/publisher"
	subscriber "github.com/burakmert236/goodswipe-tournament-service/internal/events/subscriber"
	"github.com/burakmert236/goodswipe-tournament-service/internal/handler"
	"github.com/burakmert236/goodswipe-tournament-service/internal/repository"
	"github.com/burakmert236/goodswipe-tournament-service/internal/scheduler"
	"github.com/burakmert236/goodswipe-tournament-service/internal/service"
	"github.com/nats-io/nats.go/jetstream"
)

type App struct {
	cfg               *config.Config
	grpcServer        *grpc.Server
	db                *database.DynamoDBClient
	natsClient        *natsjetstream.Client
	logger            *logger.Logger
	tournamentService service.TournamentService
	scheduler         *scheduler.Scheduler
	eventPublisher    *publisher.EventPublisher
	eventSubscriber   *subscriber.EventSubscriber

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

	if err := app.initNATS(ctx); err != nil {
		return nil, fmt.Errorf("failed to init NATS: %w", err)
	}

	if err := app.initMessagePublisher(ctx); err != nil {
		return nil, fmt.Errorf("failed to init message publisher: %w", err)
	}

	if err := app.initGRPC(); err != nil {
		return nil, fmt.Errorf("failed to init gRPC: %w", err)
	}

	if err := app.initMessageSubscriber(ctx); err != nil {
		return nil, fmt.Errorf("failed to init message subscriber: %w", err)
	}

	if err := app.initScheduler(); err != nil {
		return nil, fmt.Errorf("failed to init scheduler: %w", err)
	}

	return app, nil
}

func (a *App) initLogger() error {
	a.logger = logger.Development("tournament-service")
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

func (a *App) initNATS(ctx context.Context) error {
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

	stream := jetstream.StreamConfig{
		Name:     commonevents.TournamentEventsStream,
		Subjects: []string{commonevents.TournamentEventsWildcard},
	}

	if _, err := a.natsClient.JetStream().CreateOrUpdateStream(ctx, stream); err != nil {
		a.logger.Error("Failed to create stream",
			"error", err,
			"stream", stream.Name,
		)
		return err
	}
	a.logger.Info("Stream ready", "stream", stream.Name)

	a.cleanup = append(a.cleanup, natsClient.Close)

	return nil
}

func (a *App) initMessageSubscriber(ctx context.Context) error {
	a.eventSubscriber = subscriber.NewEventSubscriber(a.natsClient, a.tournamentService, a.logger)
	return a.eventSubscriber.Start(ctx)
}

func (a *App) initGRPC() error {
	userServiceAddr := a.cfg.Server.UserServiceAddress
	userConn, err := grpc.NewClient(userServiceAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		a.logger.Fatal("Failed to connect to User Service: %v", err)
	}

	a.cleanup = append(a.cleanup, userConn.Close)

	userClient := protogrpc.NewUserServiceClient(userConn)
	a.logger.Info("Connected to User Service at %s", userServiceAddr)

	tournamentRepo := repository.NewTournamentRepository(a.db)
	participationRepo := repository.NewParticipationRRepository(a.db)
	groupRepo := repository.NewGroupRepository(a.db)
	transactionRepo := database.NewTransactionRepository(a.db)

	a.tournamentService = service.NewTournamentService(
		tournamentRepo,
		participationRepo,
		groupRepo,
		transactionRepo,
		a.eventPublisher,
		userClient,
		a.logger,
	)

	tournamentHandler := handler.NewTournamentHandler(a.tournamentService, a.logger)

	a.grpcServer = grpc.NewServer(
		grpc.UnaryInterceptor(a.loggingInterceptor),
	)

	protogrpc.RegisterTournamentServiceServer(a.grpcServer, tournamentHandler)
	reflection.Register(a.grpcServer)

	return nil
}

func (a *App) initMessagePublisher(ctx context.Context) error {
	a.eventPublisher = publisher.NewEventPublisher(a.natsClient, a.logger)
	return nil
}

func (a *App) initScheduler() error {
	tournamentSchedular := scheduler.NewTournamentScheduler(a.tournamentService)
	a.scheduler = scheduler.NewScheduler(tournamentSchedular)

	a.cleanup = append(a.cleanup, a.scheduler.Stop)

	return nil
}

func (a *App) Start() error {
	go func() {
		lis, err := net.Listen("tcp", fmt.Sprintf(":%d", a.cfg.Server.GRPCPort))
		if err != nil {
			a.logger.Fatal("Failed to listen: %v", err)
		}

		go a.scheduler.Start()
		a.logger.Info("Tournament generation scheduler is started")

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
