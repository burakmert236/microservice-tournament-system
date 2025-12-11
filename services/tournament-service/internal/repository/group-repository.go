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

type GroupRepository interface {
	CreateGroup(ctx context.Context, group *models.Group) *apperrors.AppError
	FindAvailableGroup(ctx context.Context, tournamentId string) (*models.Group, *apperrors.AppError)

	// Transaction operations
	GetTransactionForAddingParticipant(ctx context.Context, groupId string, tournamentId string) types.Update
}

type groupRepo struct {
	db *database.DynamoDBClient
}

func NewGroupRepository(db *database.DynamoDBClient) GroupRepository {
	return &groupRepo{db: db}
}

func (r *groupRepo) CreateGroup(ctx context.Context, group *models.Group) *apperrors.AppError {
	group.PK = models.TournamentPK(group.TournamentId)
	group.SK = models.GroupSK(group.GroupId)
	group.CreatedAt = time.Now()

	item, err := attributevalue.MarshalMap(group)
	if err != nil {
		return apperrors.Wrap(err, apperrors.CodeObjectMarshalError, "failed to marshal group")
	}

	_, err = r.db.Client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName:           aws.String(r.db.Table()),
		Item:                item,
		ConditionExpression: aws.String("attribute_not_exists(PK)"),
	})

	if err != nil {
		return apperrors.Wrap(err, apperrors.CodeDatabaseError, "failed to create group")
	}

	return nil
}

func (r *groupRepo) FindAvailableGroup(ctx context.Context, tournamentId string) (*models.Group, *apperrors.AppError) {
	result, err := r.db.Client.Query(ctx, &dynamodb.QueryInput{
		TableName:              aws.String(r.db.Table()),
		KeyConditionExpression: aws.String("PK = :pk AND begins_with(SK, :sk)"),
		FilterExpression:       aws.String("participant_count < group_size"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":pk": &types.AttributeValueMemberS{Value: models.TournamentPK(tournamentId)},
			":sk": &types.AttributeValueMemberS{Value: models.GroupSKPrefix()},
		},
		Limit: aws.Int32(1),
	})

	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.CodeDatabaseError, "failed to get group")
	}

	if len(result.Items) <= 0 {
		return nil, apperrors.New(apperrors.CodeNotFound, "group not found")
	}

	var group models.Group
	if err := attributevalue.UnmarshalMap(result.Items[0], &group); err != nil {
		return nil, apperrors.Wrap(err, apperrors.CodeObjectUnmarshalError, "failed to unmarshal user")
	}

	return &group, nil
}

// Transaction Operations

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
			":inc":  &types.AttributeValueMemberN{Value: "1"},
			":zero": &types.AttributeValueMemberN{Value: "0"},
		},
		ConditionExpression: aws.String("attribute_not_exists(participant_count) OR participant_count < group_size"),
	}
}
