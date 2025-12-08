package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/burakmert236/goodswipe-common/database"
	"github.com/burakmert236/goodswipe-common/models"
)

type ParticipationRepository interface {
	GetTransactionForAddingParticipation(ctx context.Context, participation *models.Participation) (types.Put, error)
	UpdateParticipationScore(ctx context.Context, userId string, tournamentId string, gainedScore int) error
}

type participationRepo struct {
	db *database.DynamoDBClient
}

func NewParticipationRRepository(db *database.DynamoDBClient) ParticipationRepository {
	return &participationRepo{db: db}
}

func (s *participationRepo) GetTransactionForAddingParticipation(
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

func (s *participationRepo) UpdateParticipationScore(
	ctx context.Context,
	userId string,
	tournamentId string,
	gainedScore int,
) error {
	_, err := s.db.Client.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName: aws.String(s.db.Table()),
		Key: map[string]types.AttributeValue{
			"PK": &types.AttributeValueMemberS{Value: models.UserPK(userId)},
			"SK": &types.AttributeValueMemberS{Value: models.TournamentPK(tournamentId)},
		},
		UpdateExpression: aws.String("ADD score :gainedScore SET updated_at = :now"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":gainedScore": &types.AttributeValueMemberN{Value: fmt.Sprintf("%d", gainedScore)},
			":now":         &types.AttributeValueMemberS{Value: time.Now().Format(time.RFC3339)},
		},
	})

	return err
}
