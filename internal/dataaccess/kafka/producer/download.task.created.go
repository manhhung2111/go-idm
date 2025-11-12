package producer

import (
	"context"
	"encoding/json"

	"github.com/manhhung2111/go-idm/internal/utils"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	DownloadTaskCreatedTopic = "download.task.created"
)

type DownloadTaskCreated struct {
	Id uint64 `json:"id"`
}

type DownloadTaskCreatedProducer interface {
	Send(ctx context.Context, event DownloadTaskCreated) error
}

type downloadTaskCreatedProducer struct {
	client Client
	logger *zap.Logger
}

func NewDownloadTaskCreatedProducer(
	client Client,
	logger *zap.Logger,
) DownloadTaskCreatedProducer {
	return &downloadTaskCreatedProducer{
		client: client,
		logger: logger,
	}
}

// Send implements DownloadTaskCreatedProducer.
func (d *downloadTaskCreatedProducer) Send(ctx context.Context, event DownloadTaskCreated) error {
	logger := utils.LoggerWithContext(ctx, d.logger)

	eventBytes, err := json.Marshal(event)
	if err != nil {
		logger.With(zap.Error(err)).Error("failed to marshal download task created event")
		return status.Errorf(codes.Internal, "failed to marshal download task created event: %+v", err)
	}

	err = d.client.Send(ctx, DownloadTaskCreatedTopic, eventBytes)
	if err != nil {
		logger.With(zap.Error(err)).Error("failed to send download task created event")
		return status.Errorf(codes.Internal, "failed to send download task created event: %+v", err)
	}

	return nil
}
