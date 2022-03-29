package erabbitmq

import (
	"github.com/soedev/soego/core/econf"
	"github.com/soedev/soego/core/elog"
)

type Option func(c *Container)

type Container struct {
	config *config
	name   string
	logger *elog.Component
}

// DefaultContainer 返回默认Container
func DefaultContainer() *Container {
	return &Container{
		config: DefaultConfig(),
		logger: elog.EgoLogger.With(elog.FieldComponent(PackageName)),
	}
}

// Load 载入配置，初始化Container
func Load(key string) *Container {
	c := DefaultContainer()
	if err := econf.UnmarshalKey(key, &c.config, econf.WithWeaklyTypedInput(true)); err != nil {
		c.logger.Panic("parse config error", elog.FieldErr(err), elog.FieldKey(key))
		return c
	}
	c.logger = c.logger.With(elog.FieldComponentName(key))
	c.name = key
	return c
}

// Build 构建Container
func (c *Container) Build(options ...Option) *Component {
	if c.config.Debug {
		options = append(options, WithInterceptor(debugInterceptor(c.name, c.config)))
	}
	if c.config.EnableMetricInterceptor {
		options = append(options, WithInterceptor(metricInterceptor(c.name, c.config)))
	}
	for _, option := range options {
		option(c)
	}
	cmp := &Component{
		config:    c.config,
		logger:    c.logger,
		client:    Connect(c),
		consumers: make(map[string]*Consumer),
		producers: make(map[string]*Producer),
		compName:  c.name,
	}
	c.logger.Info("dial rabbitmq server", elog.String("url", c.config.Url))
	return cmp
}
