package events

import (
	"context"
	"fmt"

	"github.com/nats-io/nats.go/jetstream"

	commonevents "github.com/burakmert236/goodswipe-common/events"
	protoevents "github.com/burakmert236/goodswipe-common/generated/v1/events"
	"github.com/burakmert236/goodswipe-common/logger"
	"github.com/burakmert236/goodswipe-common/natsjetstream"
	"github.com/burakmert236/goodswipe-tournament-service/internal/service"
)

type EventSubscriber struct {
	natsClient        *natsjetstream.Client
	subscriber        *natsjetstream.Subscriber
	tournamentService service.TournamentService
	logger            *logger.Logger
}

func NewEventSubscriber(
	natsClient *natsjetstream.Client,
	tournamentService service.TournamentService,
	logger *logger.Logger,
) *EventSubscriber {
	return &EventSubscriber{
		natsClient:        natsClient,
		subscriber:        natsjetstream.NewSubscriber(natsClient),
		tournamentService: tournamentService,
		logger:            logger.With("component", "event-subscriber"),
	}
}

func (s *EventSubscriber) Start(ctx context.Context) error {
	s.logger.Info("Starting event subscriptions")

	if err := s.subscribeToUserEvents(ctx); err != nil {
		return fmt.Errorf("failed to subscribe to user events: %w", err)
	}

	s.logger.Info("All event subscriptions started")
	return nil
}

func (s *EventSubscriber) subscribeToUserEvents(ctx context.Context) error {
	cfg := natsjetstream.ConsumerConfig{
		StreamName:   commonevents.UserEventsStream,
		ConsumerName: "tournament-service-user-consumer",
		Durable:      "tournament-service-user",
		AckPolicy:    "explicit",
	}

	s.logger.Info("Subscribing to user events",
		"stream", cfg.StreamName,
		"consumer", cfg.ConsumerName,
	)

	return s.subscriber.Subscribe(ctx, cfg, s.handleUserEvents)
}

func (s *EventSubscriber) handleUserEvents(ctx context.Context, msg jetstream.Msg) error {
	subject := msg.Subject()

	s.logger.Debug("Received user event", "subject", subject)

	switch subject {
	case commonevents.UserLevelUp:
		return s.handleUserLevelUp(ctx, msg)
	default:
		s.logger.Warn("Unknown user event subject", "subject", subject)
		return nil
	}
}

func (s *EventSubscriber) handleUserLevelUp(ctx context.Context, msg jetstream.Msg) error {
	var event protoevents.UserLevelUp
	if err := natsjetstream.UnmarshalProto(msg, &event); err != nil {
		s.logger.Error("Failed to unmarshal user level up event",
			"error", err,
		)
		return fmt.Errorf("unmarshal error: %w", err)
	}

	s.logger.Info("Processing user level up event",
		"user_id", event.UserId,
		"level_increase", event.LevelIncrease,
		"new_level", event.NewLevel,
	)

	if err := s.tournamentService.UpdateParticipationScore(ctx, event.UserId, int(event.LevelIncrease)); err != nil {
		s.logger.Error("Failed to update user progress",
			"error", err,
			"user_id", event.UserId,
		)
		return fmt.Errorf("update progress error: %w", err)
	}

	s.logger.Info("User level up event processed successfully")

	return nil
}
