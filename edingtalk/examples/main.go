package main

import (
	"fmt"

	"github.com/soedev/soego"
	"github.com/soedev/soego-component/eredis"
	"github.com/soedev/soego/core/elog"

	"github.com/soedev/soego-component/edingtalk"
)

// export EGO_DEBUG=true && go run main.go --config=config.toml
func main() {
	err := soego.New().Invoker(
		invokerDingTalk,
	).Run()
	if err != nil {
		elog.Error("startup", elog.Any("err", err))
	}
}

func invokerDingTalk() error {
	redis := eredis.Load("redis").Build(eredis.WithStub())
	comp := edingtalk.Load("dingtalk").Build(edingtalk.WithERedis(redis))
	user, err := comp.GetUserInfo("5a84b3af502834d4a663d33378263b66")
	fmt.Println(user)
	fmt.Println(err)
	fmt.Println("==================================")
	err = comp.DepartmentUpdate(edingtalk.NewDepartmentUpdateReq(11111).SetDeptManagerUseridList("xxxxx"))
	fmt.Println("err", err)
	fmt.Println("==================================")
	link := &edingtalk.Link{
		PicURL:     "xxxxx",
		MessageURL: "xxx", Text: "xxx", Title: "xx",
	}
	msg := &edingtalk.Msg{
		Msgtype: edingtalk.MsgLink,
		Link:    link,
	}
	res, err := comp.CorpconversationAsyncsendV2(edingtalk.CorpconversationAsyncsendV2Req{
		Msg:        msg,
		UseridList: "xxx,xxxx",
	})
	fmt.Println(res, err)
	return nil
}
