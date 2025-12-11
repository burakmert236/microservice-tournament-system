package config

import (
	apperrors "github.com/burakmert236/goodswipe-common/errors"
	"github.com/spf13/viper"
)

type Config struct {
	AWS      AWSConfig
	DynamoDB DynamoDBConfig
	Server   ServerConfig
	NATS     NATSConfig
	Redis    RedisConfig
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
	GRPCPort                  int
	Environment               string
	LogLevel                  string
	UserServiceAddress        string
	LeaderboardServiceAddress string
}

type NATSConfig struct {
	URL                  string
	MaxReconnect         int
	ReconnectWaitSeconds int
	TimeoutSeconds       int
}

type RedisConfig struct {
	Address  string
	Password string
}

func Load(configPath string) (*Config, *apperrors.AppError) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./config")
	viper.AddConfigPath(".")
	viper.AddConfigPath(configPath)

	viper.AutomaticEnv()
	viper.SetEnvPrefix("GOODSWIPE")

	if err := viper.ReadInConfig(); err != nil {
		return nil, apperrors.Wrap(err, apperrors.CodeInternalServer, "failed to read config")
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, apperrors.Wrap(err, apperrors.CodeObjectMarshalError, "failed to marshall config")
	}

	return &cfg, nil
}
