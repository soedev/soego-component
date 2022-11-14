# egorm 组件使用指南

## 1 简介
针对jwt做了简单封装

## 2 使用方式
```bash
go get github.com/soedev/soego-component/etoken
```

## 3 配置
```go
type config struct {
    AccessTokenIss            string // JWT的签发者(iss)
    AccessTokenKey            string // JWT的加密密钥
    AccessTokenExpireInterval int64  // JWT到期时间(xp)
    TokenPrefix               string // token前缀名称，避免与数据库key冲突
}
```

## 4 用户配置

```toml
[token.test]
iss = "" # JWT的签发者(iss)
secret = "" # WT的加密密钥
expireInterval = "3600" # JWT过期时间(exp)
prefix = "/egotoken" # token前缀名称，避免与数据库key冲突
```

## 5 用户代码

配置创建一个 ``etoken`` 的配置项，其中内容按照上文配置进行填写。以上这个示例里这个配置key是``token.test``

代码中创建一个 ``etoken`` 实例 ``etoken.Load("key").Build()``，代码中的 ``key`` 和配置中的 ``key`` 要保持一致。创建完 ``gorm`` 实例后，就可以直接使用他对 ``db``
进行 ``crud`` 。

```go
etoken.Load("token").Build(etoken.WithRedis(RedisStub))
etoken.CreateAccessToken(context.Background(),time.Now().Unix())
```