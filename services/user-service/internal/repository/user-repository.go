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
	apperrors "github.com/burakmert236/goodswipe-common/errors"
	"github.com/burakmert236/goodswipe-common/models"
)

type UserRepository interface {
	Create(ctx context.Context, user *models.User) *apperrors.AppError
	GetById(ctx context.Context, userId string) (*models.User, *apperrors.AppError)
	UpdateLevelProgress(ctx context.Context, userId string, levelIncrease int, coinReward int) (*models.User, *apperrors.AppError)
	AddCoin(ctx context.Context, userId string, coin int) *apperrors.AppError

	// Transactions operations
	GetCoinDeductionTransaction(ctx context.Context, userId string, amount int) types.Update
	GetCoinAdditionTransaction(ctx context.Context, userId string, amount int) types.Update
}

type userRepo struct {
	db *database.DynamoDBClient
}

func NewUserRepository(db *database.DynamoDBClient) UserRepository {
	return &userRepo{db: db}
}

func (r *userRepo) Create(ctx context.Context, user *models.User) *apperrors.AppError {
	user.PK = models.UserPK(user.UserId)
	user.SK = models.ProfileSK()
	user.CreatedAt = time.Now().UTC()

	item, err := attributevalue.MarshalMap(user)
	if err != nil {
		return apperrors.Wrap(err, apperrors.CodeObjectMarshalError, "failed to marshall user")
	}

	_, err = r.db.Client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName:           aws.String(r.db.Table()),
		Item:                item,
		ConditionExpression: aws.String("attribute_not_exists(PK)"),
	})

	if err != nil {
		return apperrors.Wrap(err, apperrors.CodeDatabaseError, "failed to create user")
	}

	return nil
}

func (r *userRepo) GetById(ctx context.Context, userId string) (*models.User, *apperrors.AppError) {
	result, err := r.db.Client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(r.db.Table()),
		Key: map[string]types.AttributeValue{
			"PK": &types.AttributeValueMemberS{Value: models.UserPK(userId)},
			"SK": &types.AttributeValueMemberS{Value: models.ProfileSK()},
		},
	})

	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.CodeDatabaseError, "failed to get user")
	}

	if result.Item == nil {
		return nil, apperrors.New(apperrors.CodeNotFound, "user not found")
	}

	var user models.User
	if err := attributevalue.UnmarshalMap(result.Item, &user); err != nil {
		return nil, apperrors.Wrap(err, apperrors.CodeObjectUnmarshalError, "failed to unmarshal user")
	}

	return &user, nil
}

func (r *userRepo) UpdateLevelProgress(
	ctx context.Context,
	userId string,
	levelIncrease int,
	coinReward int,
) (*models.User, *apperrors.AppError) {
	if levelIncrease == 0 {
		return nil, nil
	}

	input := &dynamodb.UpdateItemInput{
		TableName: aws.String(r.db.Table()),
		Key: map[string]types.AttributeValue{
			"PK": &types.AttributeValueMemberS{Value: models.UserPK(userId)},
			"SK": &types.AttributeValueMemberS{Value: models.ProfileSK()},
		},
		UpdateExpression: aws.String("ADD #level :levelInc, coin :coinInc SET updated_at = :updatedAt"),
		ExpressionAttributeNames: map[string]string{
			"#level": "level",
		},
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":levelInc":  &types.AttributeValueMemberN{Value: fmt.Sprintf("%d", levelIncrease)},
			":coinInc":   &types.AttributeValueMemberN{Value: fmt.Sprintf("%d", coinReward)},
			":updatedAt": &types.AttributeValueMemberS{Value: time.Now().UTC().Format(time.RFC3339)},
		},
		ConditionExpression: aws.String("attribute_exists(PK)"),
		ReturnValues:        types.ReturnValueAllNew,
	}

	result, err := r.db.Client.UpdateItem(ctx, input)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.CodeDatabaseError, "failed to update level progress")
	}

	var user models.User
	if err := attributevalue.UnmarshalMap(result.Attributes, &user); err != nil {
		return nil, apperrors.Wrap(err, apperrors.CodeObjectUnmarshalError, "failed to unmarshal user")
	}

	return &user, nil
}

func (r *userRepo) AddCoin(ctx context.Context, userId string, coin int) *apperrors.AppError {
	coinAdditionTransaction := r.GetCoinAdditionTransaction(ctx, userId, coin)

	input := &dynamodb.UpdateItemInput{
		TableName:                 aws.String(r.db.Table()),
		Key:                       coinAdditionTransaction.Key,
		UpdateExpression:          coinAdditionTransaction.UpdateExpression,
		ExpressionAttributeValues: coinAdditionTransaction.ExpressionAttributeValues,
	}

	_, err := r.db.Client.UpdateItem(ctx, input)
	if err != nil {
		return apperrors.Wrap(err, apperrors.CodeDatabaseError, "failed to update coin")
	}

	return nil
}

// Transaction Operations

func (r *userRepo) GetCoinDeductionTransaction(ctx context.Context, userId string, amount int) types.Update {
	now := time.Now().UTC()

	return types.Update{
		TableName: aws.String(r.db.Table()),
		Key: map[string]types.AttributeValue{
			"PK": &types.AttributeValueMemberS{Value: models.UserPK(userId)},
			"SK": &types.AttributeValueMemberS{Value: models.ProfileSK()},
		},
		UpdateExpression:    aws.String("SET coin = coin - :amount, updated_at = :now"),
		ConditionExpression: aws.String("coin >= :amount"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":amount": &types.AttributeValueMemberN{Value: fmt.Sprintf("%d", amount)},
			":now":    &types.AttributeValueMemberS{Value: now.Format(time.RFC3339)},
		},
	}
}

func (r *userRepo) GetCoinAdditionTransaction(ctx context.Context, userId string, amount int) types.Update {
	now := time.Now().UTC()

	return types.Update{
		TableName: aws.String(r.db.Table()),
		Key: map[string]types.AttributeValue{
			"PK": &types.AttributeValueMemberS{Value: models.UserPK(userId)},
			"SK": &types.AttributeValueMemberS{Value: models.ProfileSK()},
		},
		UpdateExpression: aws.String("SET coin = coin + :amount, updated_at = :now"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":amount": &types.AttributeValueMemberN{Value: fmt.Sprintf("%d", amount)},
			":now":    &types.AttributeValueMemberS{Value: now.Format(time.RFC3339)},
		},
	}
}
