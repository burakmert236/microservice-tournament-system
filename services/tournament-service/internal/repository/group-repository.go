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

type GroupRepository interface {
	CreateGroup(ctx context.Context, group *models.Group) error
	FindAvailableGroup(ctx context.Context, tournamentId string) (*models.Group, error)
	GetTransactionForAddingParticipant(ctx context.Context, groupId string, tournamentId string) types.Update
}

type groupRepo struct {
	db *database.DynamoDBClient
}

func NewGroupRepository(db *database.DynamoDBClient) GroupRepository {
	return &groupRepo{db: db}
}

func (r *groupRepo) CreateGroup(ctx context.Context, group *models.Group) error {
	group.PK = models.GroupSK(group.GroupId)
	group.SK = models.TournamentPK(group.TournamentId)
	group.CreatedAt = time.Now()

	item, err := attributevalue.MarshalMap(group)
	if err != nil {
		return fmt.Errorf("failed to marshal group: %w", err)
	}

	_, err = r.db.Client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName:           aws.String(r.db.Table()),
		Item:                item,
		ConditionExpression: aws.String("attribute_not_exists(PK)"),
	})

	if err != nil {
		return fmt.Errorf("failed to create group: %w", err)
	}

	return nil
}

func (r *groupRepo) FindAvailableGroup(ctx context.Context, tournamentId string) (*models.Group, error) {
	result, err := r.db.Client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(r.db.Table()),
		Key: map[string]types.AttributeValue{
			"SK":      &types.AttributeValueMemberS{Value: models.TournamentPK(tournamentId)},
			"is_full": &types.AttributeValueMemberS{Value: "false"},
		},
	})

	if err != nil {
		return nil, fmt.Errorf("failed to get group: %w", err)
	}

	if result.Item == nil {
		return nil, errors.NewAppError(
			errors.ErrCodeNotFound,
			"group not found",
			nil,
		)
	}

	var group models.Group
	if err := attributevalue.UnmarshalMap(result.Item, &group); err != nil {
		return nil, fmt.Errorf("failed to unmarshal user: %w", err)
	}

	return &group, nil
}

func (r *groupRepo) GetTransactionForAddingParticipant(
	ctx context.Context,
	groupId string,
	tournamentId string,
) types.Update {
	return types.Update{
		TableName: aws.String(r.db.Table()),
		Key: map[string]types.AttributeValue{
			"PK": &types.AttributeValueMemberS{Value: models.GroupSK(groupId)},
			"SK": &types.AttributeValueMemberS{Value: models.TournamentPK(tournamentId)},
		},
		UpdateExpression: aws.String(`
			SET participant_count = if_not_exists(participant_count, :zero) + :inc,
			    is_full = if_not_exists(participant_count, :zero) + :inc >= group_size
		`),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":inc":   &types.AttributeValueMemberN{Value: "1"},
			":zero":  &types.AttributeValueMemberN{Value: "0"},
			":false": &types.AttributeValueMemberBOOL{Value: false},
		},
		ConditionExpression: aws.String("attribute_not_exists(is_full) OR is_full = :false"),
	}
}
