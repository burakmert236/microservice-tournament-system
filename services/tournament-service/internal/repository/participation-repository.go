package repository

import (
	"context"
	"errors"
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
	GetByUserAndTournament(ctx context.Context, userId, tournamentId string) (*models.Participation, error)
	UpdateRewardProcessing(ctx context.Context, userId, tournamentId string) (*models.Participation, error)
	UpdateRewardUnclaimed(ctx context.Context, userId, tournamentId string) (*models.Participation, error)
	UpdateRewardClaimed(ctx context.Context, userId, tournamentId string) (*models.Participation, error)
	UpdateParticipationScore(ctx context.Context, userId string, tournamentId string, gainedScore int) (*models.Participation, error)

	// Transactions
	GetTransactionForAddingParticipation(ctx context.Context, participation *models.Participation) (types.Put, error)
}

type participationRepo struct {
	db *database.DynamoDBClient
}

func NewParticipationRRepository(db *database.DynamoDBClient) ParticipationRepository {
	return &participationRepo{db: db}
}

func (s *participationRepo) GetByUserAndTournament(
	ctx context.Context,
	userId, tournamentId string,
) (*models.Participation, error) {
	result, err := s.db.Client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(s.db.Table()),
		Key: map[string]types.AttributeValue{
			"PK": &types.AttributeValueMemberS{Value: models.UserPK(userId)},
			"SK": &types.AttributeValueMemberS{Value: models.TournamentPK(tournamentId)},
		},
	})

	if err != nil {
		return nil, fmt.Errorf("failed to get participation by user and tournament: %w", err)
	}

	if result.Item == nil {
		return nil, nil
	}

	var participation models.Participation
	if err := attributevalue.UnmarshalMap(result.Item, &participation); err != nil {
		return nil, fmt.Errorf("failed to unmarshal participation: %w", err)
	}

	return &participation, nil
}

func (s *participationRepo) UpdateRewardProcessing(
	ctx context.Context,
	userId, tournamentId string,
) (*models.Participation, error) {
	result, err := s.db.Client.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName: aws.String(s.db.Table()),
		Key: map[string]types.AttributeValue{
			"PK": &types.AttributeValueMemberS{Value: models.UserPK(userId)},
			"SK": &types.AttributeValueMemberS{Value: models.TournamentPK(tournamentId)},
		},
		UpdateExpression: aws.String("SET reward_claim_status = :processing"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":processing": &types.AttributeValueMemberS{Value: string(models.Processing)},
			":unclaimed":  &types.AttributeValueMemberS{Value: string(models.Unclaimed)},
		},
		ConditionExpression: aws.String("attribute_exists(PK) AND reward_claim_status = :unclaimed"),
		ReturnValues:        types.ReturnValueAllNew,
	})

	if err != nil {
		var ccf *types.ConditionalCheckFailedException
		if errors.As(err, &ccf) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to update participation claim as processing: %w", err)
	}

	var participation models.Participation
	if err := attributevalue.UnmarshalMap(result.Attributes, &participation); err != nil {
		return nil, fmt.Errorf("failed to unmarshal participation: %w", err)
	}

	return &participation, nil
}

func (s *participationRepo) UpdateRewardUnclaimed(
	ctx context.Context,
	userId, tournamentId string,
) (*models.Participation, error) {
	result, err := s.db.Client.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName: aws.String(s.db.Table()),
		Key: map[string]types.AttributeValue{
			"PK": &types.AttributeValueMemberS{Value: models.UserPK(userId)},
			"SK": &types.AttributeValueMemberS{Value: models.TournamentPK(tournamentId)},
		},
		UpdateExpression: aws.String("SET reward_claim_status = :unclaimed"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":processing": &types.AttributeValueMemberS{Value: string(models.Processing)},
			":unclaimed":  &types.AttributeValueMemberS{Value: string(models.Unclaimed)},
		},
		ConditionExpression: aws.String("attribute_exists(PK) AND reward_claim_status = :processing"),
		ReturnValues:        types.ReturnValueAllNew,
	})

	if err != nil {
		var ccf *types.ConditionalCheckFailedException
		if errors.As(err, &ccf) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to update participation claim as processing: %w", err)
	}

	var participation models.Participation
	if err := attributevalue.UnmarshalMap(result.Attributes, &participation); err != nil {
		return nil, fmt.Errorf("failed to unmarshal participation: %w", err)
	}

	return &participation, nil
}

func (s *participationRepo) UpdateRewardClaimed(
	ctx context.Context,
	userId, tournamentId string,
) (*models.Participation, error) {
	result, err := s.db.Client.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName: aws.String(s.db.Table()),
		Key: map[string]types.AttributeValue{
			"PK": &types.AttributeValueMemberS{Value: models.UserPK(userId)},
			"SK": &types.AttributeValueMemberS{Value: models.TournamentPK(tournamentId)},
		},
		UpdateExpression: aws.String("SET reward_claim_status = :claimed"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":processing": &types.AttributeValueMemberS{Value: string(models.Processing)},
			":claimed":    &types.AttributeValueMemberS{Value: string(models.Claimed)},
		},
		ConditionExpression: aws.String("attribute_exists(PK) AND reward_claim_status = :processing"),
		ReturnValues:        types.ReturnValueAllNew,
	})

	if err != nil {
		var ccf *types.ConditionalCheckFailedException
		if errors.As(err, &ccf) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to update participation claim as processing: %w", err)
	}

	var participation models.Participation
	if err := attributevalue.UnmarshalMap(result.Attributes, &participation); err != nil {
		return nil, fmt.Errorf("failed to unmarshal participation: %w", err)
	}

	return &participation, nil
}

func (s *participationRepo) UpdateParticipationScore(
	ctx context.Context,
	userId string,
	tournamentId string,
	gainedScore int,
) (*models.Participation, error) {
	result, err := s.db.Client.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName: aws.String(s.db.Table()),
		Key: map[string]types.AttributeValue{
			"PK": &types.AttributeValueMemberS{Value: models.UserPK(userId)},
			"SK": &types.AttributeValueMemberS{Value: models.TournamentPK(tournamentId)},
		},
		UpdateExpression: aws.String("ADD score :gainedScore SET updated_at = :now"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":gainedScore": &types.AttributeValueMemberN{Value: fmt.Sprintf("%d", gainedScore)},
			":now":         &types.AttributeValueMemberS{Value: time.Now().UTC().Format(time.RFC3339)},
		},
		ConditionExpression: aws.String("attribute_exists(PK)"),
		ReturnValues:        types.ReturnValueAllNew,
	})

	if err != nil {
		var ccf *types.ConditionalCheckFailedException
		if errors.As(err, &ccf) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to update participation score: %w", err)
	}

	var participation models.Participation
	if err := attributevalue.UnmarshalMap(result.Attributes, &participation); err != nil {
		return nil, fmt.Errorf("failed to unmarshal participation: %w", err)
	}

	return &participation, nil
}

// Transactions

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
