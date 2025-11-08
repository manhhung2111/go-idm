package handler_consumer

import (
	"context"

	"github.com/manhhung2111/go-idm/internal/dataaccess/kafka/producer"
	"github.com/manhhung2111/go-idm/internal/utils"
	"go.uber.org/zap"
)

type DownloadTaskCreateHandler interface {
	Handle(ctx context.Context, event producer.DownloadTaskCreated) error
}

type downloadTaskCreatedHandler struct {
	logger *zap.Logger
}

func NewDownloadTaskCreatedHandler(
	logger *zap.Logger,
) DownloadTaskCreateHandler {
	return &downloadTaskCreatedHandler{
		logger: logger,
	}
}

// Handle implements DownloadTaskCreateHandler.
func (d *downloadTaskCreatedHandler) Handle(ctx context.Context, event producer.DownloadTaskCreated) error {
	logger := utils.LoggerWithContext(ctx, d.logger).With(zap.Any("event", event))
	logger.Info("download task create event received")

	return nil
}