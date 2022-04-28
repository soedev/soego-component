package tenancy

import (
	"context"
)

type Storage interface {
	//Config 用户需要实现租户逻辑
	Config(ctx context.Context, tenancyKey string) (*Config, error)
}
