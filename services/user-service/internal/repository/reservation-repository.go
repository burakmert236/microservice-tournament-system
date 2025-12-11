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

type ReservationRepository interface {
	GetById(ctx context.Context, userId, tournamentId string) (*models.Reservation, *apperrors.AppError)
	UpdateStatus(ctx context.Context, userId, tournamentId string, status models.ReservationStatus) *apperrors.AppError

	// Transaction operations
	GetCreateTransaction(ctx context.Context, reservation *models.Reservation) (types.Put, *apperrors.AppError)
	GetUpdateStatusTransaction(ctx context.Context, userId, tournamentId string, status models.ReservationStatus) types.Update
}

type reservationRepo struct {
	db *database.DynamoDBClient
}

func NewReservationRepository(db *database.DynamoDBClient) ReservationRepository {
	return &reservationRepo{db: db}
}

func (r *reservationRepo) GetById(ctx context.Context, userId, tournamentId string) (*models.Reservation, *apperrors.AppError) {
	result, err := r.db.Client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(r.db.Table()),
		Key: map[string]types.AttributeValue{
			"PK": &types.AttributeValueMemberS{Value: models.ReservationPK(userId)},
			"SK": &types.AttributeValueMemberS{Value: models.ReservationSK(tournamentId)},
		},
	})

	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.CodeDatabaseError, "failed to get reservation")
	}

	if result.Item == nil {
		return nil, apperrors.New(apperrors.CodeNotFound, "resefvation not found")
	}

	var reservation models.Reservation
	if err := attributevalue.UnmarshalMap(result.Item, &reservation); err != nil {
		return nil, apperrors.Wrap(err, apperrors.CodeObjectUnmarshalError, "failed to unmarshal reservation")
	}

	return &reservation, nil
}

func (r *reservationRepo) UpdateStatus(
	ctx context.Context,
	userId, tournamentId string,
	status models.ReservationStatus,
) *apperrors.AppError {
	transactionInfo := r.GetUpdateStatusTransaction(ctx, userId, tournamentId, status)

	_, err := r.db.Client.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName:                 aws.String(r.db.Table()),
		Key:                       transactionInfo.Key,
		UpdateExpression:          transactionInfo.UpdateExpression,
		ExpressionAttributeNames:  transactionInfo.ExpressionAttributeNames,
		ExpressionAttributeValues: transactionInfo.ExpressionAttributeValues,
	})

	if err != nil {
		return apperrors.Wrap(err, apperrors.CodeDatabaseError, "failed to update reservation status")
	}

	return nil
}

// Transaction Operations

func (r *reservationRepo) GetCreateTransaction(ctx context.Context, reservation *models.Reservation) (types.Put, *apperrors.AppError) {
	reservation.PK = models.ReservationPK(reservation.UserId)
	reservation.SK = models.ReservationSK(reservation.TournamentId)
	reservation.CreatedAt = time.Now().UTC()

	item, err := attributevalue.MarshalMap(reservation)
	if err != nil {
		return types.Put{}, apperrors.Wrap(err, apperrors.CodeObjectMarshalError, "failed to marshal reservation")
	}

	return types.Put{
		TableName:           aws.String(r.db.Table()),
		Item:                item,
		ConditionExpression: aws.String("attribute_not_exists(PK)"),
	}, nil
}

func (r *reservationRepo) GetUpdateStatusTransaction(ctx context.Context, userId, tournamentId string, status models.ReservationStatus) types.Update {
	return types.Update{
		TableName: aws.String(r.db.Table()),
		Key: map[string]types.AttributeValue{
			"PK": &types.AttributeValueMemberS{Value: models.ReservationPK(userId)},
			"SK": &types.AttributeValueMemberS{Value: models.ReservationSK(tournamentId)},
		},
		UpdateExpression: aws.String("SET #status = :status, updatedAt = :now"),
		ExpressionAttributeNames: map[string]string{
			"#status": "status",
		},
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":status": &types.AttributeValueMemberS{Value: string(status)},
			":now":    &types.AttributeValueMemberS{Value: time.Now().UTC().Format(time.RFC3339)},
		},
	}
}
