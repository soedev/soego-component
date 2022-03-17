package egitlab

import (
	"fmt"

	"github.com/soedev/soego/core/econf"
	"github.com/soedev/soego/core/elog"
)

type Option func(container *Container)

type Container struct {
	Config *config
	name   string
	logger *elog.Component
}

func Load(key string) *Container {
	c := DefaultContainer()
	if err := econf.UnmarshalKey(key, &c.Config); err != nil {
		c.logger.Panic("parse config error", elog.FieldErr(err), elog.FieldKey(key))
		return c
	}
	fmt.Println(c.Config)
	c.logger = c.logger.With(elog.FieldComponentName(key))
	c.name = key
	return c
}

func (c *Container) Build(options ...Option) *Component {
	for _, option := range options {
		option(c)
	}
	return newComponent(c.Config, c.logger)
}

func DefaultContainer() *Container {
	return &Container{
		Config: DefaultConfig(),
		logger: elog.EgoLogger.With(elog.FieldComponent(packageName)),
	}
}
