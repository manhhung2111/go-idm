package producer

import (
	"context"
	"fmt"

	"github.com/IBM/sarama"
	"github.com/manhhung2111/go-idm/internal/config"
	"github.com/manhhung2111/go-idm/internal/utils"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Client interface {
	Send(ctx context.Context, topic string, payload []byte) error
}

type client struct {
	saramaSyncProducer sarama.SyncProducer
	logger             *zap.Logger
}

func newSaramaConfig(kafkaConfig config.Kafka) *sarama.Config {
	saramaConfig := sarama.NewConfig()
	saramaConfig.Producer.Retry.Max = 1
	saramaConfig.Producer.RequiredAcks = sarama.WaitForAll
	saramaConfig.Producer.Return.Successes = true
	saramaConfig.ClientID = kafkaConfig.ClientId
	saramaConfig.Metadata.Full = true
	return saramaConfig
}

func NewClient(
	kafkaConfig config.Kafka,
	logger *zap.Logger,
) (Client, error) {
	address := kafkaConfig.Host + ":" + kafkaConfig.Port
	saramaSyncProducer, err := sarama.NewSyncProducer([]string{address}, newSaramaConfig(kafkaConfig))
	if err != nil {
		return nil, fmt.Errorf("failed to create sarama sync producer: %w", err)
	}

	return &client{
		saramaSyncProducer: saramaSyncProducer,
		logger:             logger,
	}, nil
}

func (c client) Send(ctx context.Context, topic string, payload []byte) error {
	logger := utils.LoggerWithContext(ctx, c.logger).
		With(zap.String("topic", topic)).
		With(zap.ByteString("payload", payload))

	if _, _, err := c.saramaSyncProducer.SendMessage(&sarama.ProducerMessage{
		Topic: topic,
		Value: sarama.ByteEncoder(payload),
	}); err != nil {
		logger.With(zap.Error(err)).Error("failed to produce message")
		return status.Errorf(codes.Internal, "failed to produce message: %+v", err)
	}

	return nil
}