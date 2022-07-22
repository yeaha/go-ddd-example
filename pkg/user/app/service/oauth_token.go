package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"ddd-example/pkg/user/app/adapter"
	"ddd-example/pkg/user/domain"
	"ddd-example/pkg/utils/oauth"

	uuid "github.com/satori/go.uuid"
)

// OauthTokenService 三方验证结果缓存
type OauthTokenService struct {
	Cache adapter.Cacher
}

// Save 缓存三方账号信息
func (s *OauthTokenService) Save(ctx context.Context, vendorUser *oauth.User) (token string, err error) {
	value, err := json.Marshal(vendorUser)
	if err != nil {
		err = fmt.Errorf("encode vendor user, %w", err)
		return
	}

	token = uuid.NewV4().String()
	err = s.Cache.Put(ctx, token, value, 10*time.Minute)
	return
}

// Retrieve 从缓存内读取三方账号信息
func (s *OauthTokenService) Retrieve(ctx context.Context, token string) (*oauth.User, error) {
	value, err := s.Cache.Get(ctx, token)
	if err != nil {
		if errors.Is(err, domain.ErrMissingCache) {
			return nil, domain.ErrInvalidOauthToken
		}
		return nil, err
	}

	user := &oauth.User{}
	if err := json.Unmarshal(value, &user); err != nil {
		return nil, fmt.Errorf("decode vendor user, %w", err)
	}
	return user, nil
}
