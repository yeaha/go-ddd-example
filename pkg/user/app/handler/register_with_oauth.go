package handler

import (
	"context"
	"errors"
	"fmt"

	uuid "github.com/satori/go.uuid"
	"gitlab.haochang.tv/yangyi/examine-code/pkg/user/app/adapter"
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
	Users   adapter.UserRepository
}

// Handle 三方登录，绑定或注册新账号
func (h *RegisterWithOauthHandler) Handle(ctx context.Context, args RegisterWithOauth) (user *domain.User, sessionToken string, err error) {
	vendorUser, err := h.Oauth.RetrieveVendorUser(ctx, args.VendorToken)
	if err != nil {
		err = fmt.Errorf("retrieve vendor user from cache, %w", err)
		return
	}

	if args.VerifyPassword != "" {
		user, err = h.find(ctx, args)
	} else {
		user, err = h.create(ctx, args)
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

func (h *RegisterWithOauthHandler) find(ctx context.Context, args RegisterWithOauth) (*domain.User, error) {
	user, err := h.Users.FindByEmail(ctx, domain.NormalizeEmail(args.Email))
	if err != nil {
		return nil, fmt.Errorf("find user by email, %w", err)
	} else if !user.ComparePassword(args.VerifyPassword) {
		return nil, domain.ErrWrongPassword
	}
	return user, nil
}

func (h *RegisterWithOauthHandler) create(ctx context.Context, args RegisterWithOauth) (*domain.User, error) {
	email := domain.NormalizeEmail(args.Email)
	user, err := h.Users.FindByEmail(ctx, domain.NormalizeEmail(args.Email))
	if errors.Is(err, domain.ErrUserNotFound) {
		user = &domain.User{}
		user.SetEmail(email)

		// 设置随机密码
		if err := user.SetPassword(uuid.NewV4().String()); err != nil {
			return nil, fmt.Errorf("set password, %w", err)
		}

		if err := h.Users.Create(ctx, user); err != nil {
			return nil, fmt.Errorf("create user, %w", err)
		}

		return user, nil
	} else if err != nil {
		return nil, err
	}

	return nil, domain.ErrEmailRegistered
}
