package erabbitmq

func WithInterceptor(interceptors ...Interceptor) Option {
	return func(c *Container) {
		if c.config.interceptors == nil {
			c.config.interceptors = make([]Interceptor, 0)
		}
		c.config.interceptors = append(c.config.interceptors, interceptors...)
	}
}
