package natsjetstream

import (
	"context"

	apperrors "github.com/burakmert236/goodswipe-common/errors"
	"google.golang.org/protobuf/proto"
)

type Publisher struct {
	client *Client
}

func NewPublisher(client *Client) *Publisher {
	return &Publisher{client: client}
}

func (p *Publisher) PublishProto(ctx context.Context, subject string, msg proto.Message) *apperrors.AppError {
	data, err := proto.Marshal(msg)
	if err != nil {
		return apperrors.Wrap(err, apperrors.CodeObjectMarshalError, "failed to marshal proto message")
	}

	return p.Publish(ctx, subject, data)
}

func (p *Publisher) Publish(ctx context.Context, subject string, data []byte) *apperrors.AppError {
	_, err := p.client.js.Publish(ctx, subject, data)
	if err != nil {
		return apperrors.Wrap(err, apperrors.CodeInternalServer, "failed to publish message")
	}
	return nil
}
