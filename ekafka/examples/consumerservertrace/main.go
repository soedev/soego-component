package main

import (
	"context"
	"fmt"
	"github.com/soedev/soego"
	"log"

	"github.com/segmentio/kafka-go"
	"github.com/soedev/soego-component/ekafka"
	"github.com/soedev/soego-component/ekafka/consumerserver"
	"github.com/soedev/soego/core/elog"
	"github.com/soedev/soego/server/egovernor"
)

var (
	ec *ekafka.Component
)

func main() {
	app := soego.New().Invoker(func() error {
		ec = ekafka.Load("kafka").Build()
		// 使用p1生产者生产消息
		produce(context.Background(), ec.Producer("p1"))
		return nil
	}).Serve(
		// 可以搭配其他服务模块一起使用
		egovernor.Load("server.governor").Build(),

		// 初始化 Consumer Server
		func() *consumerserver.Component {
			// 依赖 `ekafka` 管理 Kafka consumer
			ec = ekafka.Load("kafka").Build()
			cs := consumerserver.Load("kafkaConsumerServers.s1").Build(
				consumerserver.WithEkafka(ec),
			)

			// 用来接收、处理 `kafka-go` 和处理消息的回调产生的错误
			consumptionErrors := make(chan error)

			// 注册处理消息的回调函数
			cs.OnEachMessage(consumptionErrors, func(ctx context.Context, message kafka.Message) error {
				fmt.Printf("got a message: %s\n", string(message.Value))
				// 如果返回错误则会被转发给 `consumptionErrors`
				return nil
			})

			return cs
		}(),
		// 还可以启动多个 Consumer Server
	)
	if err := app.Run(); err != nil {
		elog.Panic("startup", elog.Any("err", err))
	}
}

// produce 生产消息
func produce(ctx context.Context, w *ekafka.Producer) {
	// 生产3条消息
	ctx = context.WithValue(ctx, "hello", "world")
	err := w.WriteMessages(ctx,
		&ekafka.Message{Key: []byte("Key-A"), Value: []byte("Hellohahah World!22222")},
	)
	if err != nil {
		log.Fatal("failed to write messages:", err)
	}
	if err := w.Close(); err != nil {
		log.Fatal("failed to close writer:", err)
	}
}
