package consumer

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"github.com/IBM/sarama"
	"github.com/manhhung2111/go-idm/internal/config"
	"github.com/manhhung2111/go-idm/internal/utils"
	"go.uber.org/zap"
)

type HandlerFunc func(ctx context.Context, topic string, payload []byte) error

type Consumer interface {
	RegisterHandler(topic string, handlerFunc HandlerFunc)
	Start(ctx context.Context) error
}

type partitionConsumerAndHandlerFunc struct {
	topic             string
	partitionConsumer sarama.PartitionConsumer
	handlerFunc       HandlerFunc
}

type consumer struct {
	saramaConsumer        sarama.Consumer
	topicToHandlerFuncMap map[string]HandlerFunc
	logger                *zap.Logger
}

func newSaramaConfig(kafkaConfig config.Kafka) *sarama.Config {
	saramaConfig := sarama.NewConfig()
	saramaConfig.ClientID = kafkaConfig.ClientId
	saramaConfig.Metadata.Full = true
	return saramaConfig
}

func NewConsumer(
	kafkaConfig config.Kafka,
	logger *zap.Logger,
) (Consumer, error) {
	address := kafkaConfig.Host + ":" + kafkaConfig.Port
	saramaConsumer, err := sarama.NewConsumer([]string{address}, newSaramaConfig(kafkaConfig))
	if err != nil {
		return nil, fmt.Errorf("failed to create sarama consumer: %w", err)
	}

	return &consumer{
		saramaConsumer: saramaConsumer,
		logger:         logger,
	}, nil
}

func (c *consumer) RegisterHandler(topic string, handlerFunc HandlerFunc) {
	c.topicToHandlerFuncMap[topic] = handlerFunc
}

func (c *consumer) consume(topic string, handlerFunc HandlerFunc, exitSignalChannel chan os.Signal) error {
	logger := c.logger.With(zap.String("topic", topic))

	partitionConsumer, err := c.saramaConsumer.ConsumePartition(topic, 0, sarama.OffsetOldest)
	if err != nil {
		return fmt.Errorf("failed to create sarama partition consumer: %w", err)
	}

	for {
		select {
		case message := <-partitionConsumer.Messages():
			err = handlerFunc(context.Background(), topic, message.Value)
			if err != nil {
				logger.With(zap.Error(err)).Error("failed to handle message")
			}

		case <-exitSignalChannel:
			break
		}
	}
}

func (c consumer) Start(ctx context.Context) error {
	logger := utils.LoggerWithContext(ctx, c.logger)

	exitSignalChannel := make(chan os.Signal, 1)
	signal.Notify(exitSignalChannel, os.Interrupt)

	for topic, handlerFunc := range c.topicToHandlerFuncMap {
		go func(topic string, handlerFunc HandlerFunc) {
			if err := c.consume(topic, handlerFunc, exitSignalChannel); err != nil {
				logger.
					With(zap.String("topic", topic)).
					With(zap.Error(err)).
					Error("failed to consume message from topic")
			}
		}(topic, handlerFunc)
	}

	<-exitSignalChannel
	return nil
}
