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

type RewardClaimRepository interface {
	Create(ctx context.Context, userId, tournamentId string) error
	GetByIdempotency(ctx context.Context, userId, tournamentId string) (*models.RewardClaim, error)
	Delete(ctx context.Context, userId, tournamentId string) error
}

type rewardClaimRepo struct {
	db *database.DynamoDBClient
}

func NewRewardClaimRepository(db *database.DynamoDBClient) RewardClaimRepository {
	return &rewardClaimRepo{db: db}
}

func (r *rewardClaimRepo) Create(ctx context.Context, userId, tournamentId string) error {
	rewardClaim := &models.RewardClaim{
		UserId:       userId,
		TournamentId: tournamentId,
		CreatedAt:    time.Now().UTC(),

		PK: models.RewardClaimPK(userId),
		SK: models.TournamentPK(tournamentId),
	}

	item, err := attributevalue.MarshalMap(rewardClaim)
	if err != nil {
		return fmt.Errorf("failed to marshal reward claim: %w", err)
	}

	_, err = r.db.Client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName:           aws.String(r.db.Table()),
		Item:                item,
		ConditionExpression: aws.String("attribute_not_exists(PK)"),
	})

	if err != nil {
		return fmt.Errorf("failed to create reward claim: %w", err)
	}

	return nil
}

func (r *rewardClaimRepo) GetByIdempotency(
	ctx context.Context,
	userId, tournamentId string,
) (*models.RewardClaim, error) {
	result, err := r.db.Client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(r.db.Table()),
		Key: map[string]types.AttributeValue{
			"PK": &types.AttributeValueMemberS{Value: models.RewardClaimPK(userId)},
			"SK": &types.AttributeValueMemberS{Value: models.TournamentPK(tournamentId)},
		},
	})

	if err != nil {
		return nil, fmt.Errorf("failed to get reward claim: %w", err)
	}

	if result.Item == nil {
		return nil, nil
	}

	var rewardClaim models.RewardClaim
	if err := attributevalue.UnmarshalMap(result.Item, &rewardClaim); err != nil {
		return nil, fmt.Errorf("failed to unmarshal reward claim: %w", err)
	}

	return &rewardClaim, nil
}

func (r *rewardClaimRepo) Delete(ctx context.Context, userId, tournamentId string) error {
	_, err := r.db.Client.DeleteItem(ctx, &dynamodb.DeleteItemInput{
		TableName: aws.String(r.db.Table()),
		Key: map[string]types.AttributeValue{
			"PK": &types.AttributeValueMemberS{Value: models.RewardClaimPK(userId)},
			"SK": &types.AttributeValueMemberS{Value: models.TournamentPK(tournamentId)},
		},
	})

	if err != nil {
		return fmt.Errorf("failed to delete reward claim: %w", err)
	}

	return nil
}
