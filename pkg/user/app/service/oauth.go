package service

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
	"gitlab.haochang.tv/yangyi/examine-code/pkg/user/app/adapter"
	"gitlab.haochang.tv/yangyi/examine-code/pkg/user/domain"
	"gitlab.haochang.tv/yangyi/examine-code/pkg/user/infra"
	"gitlab.haochang.tv/yangyi/examine-code/pkg/utils/oauth"
)

// OauthService 三方账号关联逻辑
type OauthService struct {
	Users adapter.UserRepository
	Oauth adapter.OauthRepository
}

// NewOauthService 构造函数
func NewOauthService(tx *sqlx.Tx) *OauthService {
	return &OauthService{
		Users: infra.NewUserDBRepository(tx),
		Oauth: infra.NewOauthDBRepository(tx),
	}
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
