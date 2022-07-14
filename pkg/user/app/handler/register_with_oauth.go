package handler

import (
	"context"
	"fmt"

	uuid "github.com/satori/go.uuid"
	"gitlab.haochang.tv/yangyi/examine-code/pkg/user/app/service"
	"gitlab.haochang.tv/yangyi/examine-code/pkg/user/domain"
)

// RegisterWithOauth 三方账号注册，参数
type RegisterWithOauth struct {
	VendorToken    string `json:"vendor_token" valid:",required"`
	Email          string `json:"email" valid:"email,required"`
	VerifyPassword string `json:"verify_password" valid:",optional"` // 绑定账号，需要提供密码
}

// RegisterWithOauthHandler 三方账号注册
type RegisterWithOauthHandler struct {
	Session *service.SessionTokenService
	Oauth   *service.OauthService
	Users   *service.UserService
}

// Handle 三方登录，绑定或注册新账号
func (h *RegisterWithOauthHandler) Handle(ctx context.Context, args RegisterWithOauth) (user *domain.User, sessionToken string, err error) {
	vendorUser, err := h.Oauth.RetrieveVendorUser(ctx, args.VendorToken)
	if err != nil {
		err = fmt.Errorf("retrieve vendor user from cache, %w", err)
		return
	}

	if args.VerifyPassword != "" {
		user, err = h.Users.Authorize(ctx, args.Email, args.VerifyPassword)
	} else {
		user, err = h.Users.Create(ctx, args.Email, uuid.NewV4().String())
	}

	if err != nil {
		return
	}

	if err = h.Oauth.Bind(ctx, user, vendorUser); err != nil {
		err = fmt.Errorf("bound vendor user, %w", err)
		return
	}

	sessionToken, err = h.Session.Generate(ctx, user)
	if err != nil {
		err = fmt.Errorf("generate session token, %w", err)
		return
	}
	return
}
