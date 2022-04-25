package job

import (
	"github.com/soedev/soego-component/eoauth2/examples/sso-one-account/server/pkg/invoker"
	"github.com/soedev/soego-component/eoauth2/storage/dao"
	"github.com/soedev/soego/core/econf"
	"github.com/soedev/soego/task/ejob"
)

func InitAdminData(ctx ejob.Context) (err error) {
	//models := []interface{}{
	//	&dao.App{},
	//}
	//gormdb := invoker.Db
	//err = gormdb.Set("gorm:table_options", "ENGINE=InnoDB").AutoMigrate(models...)
	//if err != nil {
	//	return err
	//}
	//fmt.Println("create table ok")
	err = invoker.TokenStorage.GetAPI().CreateClient(ctx.Ctx, &dao.App{
		ClientId:    "1234",
		Name:        "sso-client",
		Secret:      "5678",
		RedirectUri: econf.GetString("client.codeUrl"),
		Status:      1,
	})
	if err != nil {
		return
	}
	return nil
}
