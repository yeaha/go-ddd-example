package infra

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"ddd-example/internal/domain"
	"ddd-example/pkg/database"

	"github.com/google/uuid"
	"github.com/jackc/pgtype"
	"github.com/joyparty/entity"
)

// AccountDBRepository 用户账号，数据库存储
type AccountDBRepository struct {
	db   entity.DB
	base *entity.DomainObjectRepository[uuid.UUID, *domain.Account, *accountRow]
}

// NewAccountDBRepository 构造函数
func NewAccountDBRepository(db entity.DB) *AccountDBRepository {
	return &AccountDBRepository{
		db: db,
		base: entity.NewDomainObjectRepository(
			entity.NewRepository[uuid.UUID, *accountRow](db),
		),
	}
}

// Find 使用ID查找
func (r *AccountDBRepository) Find(ctx context.Context, accountID uuid.UUID) (*domain.Account, error) {
	a, err := r.base.Find(ctx, accountID)
	if entity.IsNotFound(err) {
		return nil, domain.ErrAccountNotFound
	} else if err != nil {
		return nil, err
	}

	return a, nil
}

// FindByEmail 根据email查找对应账号
func (r *AccountDBRepository) FindByEmail(ctx context.Context, email string) (*domain.Account, error) {
	stmt := selectAccounts.Where(colEmail.Eq(email)).Limit(1)

	row := &accountRow{}
	if err := entity.GetRecord(ctx, row, r.db, stmt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrAccountNotFound
		}
		return nil, err
	}

	return row.ToDomainObject()
}

// Create 保存新用户
func (r *AccountDBRepository) Create(ctx context.Context, account *domain.Account) error {
	return r.base.Create(ctx, account)
}

// Update 更新用户数据
func (r *AccountDBRepository) Update(ctx context.Context, account *domain.Account) error {
	return r.base.Update(ctx, account)
}

type accountRow struct {
	ID       pgtype.UUID    `db:"id,primaryKey"`
	Email    pgtype.Varchar `db:"email"`
	Password pgtype.Varchar `db:"password"`
	Setting  pgtype.JSON    `db:"setting"`
	CreateAt int64          `db:"create_at,refuseUpdate"`
	UpdateAt int64          `db:"update_at"`
}

type accountRowSetting struct {
	PasswordSalt string `json:"password_salt"`
	SessionSalt  string `json:"session_salt"`
}

func (row accountRow) TableName() string {
	return "accounts"
}

func (row *accountRow) BeforeInsert(_ context.Context) error {
	if row.ID.Status != pgtype.Present {
		if err := database.SetUUID(&row.ID, uuid.Must(uuid.NewV7())); err != nil {
			return fmt.Errorf("set id, %w", err)
		}
	}

	now := time.Now().Unix()
	row.CreateAt = now
	row.UpdateAt = now
	return nil
}

func (row *accountRow) BeforeUpdate(_ context.Context) error {
	row.UpdateAt = time.Now().Unix()
	return nil
}

func (row accountRow) CacheOption() entity.CacheOption {
	return entity.CacheOption{
		Key:        fmt.Sprintf("account:%s", uuid.UUID(row.ID.Bytes)),
		Expiration: 5 * time.Minute,
	}
}

func (row *accountRow) GetID() uuid.UUID {
	return row.ID.Bytes
}

func (row *accountRow) SetID(id uuid.UUID) error {
	return database.SetUUID(&row.ID, id)
}

func (row *accountRow) Set(_ context.Context, u *domain.Account) error {
	if err := database.SetUUID(&row.ID, u.ID); err != nil {
		return fmt.Errorf("set id, %w", err)
	} else if err := database.SetVarchar(&row.Email, u.Email); err != nil {
		return fmt.Errorf("set email, %w", err)
	} else if err := database.SetVarchar(&row.Password, u.Password); err != nil {
		return fmt.Errorf("set password, %w", err)
	}

	setting := accountRowSetting{
		PasswordSalt: u.PasswordSalt,
		SessionSalt:  u.SessionSalt,
	}
	if err := row.Setting.Set(setting); err != nil {
		return fmt.Errorf("set setting, %w", err)
	}
	return nil
}

func (row accountRow) ToDomainObject() (*domain.Account, error) {
	setting := &accountRowSetting{}
	if err := row.Setting.AssignTo(setting); err != nil {
		return nil, fmt.Errorf("decode setting, %w", err)
	}

	return &domain.Account{
		ID:           row.ID.Bytes,
		Email:        row.Email.String,
		Password:     row.Password.String,
		PasswordSalt: setting.PasswordSalt,
		SessionSalt:  setting.SessionSalt,
	}, nil
}
