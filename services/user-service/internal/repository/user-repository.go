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

type UserRepository interface {
	Create(ctx context.Context, user *models.User) error
	GetById(ctx context.Context, userId string) (*models.User, error)
	Update(ctx context.Context, user *models.User) error
	UpdateLevelProgress(ctx context.Context, userId string, levelIncrease int, coinReward int) (int, error)
	AddCoin(ctx context.Context, userId string, coin int) error
}

type userRepo struct {
	db *database.DynamoDBClient
}

func NewUserRepository(db *database.DynamoDBClient) UserRepository {
	return &userRepo{db: db}
}

// Create new user
func (r *userRepo) Create(ctx context.Context, user *models.User) error {
	user.PK = models.UserPK(user.UserId)
	user.SK = models.ProfileSK()
	user.CreatedAt = time.Now()

	item, err := attributevalue.MarshalMap(user)
	if err != nil {
		return fmt.Errorf("failed to marshal user: %w", err)
	}

	_, err = r.db.Client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName:           aws.String(r.db.Table()),
		Item:                item,
		ConditionExpression: aws.String("attribute_not_exists(PK)"),
	})

	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

// Fetch a user with user id
func (r *userRepo) GetById(ctx context.Context, userId string) (*models.User, error) {
	result, err := r.db.Client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(r.db.Table()),
		Key: map[string]types.AttributeValue{
			"PK": &types.AttributeValueMemberS{Value: models.UserPK(userId)},
			"SK": &types.AttributeValueMemberS{Value: models.ProfileSK()},
		},
	})

	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	if result.Item == nil {
		return nil, errors.NewAppError(
			errors.ErrCodeNotFound,
			"user not found",
			nil,
		)
	}

	var user models.User
	if err := attributevalue.UnmarshalMap(result.Item, &user); err != nil {
		return nil, fmt.Errorf("failed to unmarshal user: %w", err)
	}

	return &user, nil
}

// Updates a user
func (r *userRepo) Update(ctx context.Context, user *models.User) error {
	user.UpdatedAt = time.Now()

	item, err := attributevalue.MarshalMap(user)
	if err != nil {
		return fmt.Errorf("failed to marshal user: %w", err)
	}

	_, err = r.db.Client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName:           aws.String(r.db.Table()),
		Item:                item,
		ConditionExpression: aws.String("attribute_exists(PK)"),
	})

	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	return nil
}

// Increase user level by given amount
func (r *userRepo) UpdateLevelProgress(
	ctx context.Context,
	userId string,
	levelIncrease int,
	coinReward int,
) (int, error) {
	if levelIncrease == 0 {
		return 0, nil
	}

	input := &dynamodb.UpdateItemInput{
		TableName: aws.String(r.db.Table()),
		Key: map[string]types.AttributeValue{
			"PK": &types.AttributeValueMemberS{Value: models.UserPK(userId)},
			"SK": &types.AttributeValueMemberS{Value: models.ProfileSK()},
		},
		UpdateExpression: aws.String("ADD #level :levelInc, coin :coinInc SET updated_at :updatedAt"),
		ExpressionAttributeNames: map[string]string{
			"#level": "level",
		},
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":levelInc":  &types.AttributeValueMemberN{Value: fmt.Sprintf("%d", levelIncrease)},
			":coinInc":   &types.AttributeValueMemberN{Value: fmt.Sprintf("%d", coinReward)},
			":updatedAt": &types.AttributeValueMemberN{Value: time.Now().UTC().Format(time.RFC3339)},
		},
		ConditionExpression: aws.String("attribute_exists(PK)"),
		ReturnValues:        types.ReturnValueUpdatedNew,
	}

	result, err := r.db.Client.UpdateItem(ctx, input)
	if err != nil {
		return 0, fmt.Errorf("failed to update level progress: %w", err)
	}

	var user models.User
	if err := attributevalue.UnmarshalMap(result.Attributes, &user); err != nil {
		return 0, fmt.Errorf("failed to unmarshal user: %w", err)
	}

	return user.Level, nil
}

// Increase user level by given amount
func (r *userRepo) AddCoin(ctx context.Context, userId string, coin int) error {
	input := &dynamodb.UpdateItemInput{
		TableName: aws.String(r.db.Table()),
		Key: map[string]types.AttributeValue{
			"PK": &types.AttributeValueMemberS{Value: models.UserPK(userId)},
			"SK": &types.AttributeValueMemberS{Value: models.ProfileSK()},
		},
		UpdateExpression: aws.String("ADD coin :coinInc"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":coinInc": &types.AttributeValueMemberN{Value: fmt.Sprintf("%d", coin)},
		},
		ReturnValues: types.ReturnValueUpdatedNew,
	}

	_, err := r.db.Client.UpdateItem(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to update coin: %w", err)
	}

	return nil
}
