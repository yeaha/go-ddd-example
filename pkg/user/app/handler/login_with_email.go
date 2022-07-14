package handler

import (
	"context"
	"fmt"

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
	Session *service.SessionTokenService
	Users   *service.UserService
}

// Handle 执行
func (h *LoginWithEmailHandler) Handle(ctx context.Context, args LoginWithEmail) (user *domain.User, token string, err error) {
	user, err = h.Users.Authorize(ctx, args.Email, args.Password)
	if err != nil {
		err = fmt.Errorf("user authorize, %w", err)
		return
	}

	token, err = h.Session.Generate(ctx, user)
	if err != nil {
		err = fmt.Errorf("generate session token, %w", err)
		return
	}

	return
}
