package erabbitmq

import (
	"fmt"
	"github.com/soedev/soego/core/elog"
	"github.com/streadway/amqp"
	"log"
	"sync/atomic"
	"time"
)

// Consumer 消费者信息
type Consumer struct {
	config consumerConfig
	ch     *Channel
	q      amqp.Queue
}

func newConsume(client *Client, config consumerConfig, logger *elog.Component) *Consumer {
	ccCH, err := client.cc.Channel()
	if err != nil {
		logger.Panic("newConsume get channel failed", elog.FieldErr(err))
	}
	cs := &Consumer{
		config: config,
		ch: &Channel{
			Channel: ccCH,
		},
	}
	if err = cs.initConsumer(); err != nil {
		logger.Panic("initConsumer failed", elog.FieldErr(err))
	}
	logger.Info("consumer create success", elog.FieldValueAny(config))
	//支持通道重建
	go func() {
		for {
			reason, ok := <-cs.ch.Channel.NotifyClose(make(chan *amqp.Error))
			// exit this goroutine if closed by developer
			if !ok || cs.ch.IsClosed() {
				logger.Warn("channel channel closed")
				cs.ch.Close() // close again, ensure closed flag set when connection closed
				break
			}
			logger.Error(fmt.Sprintf("channel closed, reason: %v", reason))
			// reconnect if not closed by developer
			for {
				// wait 1s for connection reconnect
				time.Sleep(delay * time.Second)
				ch, err := client.cc.Channel()
				if err == nil {
					//重连后的处理： 如果通道是匿名队列的 重连后必须关掉以前的 channel 并重新初始化消息队列
					logger.Info("channel channel recreate success")
					if cs.config.Queue.Name == "" { //匿名队列
						cs.ch.Close()
						cs.ch = &Channel{
							Channel: ch,
						}
						if err = cs.initConsumer(); err != nil {
							logger.Error("recreate initConsumer failed", elog.FieldErr(err))
						}
					} else {
						cs.ch.Channel = ch
					}
					break
				}
				logger.Error("channel channel recreate failed", elog.FieldErr(err))
			}
		}

	}()
	return cs
}

//初始化生产者队列以及路由信息
func (r *Consumer) initConsumer() error {
	if r.config.Exchange.Name != "" {
		if r.config.Exchange.Kind == "" {
			r.config.Exchange.Kind = "direct"
		}
		if r.config.Exchange.Kind != "direct" && r.config.Exchange.Kind != "fanout" && r.config.Exchange.Kind != "topic" {
			return fmt.Errorf("exchange kind is error kind= %s", r.config.Exchange.Kind)
		}
		if err := r.ch.ExchangeDeclare(r.config.Exchange.Name, r.config.Exchange.Kind,
			r.config.Exchange.Durable, r.config.Exchange.AutoDelete, r.config.Exchange.Internal, r.config.Exchange.NoWait, nil); err != nil {
			return fmt.Errorf("exchange set error %s", err.Error())
		}
	} else {
		if r.config.Queue.Name == "" {
			return fmt.Errorf("failed queue/exchange not define")
		}
	}
	var err error
	r.q, err = r.ch.QueueDeclare(r.config.Queue.Name, r.config.Queue.Durable, r.config.Queue.AutoDelete, r.config.Queue.Exclusive, r.config.Queue.NoWait, nil)
	if err != nil {
		return fmt.Errorf("queue set error %s", err.Error())
	}
	if r.config.Qos.Enable {
		if err = r.ch.Qos(r.config.Qos.PrefetchCount, r.config.Qos.PrefetchSize, r.config.Qos.Global); err != nil {
			return fmt.Errorf("qos set error %s", err.Error())
		}
	}

	//bind
	if r.config.Exchange.Name != "" {
		if err = r.ch.QueueBind(r.q.Name, r.config.RoutingKey, r.config.Exchange.Name, false, nil); err != nil {
			return fmt.Errorf("exchange bind queue is error %s", err.Error())
		}
	}
	return nil
}

//对外的消息处理消息入口：  配置没有自动ack 在处理消息的时候一定要 ack(false)
func (r *Consumer) HandMessage(exclusive, noLocal, noWait bool, args amqp.Table) (<-chan amqp.Delivery, bool, error) {
	deliveries := make(chan amqp.Delivery)
	//恢复连接后 消息的处理也要支持自动回复
	go func() {
		for {
			d, err := r.ch.Consume(r.q.Name, "", r.config.AutoAck, exclusive, noLocal, noWait, args)
			if err != nil {
				log.Printf("consume failed, err: %v", err)
				time.Sleep(delay * time.Second)
				if r.ch.IsClosed() {
					break
				} else {
					continue
				}
			}

			for msg := range d {
				deliveries <- msg
			}

			// sleep before IsClose call. closed flag may not set before sleep.
			time.Sleep(delay * time.Second)

			if r.ch.IsClosed() {
				break
			}
		}
	}()
	return deliveries, !r.config.AutoAck, nil
}

func (r *Consumer) Close() {
	r.ch.Close()
}

// Channel amqp.Channel wapper
type Channel struct {
	*amqp.Channel
	closed int32
}

// IsClosed indicate closed by developer
func (ch *Channel) IsClosed() bool {
	return (atomic.LoadInt32(&ch.closed) == 1)
}

// Close ensure closed flag set
func (ch *Channel) Close() error {
	if ch.IsClosed() {
		return amqp.ErrClosed
	}

	atomic.StoreInt32(&ch.closed, 1)

	return ch.Channel.Close()
}
