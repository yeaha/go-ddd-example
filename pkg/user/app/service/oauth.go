package service

import (
	"context"
	"fmt"

	"ddd-example/pkg/user/app/adapter"
	"ddd-example/pkg/user/domain"
	"ddd-example/pkg/user/infra"
	"ddd-example/pkg/utils/oauth"

	"github.com/jmoiron/sqlx"
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
