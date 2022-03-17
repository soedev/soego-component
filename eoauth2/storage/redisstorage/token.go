package redisstorage

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/soedev/soego-component/eoauth2/storage/dto"
	"github.com/soedev/soego-component/eredis"
	"github.com/soedev/soego/core/elog"
	"go.uber.org/zap"
)

const (
	tokenRefreshLockPrefix = "ssoTokenRefreshLock:%s"
	newTokenKeyPrefix      = "ssoNewToken:%s"
)

type tokenServer struct {
	redis             *eredis.Component
	uidMapParentToken *uidMapParentToken
	parentToken       *parentToken
	subToken          *subToken
	config            *config
}

func initTokenServer(config *config, redis *eredis.Component) *tokenServer {
	return &tokenServer{
		config:            config,
		redis:             redis,
		uidMapParentToken: newUidMapParentToken(config, redis),
		parentToken:       newParentToken(config, redis),
		subToken:          newSubToken(config, redis),
	}
}

// createParentToken sso的父节点token
func (t *tokenServer) createParentToken(ctx context.Context, pToken dto.Token, uid int64, platform string) (err error) {
	// 1 设置uid 到 parent token关系
	err = t.uidMapParentToken.setToken(ctx, uid, platform, pToken)
	if err != nil {
		return fmt.Errorf("token.createParentToken: create token map failed, err:%w", err)
	}

	// 2 创建父级的token信息
	return t.parentToken.create(ctx, pToken, platform, uid)
}

func (t *tokenServer) renewParentToken(ctx context.Context, pToken dto.Token) (err error) {
	// 1 设置uid 到 parent token关系
	err = t.parentToken.renew(ctx, pToken)
	if err != nil {
		return fmt.Errorf("token.createParentToken: create token map failed, err:%w", err)
	}
	return nil
}

func (t *tokenServer) createToken(ctx context.Context, clientId string, token dto.Token, pToken string) (err error) {
	err = t.parentToken.setToken(ctx, pToken, clientId, token)
	if err != nil {
		return fmt.Errorf("tokenServer.createToken failed, err:%w", err)
	}

	// setTTL new token
	err = t.subToken.create(ctx, token, pToken, clientId)
	return
}

func (t *tokenServer) removeParentToken(ctx context.Context, pToken string) (err error) {
	return t.parentToken.delete(ctx, pToken)
}

// 获取父级token
//func (t *tokenServer) getParentToken(uid int64) (tokenInfo dto.Token, err error) {
//	return t.uidMapParentToken.getParentToken(context.Background(), uid, "web")
//}

func (t *tokenServer) getToken(clientId string, pToken string) (tokenInfo dto.Token, err error) {
	return t.parentToken.getToken(context.Background(), pToken, clientId)
}

func (t *tokenServer) getUidByParentToken(ctx context.Context, pToken string) (uid int64, err error) {
	return t.parentToken.getUid(ctx, pToken)
}

func (t *tokenServer) getParentTokenByToken(ctx context.Context, token string) (pToken string, err error) {
	// 通过子系统token，获得父节点token
	pToken, err = t.subToken.getParentToken(ctx, token)
	return
}

func (t *tokenServer) getUidByToken(ctx context.Context, token string) (uid int64, err error) {
	// 通过子系统token，获得父节点token
	pToken, err := t.getParentTokenByToken(ctx, token)
	if err != nil {
		return
	}
	return t.getUidByParentToken(ctx, pToken)
}

func (t *tokenServer) refreshToken(ctx context.Context, clientId string, pToken string) (tk *dto.Token, err error) {
	var genNewToken dto.Token
	// try to get lock
	tokenRefreshLock, err := t.redis.LockClient().Obtain(ctx, redisTokenRefreshLockKey(pToken), 100*time.Millisecond,
		eredis.WithLockOptionRetryStrategy(eredis.LinearBackoffRetry(10*time.Millisecond)))
	if err != nil {
		return nil, err
	}

	defer func() {
		err = tokenRefreshLock.Release(ctx)
		if err != nil {
			elog.Error("tokenServer.genNewToken: release redis lock failed", zap.Error(err),
				zap.String("clientId", clientId), zap.String("pToken", pToken))
		}
	}()

	// try to get new-token from cache
	{
		tk, err = t.getNewTokenFromCache(ctx, pToken)
		if err != nil && !errors.Is(err, eredis.Nil) { // no-empty error
			return nil, err
		} else if err == nil {
			return tk, nil
		} else {
			// empty cache
		}
	}
	// re-generate token
	{
		genNewToken = dto.NewToken(t.config.parentAccessExpiration)
		tk = &genNewToken
		err = t.createToken(ctx, clientId, genNewToken, pToken)
		if err != nil {
			return
		}
	}

	// write new-token to cache
	err = t.setNewTokenToCache(ctx, tk, pToken)
	if err != nil {
		elog.Error("tokenServer.genNewToken: setTTL new-token to cache failed", zap.Error(err))
		return tk, nil
	}

	return
}

func (t *tokenServer) getNewTokenFromCache(ctx context.Context, pToken string) (tk *dto.Token, err error) {
	newTokenKey := redisNewTokenKey(pToken)

	newTokenBytes, err := t.redis.GetBytes(ctx, newTokenKey)
	if err != nil {
		return nil, err
	}

	tk = &dto.Token{}
	err = tk.Unmarshal(newTokenBytes)
	if err != nil {
		return nil, err
	}

	return
}

func (t *tokenServer) setNewTokenToCache(ctx context.Context, tk *dto.Token, pToken string) (err error) {
	tkBytes, err := tk.Marshal()
	if err != nil {
		return
	}

	// write cache
	err = t.redis.SetEX(ctx, redisNewTokenKey(pToken), tkBytes, time.Minute)
	if err != nil {
		return
	}

	return
}

func redisTokenRefreshLockKey(pToken string) string {
	return fmt.Sprintf(tokenRefreshLockPrefix, pToken)
}

func redisNewTokenKey(pToken string) string {
	return fmt.Sprintf(newTokenKeyPrefix, pToken)
}
