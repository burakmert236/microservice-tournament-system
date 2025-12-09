package config

import (
	"github.com/spf13/viper"
)

type Config struct {
	AWS      AWSConfig
	DynamoDB DynamoDBConfig
	Server   ServerConfig
	NATS     NATSConfig
}

type AWSConfig struct {
	Region          string
	AccessKeyID     string
	SecretAccessKey string
	Endpoint        string
}

type DynamoDBConfig struct {
	TableName        string
	MaxRetries       int
	ReadCapacity     int64
	WriteCapacity    int64
	UseLocalEndpoint bool
}

type ServerConfig struct {
	GRPCPort           int
	Environment        string
	LogLevel           string
	UserServiceAddress string
}

type NATSConfig struct {
	URL                  string
	MaxReconnect         int
	ReconnectWaitSeconds int
	TimeoutSeconds       int
}

func Load(configPath string) (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./config")
	viper.AddConfigPath(".")
	viper.AddConfigPath(configPath)

	viper.AutomaticEnv()
	viper.SetEnvPrefix("GOODSWIPE")

	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
