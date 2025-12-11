package publisher

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

func (p *EventPublisher) PublishTournamentEntered(
	ctx context.Context,
	userId, displayName, groupId, tournamentId string,
) *apperrors.AppError {
	event := &protoevents.TournamentEntered{
		UserId:       userId,
		DisplayName:  displayName,
		GroupId:      groupId,
		TournamentId: tournamentId,
		TimeStamp:    time.Now().UTC().Unix(),
	}

	if err := p.publisher.PublishProto(ctx, commonevents.TournamentEntered, event); err != nil {
		p.logger.Error(fmt.Sprintf("Failed to publish tournament entered event: %v", err))
		return apperrors.Wrap(err, apperrors.CodeEventPublishError, "failed to publish tournament entered event")
	}

	p.logger.Info(fmt.Sprintf("Published tournament entered event for user: %s", userId))
	return nil
}

func (p *EventPublisher) PublishTournamentParticipationScoreUpdated(
	ctx context.Context,
	userId, groupId, tournamentId string,
	newScore int,
) *apperrors.AppError {
	event := &protoevents.TournamentParticipationScoreUpdated{
		UserId:       userId,
		GroupId:      groupId,
		TournamentId: tournamentId,
		NewScore:     int32(newScore),
		TimeStamp:    time.Now().UTC().Unix(),
	}

	if err := p.publisher.PublishProto(ctx, commonevents.TournamentParticipationScoreUpdated, event); err != nil {
		p.logger.Error(fmt.Sprintf("Failed to publish tournament score updated event: %v", err))
		return apperrors.Wrap(err, apperrors.CodeEventPublishError,
			"failed to publish tournament pariticipation score updated event")
	}

	p.logger.Info(fmt.Sprintf("Published tournament score updated event for user: %s", userId))
	return nil
}
