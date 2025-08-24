package handler

import (
	"context"
	"fmt"

	"ddd-example/internal/app/event"
	"ddd-example/internal/app/service"
	"ddd-example/internal/domain"
)

// Register 账号注册，参数
type Register struct {
	Email    string `json:"email" validate:"email"`
	Password string `json:"password" validate:"required"`
}

// RegisterHandler 账号注册
type RegisterHandler struct {
	Session  *service.SessionTokenService `do:""`
	Accounts *service.AccountService      `do:""`
}

// Handle 执行账号注册
func (h *RegisterHandler) Handle(ctx context.Context, args Register) (account *domain.Account, token string, err error) {
	account, err = h.Accounts.Create(ctx, args.Email, args.Password)
	if err != nil {
		return
	}

	event.Publish(event.Register{
		Account: account,
	})

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
