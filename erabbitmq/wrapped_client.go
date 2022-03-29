package erabbitmq

import (
	"fmt"
	"github.com/soedev/soego/core/elog"
	"github.com/streadway/amqp"
	"time"
)

type processor func(fn processFn) error
type processFn func(*cmd) error

type cmd struct {
	name string
	req  []interface{}
	res  interface{}
}

func logCmd(logMode bool, c *cmd, name string, res interface{}, req ...interface{}) {
	// 只有开启log模式才会记录req、res
	if logMode {
		c.name = name
		c.req = append(c.req, req...)
		c.res = res
	}
}

const delay = 3

type Client struct {
	cc        *amqp.Connection
	processor processor
	logMode   bool
}

func Connect(c *Container) (wc *Client) {
	wc = &Client{logMode: c.config.Debug}
	wc.wrapProcessor(InterceptorChain(c.config.interceptors...))
	wc.Connect(c)
	return wc
}

func (wc *Client) wrapProcessor(wrapFn func(processFn) processFn) {
	wc.processor = func(fn processFn) error {
		return wrapFn(fn)(&cmd{req: make([]interface{}, 0, 1)})
	}
}

//建立连接
func (wc *Client) Connect(container *Container) {
	var err error
	err = wc.processor(func(c *cmd) error {
		logCmd(wc.logMode, c, "Connect", nil)
		wc.cc, err = amqp.Dial(container.config.Url)
		return err
	})
	if err != nil {
		container.logger.Panic("connect fialed", elog.FieldErr(err))
	}
	//失败重连
	go func() {
		for {
			reason, ok := <-wc.cc.NotifyClose(make(chan *amqp.Error))
			if !ok {
				container.logger.Warn("connection closed")
				break
			}
			container.logger.Error(fmt.Sprintf("connection closed, reason: %v", reason))
			for {
				time.Sleep(delay * time.Second)
				err = wc.processor(func(c *cmd) error {
					logCmd(wc.logMode, c, "Connect", nil)
					conn, err := amqp.Dial(container.config.Url)
					if err == nil {
						wc.cc = conn
					}
					return err
				})
				if err == nil {
					container.logger.Info("reconnect success")
					break
				}
				container.logger.Error("reconnect failed", elog.FieldErr(err))
			}
		}
	}()
}
