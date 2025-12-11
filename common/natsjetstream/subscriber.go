package natsjetstream

import (
	"context"
	"log"

	apperrors "github.com/burakmert236/goodswipe-common/errors"
	"github.com/nats-io/nats.go/jetstream"
	"google.golang.org/protobuf/proto"
)

type Subscriber struct {
	client *Client
}

type MessageHandler func(ctx context.Context, msg jetstream.Msg) *apperrors.AppError

func NewSubscriber(client *Client) *Subscriber {
	return &Subscriber{client: client}
}

func (s *Subscriber) Subscribe(ctx context.Context, cfg ConsumerConfig, handler MessageHandler) *apperrors.AppError {
	consumerConfig := jetstream.ConsumerConfig{
		Name:    cfg.ConsumerName,
		Durable: cfg.Durable,
	}

	switch cfg.AckPolicy {
	case "explicit":
		consumerConfig.AckPolicy = jetstream.AckExplicitPolicy
	case "none":
		consumerConfig.AckPolicy = jetstream.AckNonePolicy
	case "all":
		consumerConfig.AckPolicy = jetstream.AckAllPolicy
	default:
		consumerConfig.AckPolicy = jetstream.AckExplicitPolicy
	}

	consumer, err := s.client.js.CreateOrUpdateConsumer(ctx, cfg.StreamName, consumerConfig)
	if err != nil {
		return apperrors.Wrap(err, apperrors.CodeInternalServer, "failed to create nats consumer")
	}

	_, err = consumer.Consume(func(msg jetstream.Msg) {
		if err := handler(ctx, msg); err != nil {
			log.Printf("Error handling message: %v", err)
			msg.Nak()
		} else {
			msg.Ack()
		}
	})
	if err != nil {
		apperrors.Wrap(err, apperrors.CodeInternalServer, "failed to consume nats message")
	}

	return nil
}

func UnmarshalProto(msg jetstream.Msg, pb proto.Message) *apperrors.AppError {
	err := proto.Unmarshal(msg.Data(), pb)
	return apperrors.Wrap(err, apperrors.CodeObjectUnmarshalError, "failed to unmarshal proto message")
}
