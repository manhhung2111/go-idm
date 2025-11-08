package kafka

import (
	"github.com/google/wire"
	"github.com/manhhung2111/go-idm/internal/dataaccess/kafka/consumer"
	"github.com/manhhung2111/go-idm/internal/dataaccess/kafka/producer"
)

var WireSet = wire.NewSet(
	consumer.WireSet,
	producer.WireSet,
)