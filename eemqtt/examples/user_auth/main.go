package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/eclipse/paho.golang/paho"
	"github.com/soedev/soego"
	"github.com/soedev/soego-component/eemqtt"
	"github.com/soedev/soego/core/elog"
)

var emqClient *eemqtt.Component

//创建 emqtt 测试环境   docker run -d --name emqx -p 18083:18083 -p 1883:1883 emqx/emqx:latest
func main() {
	err := soego.New().Invoker(
		initEQ,
		pub,
	).Run()
	if err != nil {
		elog.Error("startup", elog.Any("err", err))
	}
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)
	signal.Notify(sig, syscall.SIGTERM)
	<-sig
}

//初始化emqtt
func initEQ() error {
	emqClient = eemqtt.Load("emqtt").Build()
	emqClient.Start(msgHandler)
	return nil
}

func pub() error {
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
				go func(message struct {
					Count uint64
				}) {
					bytes, _ := json.Marshal(message)
					emqClient.PublishMsg("z9_8b070ac409c920da", 1, bytes)
				}(struct {
					Count uint64
				}{Count: count})
			case <-emqClient.ServerCtx.Done():
				fmt.Println("publisher done")
				return
			}
		}
	}()
	return nil
}

//接收的消息处理端
func msgHandler(ctx context.Context, pp *paho.Publish) {
	elog.Info("receive meg", elog.Any("topic", pp.Topic), elog.Any("msg", string(pp.Payload)))
	fmt.Println(string(pp.Payload))
}

//1.完善docker测试
//2.编写收发代码
//3.提交 pr
