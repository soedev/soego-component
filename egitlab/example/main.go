package main

import (
	"fmt"

	"github.com/soedev/soego"
	"github.com/soedev/soego/core/elog"
	"github.com/xanzy/go-gitlab"
	"go.uber.org/zap"

	"github.com/soedev/soego-component/egitlab"
)

func main() {
	err := soego.New().Invoker(invokerGitlab).Run()
	if err != nil {
		elog.Error("startup", elog.Any("err", err))
	}
}

func invokerGitlab() error {
	comp := egitlab.Load("gitlab").Build()
	client := comp.Client()
	user, _, err := client.Users.GetUser(11, gitlab.GetUsersOptions{})
	if err != nil {
		elog.Error("get user failed", zap.Error(err))
		return err
	}
	fmt.Printf("user:%v \n", user)
	return nil
}
