package edingtalk

import (
	"fmt"
	"github.com/soedev/soego-component/eredis"
	"github.com/soedev/soego/core/econf"
	"github.com/soedev/soego/core/elog"
)

type Option func(c *Container)

type Container struct {
	config *config
	name   string
	logger *elog.Component
}

func DefaultContainer() *Container {
	return &Container{
		config: DefaultConfig(),
		logger: elog.EgoLogger.With(elog.FieldComponent(PackageName)),
	}
}

func Load(key string) *Container {
	c := DefaultContainer()
	if err := econf.UnmarshalKey(key, &c.config); err != nil {
		c.logger.Panic("parse config error", elog.FieldErr(err), elog.FieldKey(key))
		return c
	}
	fmt.Println(c.config)
	c.logger = c.logger.With(elog.FieldComponentName(key))
	c.name = key
	return c
}

func WithERedis(redis *eredis.Component) Option {
	return func(c *Container) {
		c.config.eredis = redis
	}
}

// Build ...
func (c *Container) Build(options ...Option) *Component {
	for _, option := range options {
		option(c)
	}
	return newComponent(c.name, c.config, c.logger)
}
