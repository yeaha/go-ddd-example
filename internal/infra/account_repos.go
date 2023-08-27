package infra

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"ddd-example/internal/domain"
	"ddd-example/pkg/database"

	"github.com/jackc/pgtype"
	"github.com/joyparty/entity"
	uuid "github.com/satori/go.uuid"
)

// AccountDBRepository 用户账号，数据库存储
type AccountDBRepository struct {
	db entity.DB
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

func newAccountRow() *accountRow {
	row := &accountRow{}

	row.ID.Status = pgtype.Null
	row.Email.Status = pgtype.Null
	row.Password.Status = pgtype.Null
	row.Setting.Status = pgtype.Null

	return row
}

func (row accountRow) TableName() string {
	return tableAccounts.GetTable()
}

func (row *accountRow) OnEntityEvent(_ context.Context, ev entity.Event) error {
	if ev == entity.EventBeforeInsert {
		if row.ID.Status != pgtype.Present {
			if err := database.SetUUID(&row.ID, uuid.NewV4()); err != nil {
				return fmt.Errorf("set id, %w", err)
			}
		}

		now := time.Now().Unix()
		row.CreateAt = now
		row.UpdateAt = now
	} else if ev == entity.EventBeforeUpdate {
		row.UpdateAt = time.Now().Unix()
	}

	return nil
}

func (row accountRow) CacheOption() entity.CacheOption {
	return entity.CacheOption{
		Key:        fmt.Sprintf("account:%s", uuid.UUID(row.ID.Bytes)),
		Expiration: 5 * time.Minute,
	}
}

func (row *accountRow) Set(u *domain.Account) error {
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

func (row accountRow) toDomain() (*domain.Account, error) {
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

// NewAccountDBRepository 构造函数
func NewAccountDBRepository(db entity.DB) *AccountDBRepository {
	return &AccountDBRepository{db: db}
}

// Find 使用ID查找
func (repos *AccountDBRepository) Find(ctx context.Context, accountID uuid.UUID) (*domain.Account, error) {
	row := &accountRow{}
	if err := database.SetUUID(&row.ID, accountID); err != nil {
		return nil, fmt.Errorf("set id, %w", err)
	} else if err := entity.Load(ctx, row, repos.db); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrAccountNotFound
		}
		return nil, err
	}

	account, err := row.toDomain()
	if err != nil {
		return nil, fmt.Errorf("retrieve row values, %w", err)
	}
	return account, nil
}

// FindByEmail 根据email查找对应账号
func (repos *AccountDBRepository) FindByEmail(ctx context.Context, email string) (*domain.Account, error) {
	stmt := selectAccounts.Where(colEmail.Eq(email)).Limit(1)

	row := &accountRow{}
	if err := entity.GetRecord(ctx, row, repos.db, stmt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrAccountNotFound
		}
		return nil, err
	}

	account, err := row.toDomain()
	if err != nil {
		return nil, fmt.Errorf("retrieve row values, %w", err)
	}
	return account, nil
}

// Create 保存新用户
func (repos *AccountDBRepository) Create(ctx context.Context, account *domain.Account) error {
	row := newAccountRow()
	if err := row.Set(account); err != nil {
		return fmt.Errorf("set row values, %w", err)
	}

	_, err := entity.Insert(ctx, row, repos.db)
	if err == nil {
		account.ID = row.ID.Bytes
	}
	return err
}

// Save 更新用户数据
func (repos *AccountDBRepository) Save(ctx context.Context, account *domain.Account) error {
	row := newAccountRow()
	if err := row.Set(account); err != nil {
		return fmt.Errorf("set row values, %w", err)
	}

	return entity.Update(ctx, row, repos.db)
}
