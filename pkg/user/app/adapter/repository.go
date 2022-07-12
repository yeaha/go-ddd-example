package adapter

import (
	"context"

	uuid "github.com/satori/go.uuid"
	"gitlab.haochang.tv/yangyi/examine-code/pkg/user/domain"
)

// UserRepository 账号信息存储
type UserRepository interface {
	Find(ctx context.Context, userID uuid.UUID) (*domain.User, error)
	FindByEmail(ctx context.Context, email string) (*domain.User, error)
	Create(ctx context.Context, user *domain.User) error
	Save(ctx context.Context, user *domain.User) error
}
