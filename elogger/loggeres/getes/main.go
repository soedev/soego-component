package main

import (
	"github.com/soedev/soego"
	"github.com/soedev/soego-component/elogger/loggeres"
	"github.com/soedev/soego/core/elog"
)

func init() {
	elog.Register(&loggeres.EsWriterBuilder{})
}

//  export EGO_DEBUG=true && go run main.go --config=config.toml
func main() {
	err := soego.New().Invoker(func() error {
		elog.EgoLogger.Info("hello world2222")
		return nil
	}).Run()
	if err != nil {
		panic(err)
	}

}
