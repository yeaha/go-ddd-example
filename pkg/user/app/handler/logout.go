package handler

import (
	"context"

	"ddd-example/pkg/user/app/service"
	"ddd-example/pkg/user/domain"
)

// LogoutHandler 退出登录
type LogoutHandler struct {
	Session *service.SessionTokenService
}

// Handle 执行
func (h *LogoutHandler) Handle(ctx context.Context, user *domain.User) error {
	return h.Session.Suspend(ctx, user)
}
