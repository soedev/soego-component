package main

import (
	"context"

	"github.com/soedev/soego"
	"github.com/soedev/soego-component/eetcd"
	"github.com/soedev/soego-component/eetcd/examples/helloworld"
	"github.com/soedev/soego-component/eetcd/registry"
	"github.com/soedev/soego/core/elog"
	"github.com/soedev/soego/server"
	"github.com/soedev/soego/server/egrpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

//  export EGO_DEBUG=true && go run main.go --config=config.toml
func main() {
	if err := soego.New().
		Registry(registry.Load("registry").Build(registry.WithClientEtcd(eetcd.Load("etcd").Build()))).
		Serve(func() server.Server {
			server := egrpc.Load("server.grpc").Build()
			helloworld.RegisterGreeterServer(server.Server, &Greeter{server: server})
			return server
		}()).Run(); err != nil {
		elog.Panic("startup", elog.Any("err", err))
	}
}

type Greeter struct {
	server *egrpc.Component
}

func (g Greeter) SayHello(context context.Context, request *helloworld.HelloRequest) (*helloworld.HelloReply, error) {
	if request.Name == "error" {
		return nil, status.Error(codes.Unavailable, "error")
	}

	return &helloworld.HelloReply{
		Message: "Hello EGO, I'm " + g.server.Address(),
	}, nil
}
