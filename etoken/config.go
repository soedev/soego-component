package etoken

type Option func(c *Container)

// PackageName ..
const PackageName = "component.etoken"

// config
type config struct {
	Iss            string // JWT的签发者(iss)
	Secret         string // JWT的加密密钥
	ExpireInterval int64  // JWT到期时间(xp)
	Prefix         string // token前缀名称，避免与数据库key冲突
}

// DefaultConfig ...
func DefaultConfig() *config {
	return &config{
		Iss:            "EGO TOKEN JWT",
		Secret:         "etokenK#xo",
		ExpireInterval: 24 * 3600,
		Prefix:         "/egotoken",
	}
}
