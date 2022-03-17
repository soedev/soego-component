package consumerserver

import (
	"github.com/soedev/soego-component/ekafka"
)

// WithEkafka ...
func WithEkafka(ekafkaComponent *ekafka.Component) Option {
	return func(c *Container) {
		c.config.ekafkaComponent = ekafkaComponent
	}
}

// WithDebug enables debug mode.
func WithDebug(debug bool) Option {
	return func(c *Container) {
		c.config.Debug = debug
	}
}
