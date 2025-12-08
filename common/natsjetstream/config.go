package natsjetstream

import "time"

type Config struct {
	URL           string
	MaxReconnect  int
	ReconnectWait time.Duration
	Timeout       time.Duration
}

type ConsumerConfig struct {
	StreamName   string
	ConsumerName string
	Durable      string
	AckPolicy    string
}
