package egorm

import (
	"github.com/soedev/soego-component/egorm/manager"
	"time"
)

// Option 可选项
type Option func(c *Container)

//WithDialect dialect
func WithDialect(dialect string) Option {
	return func(c *Container) {
		c.config.Dialect = dialect
	}
}

func WithMaxIdleConns(maxIdleConns int) Option {
	return func(c *Container) {
		c.config.MaxIdleConns = maxIdleConns
	}
}

func WithMaxOpenConns(maxOpenConns int) Option {
	return func(c *Container) {
		c.config.MaxOpenConns = maxOpenConns
	}
}

func WithConnMaxLifetime(connMaxLifetime time.Duration) Option {
	return func(c *Container) {
		c.config.ConnMaxLifetime = connMaxLifetime
	}
}

//WithDebug 调试模式
func WithDebug(debug bool) Option {
	return func(c *Container) {
		c.config.Debug = debug
	}
}

// WithDSN 设置dsn
func WithDSN(dsn string) Option {
	return func(c *Container) {
		c.config.DSN = dsn
	}
}

// WithDSNParser 设置自定义dsnParser
func WithDSNParser(parser manager.DSNParser) Option {
	return func(c *Container) {
		c.dsnParser = parser
	}
}

// WithInterceptor 设置自定义拦截器
func WithInterceptor(is ...Interceptor) Option {
	return func(c *Container) {
		if c.config.interceptors == nil {
			c.config.interceptors = make([]Interceptor, 0)
		}
		c.config.interceptors = append(c.config.interceptors, is...)
	}
}
