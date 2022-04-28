package tenancy

import "time"

//Config 定义标准的数据库连接可配置参数
type Config struct {
	Dialect         string        // 数据库种类: mysql,postgres,mssql
	DSN             string        // DSN地址
	Debug           bool          // 是否开启调试
	MaxIdleConns    int           // 最大空闲连接数
	MaxOpenConns    int           // 最大活动连接数
	ConnMaxLifetime time.Duration // 连接的最大存活时间
}
