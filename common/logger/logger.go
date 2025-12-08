package logger

import (
	"fmt"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Logger struct {
	zap *zap.Logger
}

type Config struct {
	Level       string
	Format      string
	ServiceName string
}

func New(cfg Config) *Logger {
	if cfg.Level == "" {
		cfg.Level = "info"
	}
	if cfg.Format == "" {
		cfg.Format = "json"
	}

	level := parseLevel(cfg.Level)

	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		MessageKey:     "msg",
		CallerKey:      "caller",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	var encoder zapcore.Encoder
	if cfg.Format == "console" {
		encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	} else {
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	}

	core := zapcore.NewCore(
		encoder,
		zapcore.AddSync(os.Stdout),
		level,
	)

	zapLogger := zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1))

	if cfg.ServiceName != "" {
		zapLogger = zapLogger.With(zap.String("service", cfg.ServiceName))
	}

	return &Logger{zap: zapLogger}
}

func (l *Logger) Debug(msg string, fields ...interface{}) {
	l.zap.Debug(msg, convertFields(fields...)...)
}

func (l *Logger) Info(msg string, fields ...interface{}) {
	l.zap.Info(msg, convertFields(fields...)...)
}

func (l *Logger) Warn(msg string, fields ...interface{}) {
	l.zap.Warn(msg, convertFields(fields...)...)
}

func (l *Logger) Error(msg string, fields ...interface{}) {
	l.zap.Error(msg, convertFields(fields...)...)
}

func (l *Logger) Fatal(msg string, fields ...interface{}) {
	l.zap.Fatal(msg, convertFields(fields...)...)
}

func (l *Logger) With(fields ...interface{}) *Logger {
	return &Logger{
		zap: l.zap.With(convertFields(fields...)...),
	}
}

// convertFields converts key-value pairs to zap fields
func convertFields(keysAndValues ...interface{}) []zap.Field {
	if len(keysAndValues) == 0 {
		return nil
	}

	fields := make([]zap.Field, 0, len(keysAndValues)/2)

	for i := 0; i < len(keysAndValues); i += 2 {
		if i+1 >= len(keysAndValues) {
			break
		}

		key, ok := keysAndValues[i].(string)
		if !ok {
			key = fmt.Sprintf("%v", keysAndValues[i])
		}

		fields = append(fields, zap.Any(key, keysAndValues[i+1]))
	}

	return fields
}

// parseLevel converts string to zapcore.Level
func parseLevel(level string) zapcore.Level {
	switch level {
	case "debug":
		return zapcore.DebugLevel
	case "info":
		return zapcore.InfoLevel
	case "warn":
		return zapcore.WarnLevel
	case "error":
		return zapcore.ErrorLevel
	default:
		return zapcore.InfoLevel
	}
}

func Default(serviceName string) *Logger {
	return New(Config{
		Level:       "info",
		Format:      "json",
		ServiceName: serviceName,
	})
}

func Development(serviceName string) *Logger {
	return New(Config{
		Level:       "debug",
		Format:      "console",
		ServiceName: serviceName,
	})
}
