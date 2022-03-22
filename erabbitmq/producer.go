package erabbitmq

import (
	"fmt"
	"github.com/soedev/soego/core/elog"
	"github.com/streadway/amqp"
	"time"
)

const (
	producerTypeQueue    = "queue"
	producerTypeExchange = "exchange"
)

// Producer 生产者信息
type Producer struct {
	ch     *Channel
	config producerConfig
	q      amqp.Queue
}

func newProducer(client *Client, config producerConfig, logger *elog.Component) *Producer {
	ccCH, err := client.cc.Channel()
	if err != nil {
		logger.Panic("newProducer get channel failed", elog.FieldErr(err))
	}
	producer := &Producer{
		config: config,
		ch: &Channel{
			Channel: ccCH,
		},
	}
	if err = producer.initProducer(); err != nil {
		logger.Panic("initConsume failed", elog.FieldErr(err))
	}
	//支持通道重建
	logger.Info("producer create success", elog.FieldValueAny(config))
	go func() {
		for {
			reason, ok := <-producer.ch.Channel.NotifyClose(make(chan *amqp.Error))
			if !ok || producer.ch.IsClosed() {
				logger.Warn("producer channel closed")
				producer.ch.Close()
				break
			}
			logger.Error(fmt.Sprintf("producer channel closed, reason: %v", reason))
			for {
				time.Sleep(delay * time.Second)
				ch, err := client.cc.Channel()
				if err == nil {
					//重连后的处理： 如果通道是匿名队列的 重连后必须关掉以前的 channel 并重新初始化消息队列
					logger.Info("producer channel recreate success")
					producer.ch.Channel = ch
					break
				}
				logger.Error("producer channel recreate failed", elog.FieldErr(err))
			}
		}

	}()
	return producer
}
func (p *Producer) initProducer() error {
	if p.config.Type == "" {
		p.config.Type = producerTypeQueue
	}
	if p.config.Type == producerTypeQueue {
		if p.config.Queue.Name == "" {
			return fmt.Errorf("producer config queue not found")
		}
		var err error
		p.q, err = p.ch.QueueDeclare(p.config.Queue.Name, p.config.Queue.Durable, p.config.Queue.AutoDelete, p.config.Queue.Exclusive, p.config.Queue.NoWait, nil)
		if err != nil {
			return err
		}
	} else if p.config.Type == producerTypeExchange {
		if p.config.Exchange.Name == "" || p.config.Exchange.Kind == "" {
			return fmt.Errorf("producer config exchange set error")
		}
		if err := p.ch.ExchangeDeclare(p.config.Exchange.Name, p.config.Exchange.Kind, p.config.Exchange.Durable, p.config.Exchange.AutoDelete, p.config.Exchange.Internal,
			p.config.Exchange.NoWait, nil); err != nil {
			return err
		}
	} else {
		return fmt.Errorf("producer config Type is error")
	}
	return nil
}

func (p *Producer) SendMessage(body []byte) error {
	if p.config.Type == producerTypeQueue {
		return p.ch.Publish("", p.q.Name, false, false,
			amqp.Publishing{
				DeliveryMode: amqp.Persistent, //将消息标记为持久性 - 通过设置amqp.Publishing的amqp.Persistent
				ContentType:  "text/plain",
				Body:         body,
			})
	} else {
		return p.ch.Publish(p.config.Exchange.Name, p.config.RoutingKey, false, false,
			amqp.Publishing{
				DeliveryMode: amqp.Persistent,
				ContentType:  "text/plain",
				Body:         body,
			})
	}
}

func (p *Producer) Close() error {
	return p.ch.Close()
	return nil
}
