package service

import (
	"context"
	"errors"
	"fmt"

	"ddd-example/internal/app/adapter"
	"ddd-example/internal/domain"
	"ddd-example/internal/infra"

	"github.com/joyparty/entity"
)

// AccountService 账号逻辑
type AccountService struct {
	Accounts adapter.AccountRepository `do:""`
}

// NewAccountService 构造函数
func NewAccountService(db entity.DB) *AccountService {
	return &AccountService{
		Accounts: infra.NewAccountRepository(db),
	}
}

// Authorize 验证
func (s *AccountService) Authorize(ctx context.Context, email, password string) (*domain.Account, error) {
	account, err := s.Accounts.FindByEmail(ctx, domain.NormalizeEmail(email))
	if err != nil {
		return nil, err
	} else if !account.ComparePassword(password) {
		return nil, domain.ErrWrongPassword
	}
	return account, nil
}

// Create 创建新账号
func (s *AccountService) Create(ctx context.Context, email, password string) (*domain.Account, error) {
	_, err := s.Accounts.FindByEmail(ctx, domain.NormalizeEmail(email))
	if errors.Is(err, domain.ErrAccountNotFound) {
		account := &domain.Account{}
		if err := account.SetEmail(email); err != nil {
			return nil, fmt.Errorf("set email, %w", err)
		} else if err := account.SetPassword(password); err != nil {
			return nil, fmt.Errorf("set password, %w", err)
		} else if err := s.Accounts.Create(ctx, account); err != nil {
			return nil, err
		}
		return account, nil
	} else if err != nil {
		return nil, fmt.Errorf("find user by email, %w", err)
	}
	return nil, domain.ErrEmailRegistered
}
