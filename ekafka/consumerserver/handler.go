package consumerserver

import (
	"context"
	"github.com/segmentio/kafka-go"
	"github.com/soedev/soego-component/ekafka"
)

// OnEachMessageHandler ...
type OnEachMessageHandler = func(ctx context.Context, message kafka.Message) error

// OnStartHandler ...
type OnStartHandler = func(ctx context.Context, consumer *ekafka.Consumer) error

// OnConsumerGroupStartHandler ...
type OnConsumerGroupStartHandler = func(ctx context.Context, consumerGroup *ekafka.ConsumerGroup) error
