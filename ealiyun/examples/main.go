package main

import (
	"fmt"

	"github.com/soedev/soego"
	"github.com/soedev/soego/core/elog"

	"github.com/soedev/soego-component/ealiyun"
)

func main() {
	err := soego.New().Invoker(
		invoker,
	).Run()
	if err != nil {
		elog.Error("startup", elog.Any("err", err))
	}
}
func invoker() error {
	comp := ealiyun.Load("aliyun").Build()
	userName := "xxxx@xxxxxx.onaliyun.com"
	res, err := comp.CreateRamUser(ealiyun.SaveRamUserRequest{
		UserPrincipalName: userName,
		DisplayName:       "李四",
		MobilePhone:       "xxxxxxx",
		Email:             "xxxxx",
	})
	if err != nil {
		fmt.Println("createUser err:" + err.Error())
		return err
	}
	fmt.Printf("createUser res:%#v\n", res)
	fmt.Println("=============================================")
	res, err = comp.GetRamUser(userName)
	if err != nil {
		fmt.Println("createUser err:" + err.Error())
		return err
	}
	fmt.Printf("getUser res:%#v\n", res)
	return nil
}
