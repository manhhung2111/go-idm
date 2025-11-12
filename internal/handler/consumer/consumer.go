package handler_consumer

import (
	"context"
	"encoding/json"

	"go.uber.org/zap"

	"github.com/manhhung2111/go-idm/internal/dataaccess/kafka/consumer"
	"github.com/manhhung2111/go-idm/internal/dataaccess/kafka/producer"
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
	r.consumer.RegisterHandler(
		producer.DownloadTaskCreatedTopic,
		func(ctx context.Context, queueName string, payload []byte) error {
			var event producer.DownloadTaskCreated
			if err := json.Unmarshal(payload, &event); err != nil {
				return err
			}

			return r.downloadTaskCreatedHandler.Handle(ctx, event)
		},
	)

	return r.consumer.Start(ctx)
}
