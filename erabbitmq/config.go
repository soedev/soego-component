package erabbitmq

type config struct {
	Url       string                    `json:"url" toml:"url"`             //连接字符串 amqp://guest:guest@localhost:5672/
	Debug     bool                      `json:"debug" toml:"debug"`         // Debug
	Producers map[string]producerConfig `json:"producers" toml:"producers"` //生产者集合
	Consumers map[string]consumerConfig `json:"consumers" toml:"consumers"` //消费者集合
}

//生产者配置
type producerConfig struct {
	Type       string          `json:"type" toml:"type"`             //生产方式 queue(默认方式),exchange(交换机方式)
	Queue      queueDeclare    `json:"queue" toml:"queue"`           // type = queue 这个参数必须配置
	Exchange   exchangeDeclare `json:"exchange" toml:"exchange"`     // type = exchange 这个参数必须配置
	RoutingKey string          `json:"routingKey" toml:"routingKey"` //type = exchange 此参数生效
}

//接收者/消费者配置
type consumerConfig struct {
	Exchange   exchangeDeclare `json:"exchange" toml:"exchange"`     //交换机配置
	Queue      queueDeclare    `json:"queue" toml:"queue"`           //消息队列设置， 如果配置了交换机会默认生效一个 临时匿名队列（可以理解为配置了交换机 队列可以不配置）
	Qos        qos             `json:"qos" toml:"qos"`               //交换机配置了，此参数被忽略
	RoutingKey string          `json:"routingKey" toml:"routingKey"` //交换机配置了，此参数必须配置，（用于配置路由规则绑定在消息队列上）
}

//配置消费公平调度：多个消费者，谁有空闲就调度给谁，可以设置预取值为 1
type qos struct {
	Enable        bool `json:"enable" toml:"enable"`               //启用qos设置，qos 仅在消息队列方式（queue）下有效
	PrefetchCount int  `json:"prefetchCount" toml:"prefetchCount"` //预取计数
	PrefetchSize  int  `json:"prefetchSize" toml:"prefetchSize"`   //预取大小
	Global        bool `json:"global" toml:"global"`               //默认false
}

//队列配置信息
type queueDeclare struct {
	Name       string `json:"name" toml:"name"`
	Durable    bool   `json:"durable" toml:"durable"`       //消息是否序列化、宕机消息不会丢失默认true
	AutoDelete bool   `json:"autoDelete" toml:"autoDelete"` //未使用情况下自动删除消息队列
	Exclusive  bool   `json:"exclusive" toml:"exclusive"`   //没有使用连接，自动删除消息队列（作为临时匿名队列时候，自动变为true 会好一些）
	NoWait     bool   `json:"noWait" toml:"noWait"`         //true 服务器上已经声明队列时候，尝试从不同的连接修改现有队列（一般默认 false）
	//Args       amqp.Table `json:"args" toml:"args"`             //When the error return value is not nil, you can assume the queue could not be declared with these parameters, and the channel will be closed.
}

//交换机配置
type exchangeDeclare struct {
	Name       string `json:"name" toml:"name"`             //交换机名称（必填）
	Kind       string `json:"kind" toml:"kind"`             //（必填）direct, topic, headers和fanout
	Durable    bool   `json:"durable" toml:"durable"`       //默认 true
	AutoDelete bool   `json:"autoDelete" toml:"autoDelete"` //默认 false
	Internal   bool   `json:"internal" toml:"internal"`     //默认 false
	NoWait     bool   `json:"noWait" toml:"noWait"`         //默认 false
}

// DefaultConfig 返回默认配置
func DefaultConfig() *config {
	return &config{
		Debug: true,
	}
}
