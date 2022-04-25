package etoken

import (
	"context"
	"fmt"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"github.com/soedev/soego/core/elog"
)

const tokenKeyPattern = "/token/%d"

type Component struct {
	config *config
	client *redis.Client
	logger *elog.Component
}

func newComponent(cfg *config, client *redis.Client, logger *elog.Component) *Component {
	return &Component{
		config: cfg,
		client: client,
		logger: logger,
	}
}

type AccessTokenTicket struct {
	AccessToken string `json:"accessToken"`
	ExpiresIn   int64  `json:"expiresIn"`
}

func (c *Component) CreateAccessToken(ctx context.Context, uid int, startTime int64) (resp AccessTokenTicket, err error) {
	tokenString, err := c.EncodeAccessToken(uuid.New().String(), uid, startTime)
	if err != nil {
		err = fmt.Errorf("CreateAccessToken EncodeAccessToken failed, err: %w", err)
		return
	}

	err = c.client.Set(ctx, fmt.Sprintf(c.config.Prefix+tokenKeyPattern, uid), tokenString,
		time.Duration(c.config.ExpireInterval)*time.Second).Err()
	if err != nil {
		return AccessTokenTicket{}, fmt.Errorf("CreateAccessToken Set Token  failed, err: %w", err)
	}
	resp.AccessToken = tokenString
	resp.ExpiresIn = c.config.ExpireInterval
	return
}

func (c *Component) CheckAccessToken(ctx context.Context, tokenStr string) (flag bool, err error) {
	sc, err := c.DecodeAccessToken(tokenStr)
	if err != nil {
		err = fmt.Errorf("CheckAccessToken failed, err: %w", err)
		return
	}
	uid := sc["jti"].(float64)
	uidInt := int(uid)
	err = c.client.Get(ctx, fmt.Sprintf(c.config.Prefix+tokenKeyPattern, uidInt)).Err()
	if err != nil {
		err = fmt.Errorf("CheckAccessToken failed2, err: %w", err)
		return
	}
	flag = true
	return
}

func (c *Component) RefreshAccessToken(ctx context.Context, tokenStr string, startTime int64) (resp AccessTokenTicket, err error) {
	sc, err := c.DecodeAccessToken(tokenStr)
	if err != nil {
		err = fmt.Errorf("RefreshAccessToken failed, err: %w", err)
		return
	}
	uid := sc["jti"].(float64)
	uidInt := int(uid)
	return c.CreateAccessToken(ctx, uidInt, startTime)
}

func (c *Component) EncodeAccessToken(jwtId string, uid int, startTime int64) (tokenStr string, err error) {
	jwtToken := jwt.New(jwt.SigningMethodHS256)
	claims := make(jwt.MapClaims)
	claims["jti"] = jwtId                               // jwt的唯一身份标识，防止重放
	claims["iss"] = c.config.Iss                        // JWT的签发者
	claims["sub"] = uid                                 // JWT的主题
	claims["iat"] = startTime                           // JWT的签发时间
	claims["exp"] = startTime + c.config.ExpireInterval // JWT的过期时间
	jwtToken.Claims = claims
	tokenStr, err = jwtToken.SignedString([]byte(c.config.Secret))
	if err != nil {
		err = fmt.Errorf("EncodeAccessToken failed, err: %w", err)
		return
	}
	return
}

func (c *Component) DecodeAccessToken(tokenStr string) (resp map[string]interface{}, err error) {
	tokenParse, err := jwt.Parse(tokenStr, func(jwtToken *jwt.Token) (interface{}, error) {
		return []byte(c.config.Secret), nil
	})
	if err != nil {
		err = fmt.Errorf("DecodeAccessToken failed, err: %w", err)
		return
	}
	var flag bool
	resp, flag = tokenParse.Claims.(jwt.MapClaims)
	if !flag {
		err = fmt.Errorf("DecodeAccessToken failed2, err: assert error")
		return
	}
	return
}
