package handler

import (
	"context"

	"ddd-example/pkg/user/app/service"
	"ddd-example/pkg/user/domain"
)

// RenewSessionTokenHandler 构造会话凭证
type RenewSessionTokenHandler struct {
	Session *service.SessionTokenService
}

// Handle 执行构造会话凭证
func (h *RenewSessionTokenHandler) Handle(ctx context.Context, user *domain.User) (payload string, err error) {
	return h.Session.Renew(user)
}
