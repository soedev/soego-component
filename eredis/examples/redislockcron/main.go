package main

import (
	"context"
	"log"

	"github.com/soedev/soego"
	"github.com/soedev/soego/core/elog"
	"github.com/soedev/soego/task/ecron"

	"github.com/soedev/soego-component/eredis"
	"github.com/soedev/soego-component/eredis/ecronlock"
)

var (
	redis *eredis.Component
)

// export EGO_DEBUG=true && go run main.go --config=config.toml
func main() {
	err := soego.New().Invoker(initRedis).Cron(cronJob()).Run()
	if err != nil {
		elog.Panic("startup", elog.FieldErr(err))
	}
}

func initRedis() error {
	redis = eredis.Load("redis.test").Build()
	return nil
}

func cronJob() ecron.Ecron {
	locker := ecronlock.DefaultContainer().Build(ecronlock.WithClient(redis))
	cron := ecron.Load("cron.default").Build(
		ecron.WithLock(locker.NewLock("soedev/soego-component:cronjob:syncXxx")),
		ecron.WithJob(helloWorld),
	)
	return cron
}

func helloWorld(ctx context.Context) error {
	log.Println("cron job running")
	return nil
}
