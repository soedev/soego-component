package ssostorage

import (
	"context"
	"fmt"
	"time"

	"github.com/soedev/soego-component/eoauth2/server/model"
	"github.com/soedev/soego-component/eredis"
	"github.com/spf13/cast"
)

type subToken struct {
	config           *config
	redis            *eredis.Component
	fieldParentToken string
	fieldClientId    string
	fieldTokenInfo   string
	fieldCtime       string
	fieldAccessInfo  string
}

func newSubToken(config *config, redis *eredis.Component) *subToken {
	return &subToken{
		config:           config,
		fieldCtime:       "_ct", // create time
		fieldParentToken: "_pt",
		fieldClientId:    "_id",
		fieldTokenInfo:   "_t",
		fieldAccessInfo:  "_a",
		redis:            redis,
	}
}

func (s *subToken) getKey(subToken string) string {
	return fmt.Sprintf(s.config.subTokenMapParentTokenKey, subToken)
}

func (s *subToken) create(ctx context.Context, token model.SubToken, parentToken string, clientId string, accessData *AccessData) error {
	err := s.redis.HMSet(ctx, s.getKey(token.Token.Token), map[string]interface{}{
		s.fieldParentToken: parentToken,
		s.fieldClientId:    clientId,
		s.fieldCtime:       time.Now().Unix(),
		s.fieldAccessInfo:  accessData.Marshal(),
		s.fieldTokenInfo:   token.StoreData.Marshal(),
	}, time.Duration(token.Token.ExpiresIn)*time.Second)
	if err != nil {
		return fmt.Errorf("subToken.create token failed, err:%w", err)
	}
	return nil
}

func (s *subToken) getAccess(ctx context.Context, token string) (storeData *AccessData, err error) {
	infoBytes, err := s.redis.Client().HGet(ctx, s.getKey(token), s.fieldAccessInfo).Bytes()
	info := &AccessData{}
	err = info.Unmarshal(infoBytes)
	if err != nil {
		err = fmt.Errorf("subToken getAccess json unmarshal failed, err: %w", err)
		return
	}
	return info, nil
}

// remove 触发这个场景是refresh token操作，给30s时间，避免换token的时间差，prev token过早失效导致的业务问题
func (s *subToken) remove(ctx context.Context, token string) (bool, error) {
	return s.redis.Expire(ctx, s.getKey(token), 30*time.Second)
}

// 通过子系统token，获得父节点token
func (s *subToken) getParentToken(ctx context.Context, subToken string) (parentToken string, err error) {
	parentToken, err = s.redis.HGet(ctx, s.getKey(subToken), s.fieldParentToken)
	if err != nil {
		err = fmt.Errorf("subToken.getParentToken failed, %w", err)
		return
	}
	return
}

// SubTokenStore 存储的所有信息
type SubTokenStore struct {
	Ctime       int64               `json:"ctime"`
	ParentToken string              `json:"parentToken"`
	ClientId    string              `json:"clientId"`
	TokenInfo   *model.SubTokenData `json:"tokenInfo"`
	AccessInfo  *AccessData         `json:"accessInfo"`
	TTL         int64               `json:"ttl"`
}

func (p *subToken) processData(store *SubTokenStore, key string, value string) {
	switch true {
	case key == p.fieldCtime:
		store.Ctime = cast.ToInt64(value)
	case key == p.fieldParentToken:
		store.ParentToken = value
	case key == p.fieldClientId:
		store.ClientId = value
	case key == p.fieldTokenInfo:
		store.TokenInfo.Unmarshal([]byte(value))
	case key == p.fieldAccessInfo:
		store.AccessInfo.Unmarshal([]byte(value))
	}
}

func (p *subToken) getAll(ctx context.Context, token string) (output *SubTokenStore, err error) {
	allInfo, err := p.redis.Client().HGetAll(ctx, p.getKey(token)).Result()
	if err != nil {
		err = fmt.Errorf("tokgen get redis hmget string error, %w", err)
		return
	}
	output = &SubTokenStore{
		Ctime:       0,
		ParentToken: "",
		ClientId:    "",
		TokenInfo:   &model.SubTokenData{},
		AccessInfo:  &AccessData{},
		TTL:         0,
	}
	for key, value := range allInfo {
		p.processData(output, key, value)
	}
	ttl, err := p.redis.Client().TTL(ctx, p.getKey(token)).Result()
	if err != nil {
		err = fmt.Errorf("parentToken getAll failed,err: %w", err)
		return
	}
	output.TTL = ttl.Milliseconds() / 1000
	return
}
