package natsjetstream

import (
	"context"
	"fmt"

	"google.golang.org/protobuf/proto"
)

type Publisher struct {
	client *Client
}

func NewPublisher(client *Client) *Publisher {
	return &Publisher{client: client}
}

func (p *Publisher) PublishProto(ctx context.Context, subject string, msg proto.Message) error {
	data, err := proto.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal proto message: %w", err)
	}

	return p.Publish(ctx, subject, data)
}

func (p *Publisher) Publish(ctx context.Context, subject string, data []byte) error {
	_, err := p.client.js.Publish(ctx, subject, data)
	if err != nil {
		return fmt.Errorf("failed to publish message: %w", err)
	}
	return nil
}
