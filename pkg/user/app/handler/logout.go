package handler

import (
	"context"

	"gitlab.haochang.tv/yangyi/examine-code/pkg/user/app/service"
	"gitlab.haochang.tv/yangyi/examine-code/pkg/user/domain"
)

// LogoutHandler 退出登录
type LogoutHandler struct {
	Session *service.SessionTokenService
}

// Handle 执行
func (h *LogoutHandler) Handle(ctx context.Context, user *domain.User) error {
	return h.Session.Suspend(ctx, user)
}
