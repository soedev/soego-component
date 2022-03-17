package ecronlock

import (
	"sync"

	"github.com/soedev/soego/core/elog"
	"github.com/soedev/soego/task/ecron"

	"github.com/soedev/soego-component/eredis"
)

type Component struct {
	name   string
	config *Config
	logger *elog.Component
	client *eredis.Component
	mutuex sync.RWMutex
}

func newComponent(name string, config *Config, logger *elog.Component, client *eredis.Component) *Component {
	reg := &Component{
		name:   name,
		logger: logger,
		config: config,
		client: client,
	}
	return reg
}

func (c *Component) NewLock(key string) ecron.Lock {
	return newRedisLock(c.client, c.config.Prefix+key, c.logger)
}
