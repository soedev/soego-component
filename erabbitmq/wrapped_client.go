package erabbitmq

import (
	"fmt"
	"github.com/soedev/soego/core/elog"
	"github.com/streadway/amqp"
	"time"
)

const delay = 3

type Client struct {
	cc      *amqp.Connection
	logMode bool
}

func Connect(config *config, logger *elog.Component) (wc *Client) {
	wc = &Client{}
	wc.Connect(config.Url, logger)
	return wc
}

//建立连接
func (wc *Client) Connect(url string, logger *elog.Component) {
	conn, err := amqp.Dial(url)
	if err != nil {
		logger.Panic("connect fialed", elog.FieldErr(err))
	}
	wc.cc = conn
	//失败重连
	go func() {
		for {
			reason, ok := <-wc.cc.NotifyClose(make(chan *amqp.Error))
			// exit this goroutine if closed by developer
			if !ok {
				logger.Warn("connection closed")
				break
			}
			logger.Error(fmt.Sprintf("connection closed, reason: %v", reason))
			// reconnect if not closed by developer
			for {
				// wait 1s for reconnect
				time.Sleep(delay * time.Second)

				conn, err := amqp.Dial(url)
				if err == nil {
					wc.cc = conn
					logger.Info("reconnect success")
					break
				}
				logger.Error("reconnect failed", elog.FieldErr(err))
			}
		}
	}()
}
