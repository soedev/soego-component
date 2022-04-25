package main

import (
	"github.com/soedev/soego"
	"github.com/soedev/soego-component/eoauth2/examples/sso-multiple-account/server/job"
	"github.com/soedev/soego-component/eoauth2/examples/sso-multiple-account/server/pkg/invoker"
	"github.com/soedev/soego-component/eoauth2/examples/sso-multiple-account/server/pkg/server"
	"github.com/soedev/soego/core/elog"
	"github.com/soedev/soego/task/ejob"
)

//  export EGO_DEBUG=true && go run main.go
func main() {
	err := soego.New().
		Invoker(invoker.Init).
		Job(ejob.Job("init_data", job.InitAdminData)).
		Serve(
			server.ServeHttp(),
			server.ServeGrpc(),
		).Run()
	if err != nil {
		elog.Panic("startup", elog.FieldErr(err))
	}
}
