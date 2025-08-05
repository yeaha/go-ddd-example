package adapter

import (
	"context"

	"ddd-example/internal/domain"

	"github.com/google/uuid"
)

// AccountRepository 账号信息存储
type AccountRepository interface {
	Find(ctx context.Context, accountID uuid.UUID) (*domain.Account, error)
	FindByEmail(ctx context.Context, email string) (*domain.Account, error)
	Create(ctx context.Context, account *domain.Account) error
	Update(ctx context.Context, account *domain.Account) error
}

// OauthRepository 三方账号关联
type OauthRepository interface {
	Bind(ctx context.Context, accountID uuid.UUID, vendor, vendorUID string) error
	Find(ctx context.Context, vendor, vendorUID string) (uuid.UUID, error)
}
