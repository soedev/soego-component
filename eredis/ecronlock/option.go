package ecronlock

import "github.com/soedev/soego-component/eredis"

//WithClient Required. 设置 redis client
func WithClient(client *eredis.Component) Option {
	return func(c *Container) {
		c.client = client
	}
}

//WithPrefix Optional. 设置 redis 锁的 Key 前缀
func WithPrefix(prefix string) Option {
	return func(c *Container) {
		c.config.Prefix = prefix
	}
}
