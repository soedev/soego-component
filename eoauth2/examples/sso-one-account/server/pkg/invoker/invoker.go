package invoker

import (
	"github.com/soedev/soego-component/egorm"
	ssoserver "github.com/soedev/soego-component/eoauth2/server"
	"github.com/soedev/soego-component/eoauth2/storage/ssostorage"
	"github.com/soedev/soego-component/eredis"
)

var (
	SsoComponent *ssoserver.Component
	TokenStorage *ssostorage.Component
	Db           *egorm.Component
)

func Init() error {
	Db = egorm.Load("mysql").Build()
	Redis := eredis.Load("redis").Build()
	TokenStorage = ssostorage.NewComponent(
		Db,
		Redis,
	)
	SsoComponent = ssoserver.Load("sso").Build(ssoserver.WithStorage(TokenStorage.GetStorage()))
	return nil
}
