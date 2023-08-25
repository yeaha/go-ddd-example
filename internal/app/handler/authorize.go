package handler

import (
	"context"
	"ddd-example/internal/app/service"
	"ddd-example/internal/domain"
	"fmt"
)

// AuthorizeHandler 验证访问者信息
type AuthorizeHandler struct {
	Session *service.SessionTokenService
}

// Handle 执行
func (h *AuthorizeHandler) Handle(ctx context.Context, payload string) (user *domain.User, newPayload string, err error) {
	user, token, err := h.Session.Retrieve(ctx, payload)
	if err != nil {
		err = fmt.Errorf("retrieve session token, %w", err)
		return
	} else if token.IsExpired() {
		err = domain.ErrSessionTokenExpired
		return
	}

	if token.NeedRenew() {
		newPayload, err = h.Session.Renew(user)
		if err != nil {
			err = fmt.Errorf("renew session token, %w", err)
			return
		}
	}

	return
}
