package events

import (
	"context"
	"fmt"
	"time"

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

func (p *EventPublisher) PublishUserLevelUp(ctx context.Context, userId string, levelIncrease int, newLevel int) error {
	event := &protoevents.UserLevelUp{
		UserId:        userId,
		LevelIncrease: int32(levelIncrease),
		NewLevel:      int32(newLevel),
		TimeStamp:     time.Now().Unix(),
	}

	if err := p.publisher.PublishProto(ctx, commonevents.UserLevelUp, event); err != nil {
		p.logger.Error(fmt.Sprintf("Failed to publish user level up event: %v", err))
		return fmt.Errorf("failed to publish event: %w", err)
	}

	p.logger.Info(fmt.Sprintf("Published user level up event for user: %s", userId))
	return nil
}
