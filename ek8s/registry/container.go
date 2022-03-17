package registry

import (
	"github.com/soedev/soego/core/econf"
	"github.com/soedev/soego/core/elog"
	"google.golang.org/grpc/resolver"

	"github.com/soedev/soego-component/ek8s"
)

type Container struct {
	config     *Config
	name       string
	logger     *elog.Component
	client     *ek8s.Component
	fallbackRB resolver.Builder
}

func DefaultContainer() *Container {
	return &Container{
		config: DefaultConfig(),
		logger: elog.EgoLogger.With(elog.FieldComponent(ek8s.PackageName)),
	}
}

func Load(key string) *Container {
	c := DefaultContainer()
	if err := econf.UnmarshalKey(key, &c.config); err != nil {
		c.logger.Panic("parse config error", elog.FieldErr(err), elog.FieldKey(key))
		return c
	}
	c.logger = c.logger.With(elog.FieldComponentName(key))
	c.name = key
	return c
}

// Build ...
func (c *Container) Build(options ...Option) *Component {
	for _, option := range options {
		option(c)
	}
	if c.client == nil {
		if c.config.OnFailHandle == "panic" {
			c.logger.Panic("client kubernetes nil", elog.FieldKey("use WithKubernetes method"))
		} else {
			c.logger.Error("client kubernetes nil", elog.FieldKey("use WithKubernetes method"))
		}
	}
	return newComponent(c.name, c.config, c.logger, c.client)
}
