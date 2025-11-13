package handler_consumer

import (
	"context"

	"github.com/manhhung2111/go-idm/internal/dataaccess/kafka/producer"
	"github.com/manhhung2111/go-idm/internal/logic"
	"github.com/manhhung2111/go-idm/internal/utils"
	"go.uber.org/zap"
)

type DownloadTaskCreateHandler interface {
	Handle(ctx context.Context, event producer.DownloadTaskCreated) error
}

type downloadTaskCreatedHandler struct {
	downloadTaskLogic logic.DownloadTask
	logger *zap.Logger
}

func NewDownloadTaskCreatedHandler(
	downloadTaskLogic logic.DownloadTask,
	logger *zap.Logger,
) DownloadTaskCreateHandler {
	return &downloadTaskCreatedHandler{
		logger: logger,
		downloadTaskLogic: downloadTaskLogic,
	}
}

// Handle implements DownloadTaskCreateHandler.
func (d *downloadTaskCreatedHandler) Handle(ctx context.Context, event producer.DownloadTaskCreated) error {
	logger := utils.LoggerWithContext(ctx, d.logger).With(zap.Any("event", event))
	logger.Info("download task create event received")

	if err := d.downloadTaskLogic.ExecuteDownloadTask(ctx, event.Id); err != nil {
		logger.With(zap.Error(err)).Error("failed to handle download task created event")
		return err
	}

	return nil
}