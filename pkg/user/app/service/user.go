package service

import (
	"context"
	"errors"
	"fmt"

	"ddd-example/pkg/user/app/adapter"
	"ddd-example/pkg/user/domain"
	"ddd-example/pkg/user/infra"

	"github.com/jmoiron/sqlx"
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
	_, err := s.Users.FindByEmail(ctx, domain.NormalizeEmail(email))
	if errors.Is(err, domain.ErrUserNotFound) {
		user := &domain.User{}
		if err := user.SetEmail(email); err != nil {
			return nil, fmt.Errorf("set email, %w", err)
		} else if err := user.SetPassword(password); err != nil {
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
