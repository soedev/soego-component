package main

import (
	"github.com/davecgh/go-spew/spew"
	"github.com/soedev/soego"
	"github.com/soedev/soego-component/ek8s"
	"github.com/soedev/soego/core/elog"
)

func main() {
	if err := soego.New().Invoker(
		invokerGrpc,
	).Run(); err != nil {
		elog.Error("startup", elog.FieldErr(err))
	}
}

func invokerGrpc() error {
	obj := ek8s.Load("k8s").Build()
	list, err := obj.ListPodsByName("svc-oss")
	if err != nil {
		//panic(err)
	}
	spew.Dump(list)

	pods, err := obj.ListPods(ek8s.ListOptions{})
	if err != nil {
		panic(err)
	}
	spew.Dump(pods)
	return nil
}
