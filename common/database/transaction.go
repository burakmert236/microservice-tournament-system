package database

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

type TransactionBuilder struct {
	items []types.TransactWriteItem
	limit int
}

func NewTransactionBuilder() *TransactionBuilder {
	return &TransactionBuilder{
		items: make([]types.TransactWriteItem, 0),
		limit: 100,
	}
}

func (tb *TransactionBuilder) AddPut(item types.Put) error {
	if len(tb.items) >= tb.limit {
		return fmt.Errorf("transaction limit exceeded: %d items", tb.limit)
	}
	tb.items = append(tb.items, types.TransactWriteItem{
		Put: &item,
	})
	return nil
}

func (tb *TransactionBuilder) AddUpdate(item types.Update) error {
	if len(tb.items) >= tb.limit {
		return fmt.Errorf("transaction limit exceeded: %d items", tb.limit)
	}
	tb.items = append(tb.items, types.TransactWriteItem{
		Update: &item,
	})
	return nil
}

func (tb *TransactionBuilder) AddDelete(item types.Delete) error {
	if len(tb.items) >= tb.limit {
		return fmt.Errorf("transaction limit exceeded: %d items", tb.limit)
	}
	tb.items = append(tb.items, types.TransactWriteItem{
		Delete: &item,
	})
	return nil
}

func (tb *TransactionBuilder) Execute(ctx context.Context, client *dynamodb.Client) error {
	if len(tb.items) == 0 {
		return fmt.Errorf("no items in transaction")
	}

	input := &dynamodb.TransactWriteItemsInput{
		TransactItems: tb.items,
	}

	_, err := client.TransactWriteItems(ctx, input)
	return err
}

func (tb *TransactionBuilder) Count() int {
	return len(tb.items)
}
