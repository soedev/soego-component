package main

import (
	"context"

	"github.com/soedev/soego"
	"github.com/soedev/soego/client/egrpc"
	"github.com/soedev/soego/core/elog"

	"github.com/soedev/soego-component/eetcd"
	"github.com/soedev/soego-component/eetcd/examples/helloworld"
	"github.com/soedev/soego-component/eetcd/registry"
)

func main() {
	if err := soego.New().Invoker(
		invokerGrpc,
		callGrpc,
	).Run(); err != nil {
		elog.Error("startup", elog.FieldErr(err))
	}
}

var grpcComp helloworld.GreeterClient

func invokerGrpc() error {
	// 注册resolver
	registry.Load("registry").Build(registry.WithClientEtcd(eetcd.Load("etcd").Build()))
	grpcConn := egrpc.Load("grpc.test").Build()
	grpcComp = helloworld.NewGreeterClient(grpcConn.ClientConn)
	return nil
}

func callGrpc() error {
	_, err := grpcComp.SayHello(context.Background(), &helloworld.HelloRequest{
		Name: "i am client",
	})
	if err != nil {
		return err
	}

	_, err = grpcComp.SayHello(context.Background(), &helloworld.HelloRequest{
		Name: "error",
	})
	if err != nil {
		return err
	}
	return nil
}
