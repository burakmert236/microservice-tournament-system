package database

import (
	"context"
)

type TransactionRepository interface {
	Execute(ctx context.Context, transactionBuilder *TransactionBuilder) error
}

type transactionRepo struct {
	db *DynamoDBClient
}

func NewTransactionRepository(db *DynamoDBClient) TransactionRepository {
	return &transactionRepo{db: db}
}

func (r *transactionRepo) Execute(ctx context.Context, transactionBuilder *TransactionBuilder) error {
	return transactionBuilder.Execute(ctx, r.db.Client)
}
