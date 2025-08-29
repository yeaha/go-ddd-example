package infra

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"ddd-example/internal/app/adapter"
	"ddd-example/internal/domain"
	"ddd-example/pkg/database"

	"github.com/google/uuid"
	"github.com/jackc/pgtype"
	"github.com/jmoiron/sqlx"
	"github.com/joyparty/entity"
	"github.com/samber/do/v2"
)

// accountDBRepository 用户账号，数据库存储
type accountDBRepository struct {
	db   entity.DB
	base *entity.DomainObjectRepository[uuid.UUID, *domain.Account, *accountRow]
}

// AccountRepositoryProvider 账户仓库提供者
func AccountRepositoryProvider(injector do.Injector) (adapter.AccountRepository, error) {
	return newAccountDBRepository(do.MustInvoke[*sqlx.DB](injector)), nil
}

// NewAccountRepositoryTx returns a new AccountDBRepository with a transaction.
func NewAccountRepositoryTx(tx *sqlx.Tx) adapter.AccountRepository {
	return newAccountDBRepository(tx)
}

func newAccountDBRepository(db entity.DB) *accountDBRepository {
	return &accountDBRepository{
		db: db,
		base: entity.NewDomainObjectRepository(
			entity.NewRepository[uuid.UUID, *accountRow](db),
		),
	}
}

// Find 使用ID查找
func (r *accountDBRepository) Find(ctx context.Context, accountID uuid.UUID) (*domain.Account, error) {
	a, err := r.base.Find(ctx, accountID)
	if entity.IsNotFound(err) {
		return nil, domain.ErrAccountNotFound
	} else if err != nil {
		return nil, err
	}

	return a, nil
}

// FindByEmail 根据email查找对应账号
func (r *accountDBRepository) FindByEmail(ctx context.Context, email string) (*domain.Account, error) {
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
func (r *accountDBRepository) Create(ctx context.Context, account *domain.Account) error {
	return r.base.Create(ctx, account)
}

// Update 更新用户数据
func (r *accountDBRepository) Update(ctx context.Context, account *domain.Account) error {
	return r.base.Update(ctx, account)
}

type accountRow struct {
	baseEntity

	Email    pgtype.Text `db:"email"`
	Password pgtype.Text `db:"password"`
	Setting  pgtype.JSON `db:"setting"`
}

type accountRowSetting struct {
	PasswordSalt string `json:"password_salt"`
	SessionSalt  string `json:"session_salt"`
}

func (row accountRow) TableName() string {
	return "accounts"
}

func (row accountRow) CacheOption() entity.CacheOption {
	return entity.CacheOption{
		Key:        fmt.Sprintf("account:%s", row.GetID()),
		Expiration: 5 * time.Minute,
	}
}

func (row *accountRow) Set(_ context.Context, a *domain.Account) error {
	setting := accountRowSetting{
		PasswordSalt: a.PasswordSalt,
		SessionSalt:  a.SessionSalt,
	}
	if err := row.Setting.Set(setting); err != nil {
		return fmt.Errorf("set setting, %w", err)
	}

	return errors.Join(
		row.SetID(a.ID),
		database.SetText(&row.Email, a.Email),
		database.SetText(&row.Password, a.Password),
	)
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
