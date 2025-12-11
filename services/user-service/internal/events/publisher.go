package events

import (
	"context"
	"fmt"
	"time"

	apperrors "github.com/burakmert236/goodswipe-common/errors"
	commonevents "github.com/burakmert236/goodswipe-common/events"
	protoevents "github.com/burakmert236/goodswipe-common/generated/v1/events"
	"github.com/burakmert236/goodswipe-common/logger"
	"github.com/burakmert236/goodswipe-common/natsjetstream"
)

type EventPublisher struct {
	publisher *natsjetstream.Publisher
	logger    *logger.Logger
}

func NewEventPublisher(client *natsjetstream.Client, logger *logger.Logger) *EventPublisher {
	return &EventPublisher{
		publisher: natsjetstream.NewPublisher(client),
		logger:    logger,
	}
}

func (p *EventPublisher) PublishUserCreated(ctx context.Context, userId, displayName string) *apperrors.AppError {
	event := &protoevents.UserCreated{
		UserId:      userId,
		DisplayName: displayName,
		TimeStamp:   time.Now().UTC().Unix(),
	}

	if err := p.publisher.PublishProto(ctx, commonevents.UserCreated, event); err != nil {
		p.logger.Error(fmt.Sprintf("Failed to publish user created event: %v", err))
		return apperrors.Wrap(err, apperrors.CodeEventPublishError, "failed to publish user created event")
	}

	p.logger.Info(fmt.Sprintf("Published user created event for user: %s", userId))
	return nil
}

func (p *EventPublisher) PublishUserLevelUp(ctx context.Context, userId string, levelIncrease int, newLevel int) *apperrors.AppError {
	event := &protoevents.UserLevelUp{
		UserId:        userId,
		LevelIncrease: int32(levelIncrease),
		NewLevel:      int32(newLevel),
		TimeStamp:     time.Now().UTC().Unix(),
	}

	if err := p.publisher.PublishProto(ctx, commonevents.UserLevelUp, event); err != nil {
		p.logger.Error(fmt.Sprintf("Failed to publish user level up event: %v", err))
		return apperrors.Wrap(err, apperrors.CodeEventPublishError, "failed to publish user level up event")
	}

	p.logger.Info(fmt.Sprintf("Published user level up event for user: %s", userId))
	return nil
}
