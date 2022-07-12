package handler

import (
	"context"
	"fmt"

	"gitlab.haochang.tv/yangyi/examine-code/pkg/user/app/adapter"
	"gitlab.haochang.tv/yangyi/examine-code/pkg/user/app/service"
	"gitlab.haochang.tv/yangyi/examine-code/pkg/user/domain"
)

// LoginWithEmail 使用Email登录，参数
type LoginWithEmail struct {
	Email    string `json:"email" valid:"email,required"`
	Password string `json:"password" valid:",required"`
}

// LoginWithEmailHandler 使用Email登录
type LoginWithEmailHandler struct {
	Users   adapter.UserRepository
	Session *service.SessionTokenService
}

// Handle 执行
func (h *LoginWithEmailHandler) Handle(ctx context.Context, args LoginWithEmail) (user *domain.User, token string, err error) {
	user, err = h.Users.FindByEmail(ctx, domain.NormalizeEmail(args.Email))
	if err != nil {
		err = fmt.Errorf("find user, %w", err)
		return
	} else if !user.ComparePassword(args.Password) {
		return nil, "", domain.ErrWrongPassword
	}

	token, err = h.Session.Generate(ctx, user)
	if err != nil {
		err = fmt.Errorf("generate session token, %w", err)
		return
	}

	return
}
