package database

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	aws_config "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/burakmert236/goodswipe-common/config"
)

type DynamoDBClient struct {
	Client    *dynamodb.Client
	TableName string
}

func NewDynamoDBClient(cfg *config.Config) (*DynamoDBClient, error) {
	ctx := context.Background()

	var awsCfg aws.Config
	var err error

	if cfg.DynamoDB.UseLocalEndpoint {
		// Local DynamoDB for development
		awsCfg, err = aws_config.LoadDefaultConfig(ctx,
			aws_config.WithRegion(cfg.AWS.Region),
			aws_config.WithBaseEndpoint(cfg.AWS.Endpoint),
			aws_config.WithCredentialsProvider(
				credentials.NewStaticCredentialsProvider(
					"dummy",
					"dummy",
					"",
				),
			),
		)
	} else {
		// Production AWS
		awsCfg, err = aws_config.LoadDefaultConfig(ctx,
			aws_config.WithRegion(cfg.AWS.Region),
		)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	client := dynamodb.NewFromConfig(awsCfg, func(o *dynamodb.Options) {
		o.RetryMaxAttempts = cfg.DynamoDB.MaxRetries
	})

	return &DynamoDBClient{
		Client:    client,
		TableName: cfg.DynamoDB.TableName,
	}, nil
}

// Helper method to get table name
func (c *DynamoDBClient) Table() string {
	return c.TableName
}
