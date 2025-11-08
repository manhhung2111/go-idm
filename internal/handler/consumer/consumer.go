package handler_consumer

import (
	"context"
	"encoding/json"
	"fmt"

	"go.uber.org/zap"

	"github.com/manhhung2111/go-idm/internal/dataaccess/kafka/consumer"
	"github.com/manhhung2111/go-idm/internal/dataaccess/kafka/producer"
	"github.com/manhhung2111/go-idm/internal/utils"
)

type Root interface {
	Start(ctx context.Context) error
}

type root struct {
	downloadTaskCreatedHandler DownloadTaskCreateHandler
	consumer                   consumer.Consumer
	logger                     *zap.Logger
}

func NewRoot(
	downloadTaskCreatedHandler DownloadTaskCreateHandler,
	consumer consumer.Consumer,
	logger *zap.Logger,
) Root {
	return &root{
		downloadTaskCreatedHandler: downloadTaskCreatedHandler,
		consumer:                   consumer,
		logger:                     logger,
	}
}

func (r root) Start(ctx context.Context) error {
	logger := utils.LoggerWithContext(ctx, r.logger)

	if err := r.consumer.RegisterHandler(
		producer.DownloadTaskCreatedTopic,
		func(ctx context.Context, queueName string, payload []byte) error {
			var event producer.DownloadTaskCreated
			if err := json.Unmarshal(payload, &event); err != nil {
				return err
			}

			return r.downloadTaskCreatedHandler.Handle(ctx, event)
		}); err != nil {
		logger.With(zap.Error(err)).Error("failed to register download task created handler")
		return fmt.Errorf("failed to register download task created handler: %w", err)
	}

	return r.consumer.Start(ctx)
}
