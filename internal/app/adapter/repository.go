package adapter

import (
	"context"

	"ddd-example/internal/domain"

	uuid "github.com/satori/go.uuid"
)

// UserRepository 账号信息存储
type UserRepository interface {
	Find(ctx context.Context, userID uuid.UUID) (*domain.User, error)
	FindByEmail(ctx context.Context, email string) (*domain.User, error)
	Create(ctx context.Context, user *domain.User) error
	Save(ctx context.Context, user *domain.User) error
}

// OauthRepository 三方账号关联
type OauthRepository interface {
	Bind(ctx context.Context, userID uuid.UUID, vendor, vendorUID string) error
	Find(ctx context.Context, vendor, vendorUID string) (uuid.UUID, error)
}
