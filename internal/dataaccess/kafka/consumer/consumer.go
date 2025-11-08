package consumer

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"github.com/IBM/sarama"
	"github.com/manhhung2111/go-idm/internal/config"
	"go.uber.org/zap"
)

type HandlerFunc func(ctx context.Context, topic string, payload []byte) error

type Consumer interface {
	RegisterHandler(topic string, handlerFunc HandlerFunc) error
	Start(ctx context.Context) error
}

type partitionConsumerAndHandlerFunc struct {
	topic         string
	partitionConsumer sarama.PartitionConsumer
	handlerFunc       HandlerFunc
}

type consumer struct {
	saramaConsumer                      sarama.Consumer
	partitionConsumerAndHandlerFuncList []partitionConsumerAndHandlerFunc
	logger                              *zap.Logger
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

func (c *consumer) RegisterHandler(topic string, handlerFunc HandlerFunc) error {
	partitionConsumer, err := c.saramaConsumer.ConsumePartition(topic, 0, sarama.OffsetOldest)
	if err != nil {
		return fmt.Errorf("failed to create sarama partition consumer: %w", err)
	}

	c.partitionConsumerAndHandlerFuncList = append(
		c.partitionConsumerAndHandlerFuncList,
		partitionConsumerAndHandlerFunc{
			topic:         topic,
			partitionConsumer: partitionConsumer,
			handlerFunc:       handlerFunc,
		})

	return nil
}

func (c consumer) Start(_ context.Context) error {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt)

	for i := range c.partitionConsumerAndHandlerFuncList {
		go func(i int) {
			topic := c.partitionConsumerAndHandlerFuncList[i].topic
			partitionConsumer := c.partitionConsumerAndHandlerFuncList[i].partitionConsumer
			handlerFunc := c.partitionConsumerAndHandlerFuncList[i].handlerFunc
			logger := c.logger.With(zap.String("topic", topic))

			for {
				select {
				case message := <-partitionConsumer.Messages():
					if err := handlerFunc(context.Background(), topic, message.Value); err != nil {
						logger.With(zap.Error(err)).Error("failed to handle message")
					}

				case <-signals:
					// exit the goroutine when a signal is received
					return
				}
			}
		}(i)
	}

	<-signals
	return nil
}