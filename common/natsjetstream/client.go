package natsjetstream

import (
	"fmt"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
)

type Client struct {
	conn *nats.Conn
	js   jetstream.JetStream
	cfg  *Config
}

func NewClient(cfg *Config) (*Client, error) {
	opts := []nats.Option{
		nats.MaxReconnects(cfg.MaxReconnect),
		nats.ReconnectWait(cfg.ReconnectWait),
		nats.Timeout(cfg.Timeout),
		nats.DisconnectErrHandler(func(nc *nats.Conn, err error) {
			if err != nil {
				fmt.Printf("NATS disconnected: %v\n", err)
			}
		}),
		nats.ReconnectHandler(func(nc *nats.Conn) {
			fmt.Printf("NATS reconnected to %s\n", nc.ConnectedUrl())
		}),
	}

	nc, err := nats.Connect(cfg.URL, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to NATS: %w", err)
	}

	js, err := jetstream.New(nc)
	if err != nil {
		nc.Close()
		return nil, fmt.Errorf("failed to create JetStream context: %w", err)
	}

	client := &Client{
		conn: nc,
		js:   js,
		cfg:  cfg,
	}

	return client, nil
}

func (c *Client) Close() error {
	if c.conn != nil {
		c.conn.Close()
	}
	return nil
}

func (c *Client) JetStream() jetstream.JetStream {
	return c.js
}

func (c *Client) Conn() *nats.Conn {
	return c.conn
}

func (c *Client) IsConnected() bool {
	return c.conn != nil && c.conn.IsConnected()
}
