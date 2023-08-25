package handler

import (
	"context"
	"fmt"

	"ddd-example/internal/app/event"
	"ddd-example/internal/app/service"
	"ddd-example/internal/domain"
)

// LoginWithEmail 使用Email登录，参数
type LoginWithEmail struct {
	Email    string `json:"email" valid:"email,required"`
	Password string `json:"password" valid:",required"`
}

// LoginWithEmailHandler 使用Email登录
type LoginWithEmailHandler struct {
	Session  *service.SessionTokenService
	Accounts *service.AccountService
}

// Handle 执行
func (h *LoginWithEmailHandler) Handle(ctx context.Context, args LoginWithEmail) (account *domain.Account, token string, err error) {
	account, err = h.Accounts.Authorize(ctx, args.Email, args.Password)
	if err != nil {
		err = fmt.Errorf("account authorize, %w", err)
		return
	}

	token, err = h.Session.Generate(ctx, account)
	if err != nil {
		err = fmt.Errorf("generate session token, %w", err)
		return
	}

	event.Publish(event.Login{
		Account: account,
	})
	return
}
