package main

import (
	"context"
	"fmt"

	"github.com/soedev/soego"
	"github.com/soedev/soego-component/egorm"
	"github.com/soedev/soego/core/elog"
	"github.com/soedev/soego/examples/helloworld"
	"github.com/soedev/soego/server"
	"github.com/soedev/soego/server/egrpc"
	"gorm.io/gorm"
)

var db *gorm.DB

type User struct {
	Id       int    `gorm:"not null" json:"id"`
	Nickname string `gorm:"not null" json:"name"`
}

func (User) TableName() string {
	return "user"
}

//  export EGO_DEBUG=true && go run main.go --config=config.toml
func main() {
	if err := soego.New().Invoker(func() error {
		db = egorm.Load("mysql.test").Build()
		return nil
	}).Serve(func() server.Server {
		server := egrpc.Load("server.grpc").Build()
		helloworld.RegisterGreeterServer(server.Server, &Greeter{server: server})
		return server
	}()).Run(); err != nil {
		elog.Panic("startup", elog.FieldErr(err))
	}
}

// Greeter ...
type Greeter struct {
	server *egrpc.Component
	helloworld.UnimplementedGreeterServer
}

// SayHello ...
func (g Greeter) SayHello(ctx context.Context, request *helloworld.HelloRequest) (*helloworld.HelloResponse, error) {
	var user User
	err := db.WithContext(ctx).Where("id = ?", 100).First(&user).Error
	if err != nil {
		return nil, fmt.Errorf("sql err: %w", err)
	}
	return &helloworld.HelloResponse{
		Message: "Hello EGO, I'm " + g.server.Address(),
	}, nil
}
