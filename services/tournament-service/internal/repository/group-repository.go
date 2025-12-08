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
	group.PK = models.TournamentPK(group.TournamentId)
	group.SK = models.GroupSK(group.GroupId)
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
	result, err := r.db.Client.Query(ctx, &dynamodb.QueryInput{
		TableName:              aws.String(r.db.Table()),
		KeyConditionExpression: aws.String("PK = :pk AND begins_with(SK, :sk)"),
		FilterExpression:       aws.String("participant_count < group_size"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":pk":    &types.AttributeValueMemberS{Value: models.TournamentPK(tournamentId)},
			":sk":    &types.AttributeValueMemberS{Value: models.GroupSKPrefix()},
			":false": &types.AttributeValueMemberBOOL{Value: false},
		},
		Limit: aws.Int32(1),
	})

	if err != nil {
		return nil, fmt.Errorf("failed to get group: %w", err)
	}

	if len(result.Items) <= 0 {
		return nil, errors.NewAppError(
			errors.ErrCodeNotFound,
			"group not found",
			nil,
		)
	}

	var group models.Group
	if err := attributevalue.UnmarshalMap(result.Items[0], &group); err != nil {
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
			"PK": &types.AttributeValueMemberS{Value: models.TournamentPK(tournamentId)},
			"SK": &types.AttributeValueMemberS{Value: models.GroupSK(groupId)},
		},
		UpdateExpression: aws.String(`
			SET participant_count = if_not_exists(participant_count, :zero) + :inc
		`),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":inc":   &types.AttributeValueMemberN{Value: "1"},
			":zero":  &types.AttributeValueMemberN{Value: "0"},
			":false": &types.AttributeValueMemberBOOL{Value: false},
		},
		ConditionExpression: aws.String("attribute_not_exists(participant_count) OR participant_count < group_size"),
	}
}
