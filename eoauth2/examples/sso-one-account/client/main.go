package main

import (
	"github.com/soedev/soego"
	"github.com/soedev/soego-component/eoauth2/examples/sso-one-account/client/pkg/invoker"
	"github.com/soedev/soego-component/eoauth2/examples/sso-one-account/client/pkg/server"
	"github.com/soedev/soego/core/elog"
)

//  export EGO_DEBUG=true && go run main.go
func main() {
	err := soego.New().
		Invoker(invoker.Init).
		Serve(
			server.ServeHttp(),
		).Run()
	if err != nil {
		elog.Panic("startup", elog.FieldErr(err))
	}
}
