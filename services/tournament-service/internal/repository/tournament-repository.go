package repository

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"

	"github.com/burakmert236/goodswipe-common/database"
	apperrors "github.com/burakmert236/goodswipe-common/errors"
	"github.com/burakmert236/goodswipe-common/models"
)

type TournamentRepository interface {
	Create(ctx context.Context, Tournament *models.Tournament) *apperrors.AppError
	GetActiveTournament(ctx context.Context) (*models.Tournament, *apperrors.AppError)
	GetById(ctx context.Context, tournamentId string) (*models.Tournament, *apperrors.AppError)
	Update(ctx context.Context, Tournament *models.Tournament) *apperrors.AppError
}

type tournamentRepo struct {
	db *database.DynamoDBClient
}

func NewTournamentRepository(db *database.DynamoDBClient) TournamentRepository {
	return &tournamentRepo{db: db}
}

func (r *tournamentRepo) Create(ctx context.Context, tournament *models.Tournament) *apperrors.AppError {
	tournament.PK = models.TournamentPK(tournament.TournamentId)
	tournament.SK = models.MetaSK()
	tournament.GSI1PK = models.TournamentGSI1PK()
	tournament.GSI1SK = models.StartTimeGSI1SK(tournament.StartsAt.Format(time.RFC3339))
	tournament.CreatedAt = time.Now().UTC()

	item, err := attributevalue.MarshalMap(tournament)
	if err != nil {
		return apperrors.Wrap(err, apperrors.CodeObjectMarshalError, "failed to marshal tournament")
	}

	_, err = r.db.Client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName:           aws.String(r.db.Table()),
		Item:                item,
		ConditionExpression: aws.String("attribute_not_exists(PK)"),
	})

	if err != nil {
		return apperrors.Wrap(err, apperrors.CodeDatabaseError, "failed to create tournament")
	}

	return nil
}

func (r *tournamentRepo) GetActiveTournament(ctx context.Context) (*models.Tournament, *apperrors.AppError) {
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
		return nil, apperrors.Wrap(err, apperrors.CodeDatabaseError, "failed to get tournament")
	}

	if len(result.Items) <= 0 {
		return nil, apperrors.New(apperrors.CodeNotFound, "tournament not found")
	}

	var Tournament models.Tournament
	if err := attributevalue.UnmarshalMap(result.Items[0], &Tournament); err != nil {
		return nil, apperrors.Wrap(err, apperrors.CodeObjectUnmarshalError, "failed to unmarshal tournament")
	}

	return &Tournament, nil
}

func (r *tournamentRepo) GetById(ctx context.Context, tournamentId string) (*models.Tournament, *apperrors.AppError) {
	result, err := r.db.Client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(r.db.Table()),
		Key: map[string]types.AttributeValue{
			"PK": &types.AttributeValueMemberS{Value: models.TournamentPK(tournamentId)},
			"SK": &types.AttributeValueMemberS{Value: models.MetaSK()},
		},
	})

	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.CodeDatabaseError, "failed to get tournament")
	}

	if result.Item == nil {
		return nil, apperrors.New(apperrors.CodeNotFound, "tournament not found")
	}

	var tournament models.Tournament
	if err := attributevalue.UnmarshalMap(result.Item, &tournament); err != nil {
		return nil, apperrors.Wrap(err, apperrors.CodeObjectUnmarshalError, "failed to unmarshal tournament")
	}

	return &tournament, nil
}

func (r *tournamentRepo) Update(ctx context.Context, Tournament *models.Tournament) *apperrors.AppError {
	Tournament.UpdatedAt = time.Now().UTC()

	item, err := attributevalue.MarshalMap(Tournament)
	if err != nil {
		return apperrors.Wrap(err, apperrors.CodeObjectMarshalError, "failed to marshal tournament")
	}

	_, err = r.db.Client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName:           aws.String(r.db.Table()),
		Item:                item,
		ConditionExpression: aws.String("attribute_exists(PK)"),
	})

	if err != nil {
		return apperrors.Wrap(err, apperrors.CodeDatabaseError, "failed to update tournament")
	}

	return nil
}
