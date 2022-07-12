package handler

import (
	"context"
	"errors"
	"fmt"

	"gitlab.haochang.tv/yangyi/examine-code/pkg/user/app/adapter"
	"gitlab.haochang.tv/yangyi/examine-code/pkg/user/app/service"
	"gitlab.haochang.tv/yangyi/examine-code/pkg/user/domain"
)

// Register 账号注册，参数
type Register struct {
	Email    string `json:"email" valid:"email,required"`
	Password string `json:"password" valid:"required"`
}

// RegisterHandler 账号注册
type RegisterHandler struct {
	User    adapter.UserRepository
	Session *service.SessionTokenService
}

// Handle 执行账号注册
func (h *RegisterHandler) Handle(ctx context.Context, args Register) (user *domain.User, token string, err error) {
	_, err = h.User.FindByEmail(ctx, domain.NormalizeEmail(args.Email))
	if !errors.Is(err, domain.ErrUserNotFound) {
		if err == nil {
			err = domain.ErrEmailRegistered
			return
		}

		err = fmt.Errorf("find user by email, %w", err)
		return
	}

	user = &domain.User{}
	user.SetEmail(args.Email)
	if err = user.SetPassword(args.Password); err != nil {
		err = fmt.Errorf("set password, %w", err)
		return
	} else if err = h.User.Create(ctx, user); err != nil {
		err = fmt.Errorf("save user, %w", err)
		return
	}

	token, err = h.Session.Generate(ctx, user)
	if err != nil {
		err = fmt.Errorf("generate session token, %w", err)
	}
	return
}
