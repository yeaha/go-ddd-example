package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"
	"gitlab.haochang.tv/yangyi/examine-code/pkg/user/app/adapter"
	"gitlab.haochang.tv/yangyi/examine-code/pkg/user/domain"
	"gitlab.haochang.tv/yangyi/examine-code/pkg/user/infra"
)

// UserService 账号逻辑
type UserService struct {
	Users adapter.UserRepository
}

// NewUserService 构造函数
func NewUserService(tx *sqlx.Tx) *UserService {
	return &UserService{
		Users: infra.NewUserDBRepository(tx),
	}
}

// Authorize 验证
func (s *UserService) Authorize(ctx context.Context, email, password string) (*domain.User, error) {
	user, err := s.Users.FindByEmail(ctx, domain.NormalizeEmail(email))
	if err != nil {
		return nil, err
	} else if !user.ComparePassword(password) {
		return nil, domain.ErrWrongPassword
	}
	return user, nil
}

// Create 创建新账号
func (s *UserService) Create(ctx context.Context, email, password string) (*domain.User, error) {
	user, err := s.Users.FindByEmail(ctx, domain.NormalizeEmail(email))
	if errors.Is(err, domain.ErrUserNotFound) {
		user = &domain.User{}
		user.SetEmail(email)

		if err := user.SetPassword(password); err != nil {
			return nil, fmt.Errorf("set password, %w", err)
		} else if err := s.Users.Create(ctx, user); err != nil {
			return nil, err
		}
		return user, nil
	} else if err != nil {
		return nil, fmt.Errorf("find user by email, %w", err)
	}
	return nil, domain.ErrEmailRegistered
}
