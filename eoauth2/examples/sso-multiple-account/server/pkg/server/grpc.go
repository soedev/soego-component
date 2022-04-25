package server

import (
	"context"
	"fmt"

	ssov1 "github.com/soedev/soego-component/eoauth2/examples/sso-multiple-account/proto"
	"github.com/soedev/soego-component/eoauth2/examples/sso-multiple-account/server/pkg/invoker"
	"github.com/soedev/soego-component/eoauth2/server"
	"github.com/soedev/soego/server/egrpc"
	"github.com/spf13/cast"
)

func ServeGrpc() *egrpc.Component {
	component := egrpc.Load("server.grpc").Build()
	ssov1.RegisterSsoServer(component.Server, &SsoGrpc{})
	return component
}

// SsoGrpc ...
type SsoGrpc struct {
	ssov1.UnimplementedSsoServer
}

func (SsoGrpc) GetToken(ctx context.Context, req *ssov1.GetTokenRequest) (*ssov1.GetTokenResponse, error) {
	ar := invoker.SsoComponent.HandleAccessRequest(ctx, server.ParamAccessRequest{
		Method:    "POST",
		GrantType: string(server.AUTHORIZATION_CODE),
		AccessRequestParam: server.AccessRequestParam{
			Code: req.Code,
			ClientAuthParam: server.ClientAuthParam{
				Authorization: req.Authorization,
			},
		},
	})
	err := ar.Build(server.WithAccessRequestAuthorized(true))
	if err != nil {
		return nil, fmt.Errorf("GetToken error, %w", err)
	}
	return &ssov1.GetTokenResponse{
		Token:     ar.GetOutput("access_token").(string),
		ExpiresIn: int64(ar.GetOutput("expires_in").(int32)),
	}, nil
}
func (SsoGrpc) RefreshToken(ctx context.Context, req *ssov1.RefreshTokenRequest) (resp *ssov1.RefreshTokenResponse, err error) {
	// todo 这里不是每次请求都刷新，根据过期时间，自动判断去做刷新。强制刷新接口后面也可以提供
	// 访问该接口的时间，通常大于1/2过期时间，我们才会触发refresh token操作
	// 例如token需要14天过期，那么这里会判断时间到7天之后，才会触发token
	// 换token操作，需要注意并发问题
	ar := invoker.SsoComponent.HandleAccessRequest(ctx, server.ParamAccessRequest{
		Method:    "POST",
		GrantType: string(server.REFRESH_TOKEN),
		AccessRequestParam: server.AccessRequestParam{
			Code: req.Code,
			ClientAuthParam: server.ClientAuthParam{
				Authorization: req.Authorization,
			},
		},
	})

	err = ar.Build(server.WithAccessRequestAuthorized(true))
	if err != nil {
		return nil, fmt.Errorf("RefreshToken error, %w", err)
	}
	resp = &ssov1.RefreshTokenResponse{
		Token:     ar.GetOutput("access_token").(string),
		ExpiresIn: int64(ar.GetOutput("expires_in").(int32)),
	}
	return resp, nil
}
func (SsoGrpc) RemoveToken(ctx context.Context, req *ssov1.RemoveTokenRequest) (*ssov1.RemoveTokenResponse, error) {
	resp := &ssov1.RemoveTokenResponse{}
	err := invoker.TokenStorage.RemoveAllAccess(ctx, req.Token)
	return resp, err
}
func (SsoGrpc) GetUsersByToken(ctx context.Context, req *ssov1.GetUsersByTokenRequest) (*ssov1.GetUsersByTokenResponse, error) {
	uids, err := invoker.TokenStorage.GetUidsByToken(ctx, req.Token)
	if err != nil {
		return nil, err
	}
	output := make([]*ssov1.User, 0)
	for _, uid := range uids {
		output = append(output, &ssov1.User{
			Uid:      uid,
			Nickname: "askuy" + cast.ToString(uid),
		})
	}
	return &ssov1.GetUsersByTokenResponse{
		User: output,
	}, nil
}
