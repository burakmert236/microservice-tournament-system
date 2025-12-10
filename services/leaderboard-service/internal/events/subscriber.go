package events

import (
	"context"
	"fmt"

	"github.com/nats-io/nats.go/jetstream"

	commonevents "github.com/burakmert236/goodswipe-common/events"
	protoevents "github.com/burakmert236/goodswipe-common/generated/v1/events"
	"github.com/burakmert236/goodswipe-common/logger"
	"github.com/burakmert236/goodswipe-common/natsjetstream"
	"github.com/burakmert236/goodswipe-leaderboard-service/internal/service"
)

type EventSubscriber struct {
	natsClient         *natsjetstream.Client
	subscriber         *natsjetstream.Subscriber
	leaderboardService service.LeaderboardService
	logger             *logger.Logger
}

func NewEventSubscriber(
	natsClient *natsjetstream.Client,
	leaderboardService service.LeaderboardService,
	logger *logger.Logger,
) *EventSubscriber {
	return &EventSubscriber{
		natsClient:         natsClient,
		subscriber:         natsjetstream.NewSubscriber(natsClient),
		leaderboardService: leaderboardService,
		logger:             logger.With("component", "event-subscriber"),
	}
}

func (s *EventSubscriber) Start(ctx context.Context) error {
	s.logger.Info("Starting event subscriptions")

	if err := s.subscribeToUserEvents(ctx); err != nil {
		return fmt.Errorf("failed to subscribe to user events: %w", err)
	}

	if err := s.subscribeToTournamentEvents(ctx); err != nil {
		return fmt.Errorf("failed to subscribe to tournament events: %w", err)
	}

	s.logger.Info("All event subscriptions started")
	return nil
}

func (s *EventSubscriber) subscribeToUserEvents(ctx context.Context) error {
	cfg := natsjetstream.ConsumerConfig{
		StreamName:   commonevents.UserEventsStream,
		ConsumerName: "leaderboard-service-user-consumer",
		Durable:      "leaderboard-service-user-consumer",
		AckPolicy:    "explicit",
	}

	s.logger.Info("Subscribing to user events",
		"stream", cfg.StreamName,
		"consumer", cfg.ConsumerName,
	)

	return s.subscriber.Subscribe(ctx, cfg, s.handleUserEvents)
}

func (s *EventSubscriber) subscribeToTournamentEvents(ctx context.Context) error {
	cfg := natsjetstream.ConsumerConfig{
		StreamName:   commonevents.TournamentEventsStream,
		ConsumerName: "leaderboard-service-tournament-consumer",
		Durable:      "leaderboard-service-tournament-consumer",
		AckPolicy:    "explicit",
	}

	s.logger.Info("Subscribing to tournament events",
		"stream", cfg.StreamName,
		"consumer", cfg.ConsumerName,
	)

	return s.subscriber.Subscribe(ctx, cfg, s.handleTournamentEvents)
}

func (s *EventSubscriber) handleUserEvents(ctx context.Context, msg jetstream.Msg) error {
	subject := msg.Subject()

	s.logger.Debug("Received user event", "subject", subject)

	switch subject {
	case commonevents.UserCreated:
		return s.handleUserCreated(ctx, msg)
	default:
		s.logger.Warn("Unknown user event subject", "subject", subject)
		return nil
	}
}

func (s *EventSubscriber) handleTournamentEvents(ctx context.Context, msg jetstream.Msg) error {
	subject := msg.Subject()

	s.logger.Debug("Received tournament event", "subject", subject)

	switch subject {
	case commonevents.TournamentEntered:
		return s.handleTournamentEntered(ctx, msg)
	case commonevents.TournamentParticipationScoreUpdated:
		return s.handleTournamentParticipationScoreUpdated(ctx, msg)
	default:
		s.logger.Warn("Unknown tournament event subject", "subject", subject)
		return nil
	}
}

func (s *EventSubscriber) handleUserCreated(ctx context.Context, msg jetstream.Msg) error {
	var event protoevents.UserCreated
	if err := natsjetstream.UnmarshalProto(msg, &event); err != nil {
		s.logger.Error("Failed to unmarshal user created event",
			"error", err,
		)
		return fmt.Errorf("unmarshal error: %w", err)
	}

	s.logger.Info("Processing user created event",
		"user_id", event.UserId,
		"display_name", event.DisplayName,
	)

	if err := s.leaderboardService.AddGlobalUser(ctx, event.UserId, event.DisplayName); err != nil {
		s.logger.Error("Failed to add global user",
			"error", err,
			"user_id", event.UserId,
		)
		return fmt.Errorf("global user error: %w", err)
	}

	s.logger.Info("User created event processed successfully")

	return nil
}

func (s *EventSubscriber) handleTournamentEntered(ctx context.Context, msg jetstream.Msg) error {
	var event protoevents.TournamentEntered
	if err := natsjetstream.UnmarshalProto(msg, &event); err != nil {
		s.logger.Error("Failed to unmarshal tournament entered event",
			"error", err,
		)
		return fmt.Errorf("unmarshal error: %w", err)
	}

	s.logger.Info("Processing tournament entered event",
		"user_id", event.UserId,
	)

	if err := s.leaderboardService.AddUserToTournament(ctx, event.UserId, event.DisplayName, event.GroupId, event.TournamentId); err != nil {
		s.logger.Error("Failed to add tournament user",
			"error", err,
			"user_id", event.UserId,
		)
		return fmt.Errorf("tournament user error: %w", err)
	}

	s.logger.Info("Tournament entered event processed successfully")

	return nil
}

func (s *EventSubscriber) handleTournamentParticipationScoreUpdated(ctx context.Context, msg jetstream.Msg) error {
	var event protoevents.TournamentParticipationScoreUpdated
	if err := natsjetstream.UnmarshalProto(msg, &event); err != nil {
		s.logger.Error("Failed to unmarshal tournament participation score updated event",
			"error", err,
		)
		return fmt.Errorf("unmarshal error: %w", err)
	}

	s.logger.Info("Processing tournament participation score updated event",
		"user_id", event.UserId,
	)

	if err := s.leaderboardService.UpdateTournamentScore(ctx, event.UserId, event.DisplayName, event.TournamentId, int(event.NewScore)); err != nil {
		s.logger.Error("Failed to update tournament score",
			"error", err,
			"user_id", event.UserId,
		)
		return fmt.Errorf("tournament participation score update error: %w", err)
	}

	s.logger.Info("Tournament participation score updated event processed successfully")

	return nil
}
