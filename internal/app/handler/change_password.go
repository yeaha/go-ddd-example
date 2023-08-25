package handler

import (
	"context"
	"fmt"

	"ddd-example/internal/app/adapter"
	"ddd-example/internal/domain"
)

// ChangePassword 替换密码，参数
type ChangePassword struct {
	User        *domain.User `json:"-"`
	NewPassword string       `json:"new_password" valid:",required"`
	OldPassword string       `json:"old_password" valid:",required"`
}

// ChangePasswordHandler 替换密码
type ChangePasswordHandler struct {
	User adapter.UserRepository
}

// Handle 执行替换密码
func (h *ChangePasswordHandler) Handle(ctx context.Context, args ChangePassword) error {
	user := args.User
	if !user.ComparePassword(args.OldPassword) {
		return fmt.Errorf("compare old password, %w", domain.ErrWrongPassword)
	} else if err := user.SetPassword(args.NewPassword); err != nil {
		return fmt.Errorf("set new password, %w", err)
	} else if err := h.User.Save(ctx, user); err != nil {
		return fmt.Errorf("save user, %w", err)
	}
	return nil
}
