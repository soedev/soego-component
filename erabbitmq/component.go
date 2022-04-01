package erabbitmq

import (
	"github.com/soedev/soego/core/elog"
	"github.com/streadway/amqp"
	"sync"
	"time"
)

const PackageName = "component.erabbitmq"

type Component struct {
	config    *config
	logger    *elog.Component
	client    *Client
	consumers map[string]*Consumer
	producers map[string]*Producer
	csMu      sync.RWMutex
	pMu       sync.RWMutex
	compName  string
}

//根据name 生成单个消费者
func (cmp *Component) Consumer(name string) *Consumer {
	cmp.csMu.RLock()

	if consumer, ok := cmp.consumers[name]; ok {
		cmp.csMu.RUnlock()
		return consumer
	}

	cmp.csMu.RUnlock()
	cmp.csMu.Lock()

	if consumer, ok := cmp.consumers[name]; ok {
		cmp.csMu.Unlock()
		return consumer
	}

	config, ok := cmp.config.Consumers[name]
	if !ok {
		cmp.csMu.Unlock()
		cmp.logger.Panic("simple config not exists", elog.String("name", name))
	}
	consumer := newConsume(cmp.client, config, cmp.logger)
	//simple.setProcessor(cmp.interceptorServerChain())
	cmp.consumers[name] = consumer
	cmp.csMu.Unlock()
	return cmp.consumers[name]
}

// Producer 生成单个生产者
func (cmp *Component) Producer(name string) *Producer {
	cmp.pMu.RLock()

	if producer, ok := cmp.producers[name]; ok {
		cmp.pMu.RUnlock()
		return producer
	}

	cmp.pMu.RUnlock()
	cmp.pMu.Lock()

	if producer, ok := cmp.producers[name]; ok {
		cmp.pMu.Unlock()
		return producer
	}

	config, ok := cmp.config.Producers[name]
	if !ok {
		cmp.logger.Panic("producer config not exists", elog.String("name", name))
	}
	producer := newProducer(cmp.client, config, cmp.logger)
	//producer.setProcessor(cmp.interceptorClientChain())
	cmp.producers[name] = producer
	cmp.pMu.Unlock()
	return cmp.producers[name]
}

//批量初始化消费者
func (cmp *Component) InitConsumers(msgHandle AckHandle) {
	if len(cmp.config.Consumers) == 0 {
		cmp.logger.Panic("consumes config len == 0")
	}
	for cs := range cmp.config.Consumers {
		consumer := cmp.Consumer(cs)
		if handle, needAck, err := consumer.HandMessage(false, false, false, nil); err == nil {
			go msgHandle(handle, needAck)
		}
	}
}

//根据消费者名 来发送消息
func (cmp *Component) SendMessageByName(name string, body []byte) {
	cmp.Producer(name).SendMessage(body)
}

//自定义发送接口
func (cmp *Component) PublishingByName(name string, publishing amqp.Publishing) {
	cmp.Producer(name).Publishing(publishing)
}

func (cmp *Component) Close() {
	time.Sleep(time.Second * 5)
	cmp.pMu.Lock()
	for ps := range cmp.producers {
		cmp.producers[ps].ch.Close()
	}

	time.Sleep(time.Second * 5)
	for cs := range cmp.consumers {
		cmp.producers[cs].ch.Close()
	}
	cmp.pMu.Unlock()

	time.Sleep(time.Second * 5)
	cmp.client.cc.Close()
}
