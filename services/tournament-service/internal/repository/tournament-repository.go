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
	GetActiveTournament(ctx context.Context) (*models.Tournament, error)
	GetById(ctx context.Context, tournamentId string) (*models.Tournament, error)
	Update(ctx context.Context, Tournament *models.Tournament) error
}

type tournamentRepo struct {
	db *database.DynamoDBClient
}

func NewTournamentRepository(db *database.DynamoDBClient) TournamentRepository {
	return &tournamentRepo{db: db}
}

func (r *tournamentRepo) Create(ctx context.Context, tournament *models.Tournament) error {
	tournament.PK = models.TournamentPK(tournament.TournamentId)
	tournament.SK = models.MetaSK()
	tournament.GSI1PK = models.TournamentGSI1PK()
	tournament.GSI1SK = models.StartTimeGSI1SK(tournament.StartsAt.Format(time.RFC3339))
	tournament.CreatedAt = time.Now()

	item, err := attributevalue.MarshalMap(tournament)
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

func (r *tournamentRepo) GetActiveTournament(ctx context.Context) (*models.Tournament, error) {
	result, err := r.db.Client.Query(ctx, &dynamodb.QueryInput{
		TableName:              aws.String(r.db.Table()),
		IndexName:              aws.String("GSI1"),
		KeyConditionExpression: aws.String("GSI1PK = :current"),
		FilterExpression:       aws.String("starts_at <= :now AND ends_at >= :now"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":current": &types.AttributeValueMemberS{Value: models.TournamentGSI1PK()},
			":now":     &types.AttributeValueMemberS{Value: time.Now().UTC().Format(time.RFC3339)},
		},
		Limit: aws.Int32(1),
	})

	if err != nil {
		return nil, fmt.Errorf("failed to get Tournament: %w", err)
	}

	if len(result.Items) <= 0 {
		return nil, errors.NewAppError(
			errors.ErrCodeNotFound,
			"Tournament not found",
			nil,
		)
	}

	var Tournament models.Tournament
	if err := attributevalue.UnmarshalMap(result.Items[0], &Tournament); err != nil {
		return nil, fmt.Errorf("failed to unmarshal Tournament: %w", err)
	}

	return &Tournament, nil
}

func (r *tournamentRepo) GetById(ctx context.Context, tournamentId string) (*models.Tournament, error) {
	result, err := r.db.Client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(r.db.Table()),
		Key: map[string]types.AttributeValue{
			"PK": &types.AttributeValueMemberS{Value: models.TournamentPK(tournamentId)},
			"SK": &types.AttributeValueMemberS{Value: models.MetaSK()},
		},
	})

	if err != nil {
		return nil, fmt.Errorf("failed to get tournament: %w", err)
	}

	if result.Item == nil {
		return nil, errors.NewAppError(
			errors.ErrCodeNotFound,
			"tournament not found",
			nil,
		)
	}

	var tournament models.Tournament
	if err := attributevalue.UnmarshalMap(result.Item, &tournament); err != nil {
		return nil, fmt.Errorf("failed to unmarshal tournament: %w", err)
	}

	return &tournament, nil
}

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
