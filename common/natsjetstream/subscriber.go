package natsjetstream

import (
	"context"
	"fmt"
	"log"

	"github.com/nats-io/nats.go/jetstream"
	"google.golang.org/protobuf/proto"
)

type Subscriber struct {
	client *Client
}

type MessageHandler func(ctx context.Context, msg jetstream.Msg) error

func NewSubscriber(client *Client) *Subscriber {
	return &Subscriber{client: client}
}

func (s *Subscriber) Subscribe(ctx context.Context, cfg ConsumerConfig, handler MessageHandler) error {
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
		return fmt.Errorf("failed to create consumer: %w", err)
	}

	_, err = consumer.Consume(func(msg jetstream.Msg) {
		if err := handler(ctx, msg); err != nil {
			log.Printf("Error handling message: %v", err)
			msg.Nak()
		} else {
			msg.Ack()
		}
	})

	return err
}

func UnmarshalProto(msg jetstream.Msg, pb proto.Message) error {
	return proto.Unmarshal(msg.Data(), pb)
}
