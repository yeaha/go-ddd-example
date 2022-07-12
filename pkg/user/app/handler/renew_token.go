package handler

import (
	"context"

	"gitlab.haochang.tv/yangyi/examine-code/pkg/user/app/service"
	"gitlab.haochang.tv/yangyi/examine-code/pkg/user/domain"
)

// RenewTokenHandler 构造会话凭证
type RenewTokenHandler struct {
	Session *service.SessionTokenService
}

// Handle 执行构造会话凭证
func (h *RenewTokenHandler) Handle(ctx context.Context, user *domain.User) (payload string, err error) {
	return h.Session.Renew(user)
}
