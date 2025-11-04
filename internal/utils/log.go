package utils

import (
	"context"

	"github.com/manhhung2111/go-idm/internal/config"
	"go.uber.org/zap"
)

func getZapLoggerLevel(level string) zap.AtomicLevel {
	switch level {
	case "debug":
		return zap.NewAtomicLevelAt(zap.DebugLevel)
	case "info":
		return zap.NewAtomicLevelAt(zap.InfoLevel)
	case "warn":
		return zap.NewAtomicLevelAt(zap.WarnLevel)
	case "error":
		return zap.NewAtomicLevelAt(zap.ErrorLevel)
	case "panic":
		return zap.NewAtomicLevelAt(zap.PanicLevel)
	default:
		return zap.NewAtomicLevelAt(zap.InfoLevel)
	}
}

func InitializeLogger(logConfig config.Log) (*zap.Logger, func(), error) {
	zapLoggerConfig := zap.NewProductionConfig()
	zapLoggerConfig.Level = getZapLoggerLevel(logConfig.Level)

	logger, err := zapLoggerConfig.Build()
	if err != nil {
		return nil, nil, err
	}

	cleanup := func() {
		// deliberately ignore the returned error here
		_ = logger.Sync()
	}

	return logger, cleanup, err
}

func LoggerWithContext(_ context.Context, logger *zap.Logger) *zap.Logger {
	return logger
}
