package main

import (
	"encoding/json"
	"fmt"
	"github.com/soedev/soego"
	"github.com/soedev/soego-component/erabbitmq"
	"github.com/soedev/soego/core/elog"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

func main() {
	err := soego.New().Invoker(
		initRabbitMq,
	).Run()
	if err != nil {
		elog.Error("startup", elog.Any("err", err))
	}
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)
	signal.Notify(sig, syscall.SIGTERM)
	<-sig
}

var rmqClient *erabbitmq.Component

//初始化emqtt
func initRabbitMq() error {
	rmqClient = erabbitmq.Load("rabbitmq").Build()
	handMessage(rmqClient.Consumer("c1"))
	sendMsg(rmqClient.Producer("p1"))
	return nil
}

func handMessage(consumer *erabbitmq.Consumer) {
	msgs, needAck, err := consumer.HandMessage(false, false, false, nil)
	if err != nil {
		elog.Error(err.Error())
	}
	go func() {
		for d := range msgs {
			elog.Info(fmt.Sprintf(" [x] %s", d.Body))
			if needAck {
				d.Ack(false)
			}
		}
	}()
}
func sendMsg(producer *erabbitmq.Producer) {
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		var count uint64
		t := time.NewTicker(time.Second * 5)
		for {
			select {
			case <-t.C:
				count += 1
				msg, err := json.Marshal(struct {
					Count uint64
				}{Count: count})
				if err != nil {
					panic(err)
				}
				// Publish will block so we run it in a goRoutine
				go func(msg []byte) {
					elog.Info(fmt.Sprintf("sending %s", string(msg)))
					if err := producer.SendMessage(msg); err != nil {
						elog.Error("send error：" + err.Error())
					}
				}(msg)
			}
		}
	}()
}
