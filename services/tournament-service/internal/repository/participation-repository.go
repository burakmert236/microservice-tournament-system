package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/burakmert236/goodswipe-common/database"
	"github.com/burakmert236/goodswipe-common/models"
)

type ParticipationRepository interface {
	GetTransactionForAddingParticipation(ctx context.Context, participation *models.Participation) (types.Put, error)
}

type participationRRepo struct {
	db *database.DynamoDBClient
}

func NewParticipationRRepository(db *database.DynamoDBClient) ParticipationRepository {
	return &participationRRepo{db: db}
}

func (s *participationRRepo) GetTransactionForAddingParticipation(
	ctx context.Context,
	participation *models.Participation,
) (types.Put, error) {
	participation.PK = models.UserPK(participation.UserId)
	participation.SK = models.TournamentPK(participation.TournamentId)
	participation.CreatedAt = time.Now()

	item, err := attributevalue.MarshalMap(participation)
	if err != nil {
		return types.Put{}, fmt.Errorf("failed to marshal user: %w", err)
	}

	return types.Put{
		TableName:           aws.String(s.db.Table()),
		Item:                item,
		ConditionExpression: aws.String("attribute_not_exists(PK)"),
	}, nil
}
