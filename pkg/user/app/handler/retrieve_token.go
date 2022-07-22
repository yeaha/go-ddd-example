package handler

import (
	"context"

	"ddd-example/pkg/user/app/service"
	"ddd-example/pkg/user/domain"
)

// RetrieveTokenHandler 解码会话凭证
type RetrieveTokenHandler struct {
	Session *service.SessionTokenService
}

// Handle 执行会话凭证解码
func (h *RetrieveTokenHandler) Handle(ctx context.Context, token string) (*domain.User, domain.SessionToken, error) {
	return h.Session.Retrieve(ctx, token)
}
