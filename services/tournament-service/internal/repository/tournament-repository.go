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
	"github.com/burakmert236/goodswipe-common/errors"
	"github.com/burakmert236/goodswipe-common/models"
)

type TournamentRepository interface {
	Create(ctx context.Context, Tournament *models.Tournament) error
	GetById(ctx context.Context, TournamentId string) (*models.Tournament, error)
	Update(ctx context.Context, Tournament *models.Tournament) error
}

type tournamentRepo struct {
	db *database.DynamoDBClient
}

func NewTournamentRepository(db *database.DynamoDBClient) TournamentRepository {
	return &tournamentRepo{db: db}
}

// Create new Tournament
func (r *tournamentRepo) Create(ctx context.Context, Tournament *models.Tournament) error {
	Tournament.PK = models.TournamentPK(Tournament.TournamentId)
	Tournament.SK = models.MetaSK()
	Tournament.CreatedAt = time.Now()

	item, err := attributevalue.MarshalMap(Tournament)
	if err != nil {
		return fmt.Errorf("failed to marshal Tournament: %w", err)
	}

	_, err = r.db.Client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName:           aws.String(r.db.Table()),
		Item:                item,
		ConditionExpression: aws.String("attribute_not_exists(PK)"),
	})

	if err != nil {
		return fmt.Errorf("failed to create Tournament: %w", err)
	}

	return nil
}

// Fetch a Tournament with Tournament id
func (r *tournamentRepo) GetById(ctx context.Context, TournamentId string) (*models.Tournament, error) {
	result, err := r.db.Client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(r.db.Table()),
		Key: map[string]types.AttributeValue{
			"PK": &types.AttributeValueMemberS{Value: models.TournamentPK(TournamentId)},
			"SK": &types.AttributeValueMemberS{Value: models.MetaSK()},
		},
	})

	if err != nil {
		return nil, fmt.Errorf("failed to get Tournament: %w", err)
	}

	if result.Item == nil {
		return nil, errors.NewAppError(
			errors.ErrCodeNotFound,
			"Tournament not found",
			nil,
		)
	}

	var Tournament models.Tournament
	if err := attributevalue.UnmarshalMap(result.Item, &Tournament); err != nil {
		return nil, fmt.Errorf("failed to unmarshal Tournament: %w", err)
	}

	return &Tournament, nil
}

// Updates a Tournament
func (r *tournamentRepo) Update(ctx context.Context, Tournament *models.Tournament) error {
	Tournament.UpdatedAt = time.Now()

	item, err := attributevalue.MarshalMap(Tournament)
	if err != nil {
		return fmt.Errorf("failed to marshal Tournament: %w", err)
	}

	_, err = r.db.Client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName:           aws.String(r.db.Table()),
		Item:                item,
		ConditionExpression: aws.String("attribute_exists(PK)"),
	})

	if err != nil {
		return fmt.Errorf("failed to update Tournament: %w", err)
	}

	return nil
}
