# eredis 组件使用指南
## 简介 

对 [go-redis](https://github.com/go-redis/redis) 进行了轻量封装，并提供了以下功能：

- 规范了标准配置格式，提供了统一的 Load().Build() 方法。
- 支持自定义拦截器
- 提供了默认的 Debug 拦截器，开启 Debug 后可输出 Request、Response 至终端。
- 提供了默认的 Metric 拦截器，开启后可采集 Prometheus 指标数据
- 提供了redis的分布式锁
- 提供了redis的分布式锁的定时任务

## 快速上手

使用样例可参考 [examples](./examples/redis/main.go)
使用样例可参考 [examples](./examples/redislockcron/main.go)

