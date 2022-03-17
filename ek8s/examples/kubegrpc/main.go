package main

import (
	"fmt"

	"github.com/soedev/soego"
	"github.com/soedev/soego/client/egrpc"
	"github.com/soedev/soego/core/elog"

	"github.com/soedev/soego-component/ek8s"
	"github.com/soedev/soego-component/ek8s/examples/kubegrpc/helloworld"
	"github.com/soedev/soego-component/ek8s/registry"
)

func main() {
	if err := soego.New().Invoker(
		invokerGrpc,
	).Run(); err != nil {
		elog.Error("startup", elog.FieldErr(err))
	}
}

func invokerGrpc() error {
	// 构建k8s registry，并注册为grpc resolver
	registry.Load("registry").Build(
		registry.WithClient(ek8s.Load("k8s").Build()),
	)
	// 构建gRPC.ClientConn组件
	grpcConn := egrpc.Load("grpc.test").Build()
	// 构建gRPC Client组件
	grpcComp := helloworld.NewGreeterClient(grpcConn.ClientConn)
	fmt.Printf("client--------------->"+"%+v\n", grpcComp)
	return nil
}
