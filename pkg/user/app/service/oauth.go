package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	uuid "github.com/satori/go.uuid"
	"gitlab.haochang.tv/yangyi/examine-code/pkg/user/app/adapter"
	"gitlab.haochang.tv/yangyi/examine-code/pkg/user/domain"
	"gitlab.haochang.tv/yangyi/examine-code/pkg/utils/oauth"
)

// OauthService 三方账号关联逻辑
type OauthService struct {
	Users adapter.UserRepository
	Oauth adapter.OauthRepository
	Cache adapter.Cacher
}

// Find 查询关联账号
func (s *OauthService) Find(ctx context.Context, vendorUser *oauth.User) (*domain.User, error) {
	userID, err := s.Oauth.Find(ctx, vendorUser.Vendor, vendorUser.ID)
	if err != nil {
		return nil, fmt.Errorf("find user_id, %w", err)
	}

	return s.Users.Find(ctx, userID)
}

// Bind 绑定账号
func (s *OauthService) Bind(ctx context.Context, user *domain.User, vendorUser *oauth.User) error {
	return s.Oauth.Bind(ctx, user.ID, vendorUser.Vendor, vendorUser.ID)
}

// CacheVendorUser 缓存三方账号信息
func (s *OauthService) CacheVendorUser(ctx context.Context, vendorUser *oauth.User) (token string, err error) {
	value, err := json.Marshal(vendorUser)
	if err != nil {
		err = fmt.Errorf("encode vendor user, %w", err)
		return
	}

	token = uuid.NewV4().String()
	err = s.Cache.Put(ctx, token, value, 10*time.Minute)
	return
}

// RetrieveVendorUser 从缓存内读取三方账号信息
func (s *OauthService) RetrieveVendorUser(ctx context.Context, token string) (*oauth.User, error) {
	value, err := s.Cache.Get(ctx, token)
	if err != nil {
		return nil, err
	}

	user := &oauth.User{}
	if err := json.Unmarshal(value, &user); err != nil {
		return nil, fmt.Errorf("decode vendor user, %w", err)
	}
	return user, nil
}
