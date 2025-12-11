package database

import (
	"context"

	"github.com/burakmert236/goodswipe-common/errors"
)

type TransactionRepository interface {
	Execute(ctx context.Context, transactionBuilder *TransactionBuilder) *errors.AppError
}

type transactionRepo struct {
	db *DynamoDBClient
}

func NewTransactionRepository(db *DynamoDBClient) TransactionRepository {
	return &transactionRepo{db: db}
}

func (r *transactionRepo) Execute(ctx context.Context, transactionBuilder *TransactionBuilder) *errors.AppError {
	return transactionBuilder.Execute(ctx, r.db.Client)
}
