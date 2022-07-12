package infra

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgtype"
	"github.com/joyparty/entity"
	uuid "github.com/satori/go.uuid"
	"gitlab.haochang.tv/yangyi/examine-code/pkg/user/domain"
	"gitlab.haochang.tv/yangyi/examine-code/pkg/utils/database"
)

// UserDBRepository 用户账号，数据库存储
type UserDBRepository struct {
	db entity.DB
}

type userRow struct {
	ID       pgtype.UUID        `db:"id,primaryKey"`
	Email    pgtype.Varchar     `db:"email"`
	Password pgtype.Varchar     `db:"password"`
	Setting  pgtype.JSONB       `db:"setting"`
	CreateAt pgtype.Timestamptz `db:"create_at,refuseUpdate"`
	UpdateAt pgtype.Timestamptz `db:"update_at"`
}

type userRowSetting struct {
	PasswordSalt string `json:"password_salt"`
	SessionSalt  string `json:"session_salt"`
}

func newUserRow() *userRow {
	row := &userRow{}

	row.ID.Status = pgtype.Null
	row.Email.Status = pgtype.Null
	row.Password.Status = pgtype.Null
	row.Setting.Status = pgtype.Null
	row.CreateAt.Status = pgtype.Null
	row.UpdateAt.Status = pgtype.Null

	return row
}

func (row userRow) TableName() string {
	return fmt.Sprintf("%s.%s", tableUsers.GetSchema(), tableUsers.GetTable())
}

func (row *userRow) OnEntityEvent(_ context.Context, ev entity.Event) error {
	if ev == entity.EventBeforeInsert {
		if row.ID.Status != pgtype.Present {
			if err := database.SetUUID(&row.ID, uuid.NewV4()); err != nil {
				return fmt.Errorf("set id, %w", err)
			}
		}

		now := time.Now()
		database.SetTimestamptz(&row.CreateAt, now)
		database.SetTimestamptz(&row.UpdateAt, now)
	} else if ev == entity.EventBeforeUpdate {
		database.SetTimestamptz(&row.UpdateAt, time.Now())
	}

	return nil
}

func (row userRow) CacheOption() entity.CacheOption {
	return entity.CacheOption{
		Key:        fmt.Sprintf("user:%s", uuid.UUID(row.ID.Bytes)),
		Expiration: 5 * time.Minute,
	}
}

func (row *userRow) Set(u *domain.User) error {
	if err := database.SetUUID(&row.ID, u.ID); err != nil {
		return fmt.Errorf("set id, %w", err)
	} else if err := database.SetVarchar(&row.Email, u.Email); err != nil {
		return fmt.Errorf("set email, %w", err)
	} else if err := database.SetVarchar(&row.Password, u.Password); err != nil {
		return fmt.Errorf("set password, %w", err)
	}

	setting := userRowSetting{
		PasswordSalt: u.PasswordSalt,
		SessionSalt:  u.SessionSalt,
	}
	if err := row.Setting.Set(setting); err != nil {
		return fmt.Errorf("set setting, %w", err)
	}
	return nil
}

func (row userRow) toDomain() (*domain.User, error) {
	setting := &userRowSetting{}
	if err := row.Setting.AssignTo(setting); err != nil {
		return nil, fmt.Errorf("decode setting, %w", err)
	}

	return &domain.User{
		ID:           row.ID.Bytes,
		Email:        row.Email.String,
		Password:     row.Password.String,
		PasswordSalt: setting.PasswordSalt,
		SessionSalt:  setting.SessionSalt,
	}, nil
}

// NewUserDBRepository 构造函数
func NewUserDBRepository(db entity.DB) *UserDBRepository {
	return &UserDBRepository{db: db}
}

// Find 使用ID查找
func (repos *UserDBRepository) Find(ctx context.Context, userID uuid.UUID) (*domain.User, error) {
	row := &userRow{}
	if err := database.SetUUID(&row.ID, userID); err != nil {
		return nil, fmt.Errorf("set id, %w", err)
	} else if err := entity.Load(ctx, row, repos.db); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrUserNotFound
		}
		return nil, err
	}

	user, err := row.toDomain()
	if err != nil {
		return nil, fmt.Errorf("retrieve row values, %w", err)
	}
	return user, nil
}

// FindByEmail 根据email查找对应账号
func (repos *UserDBRepository) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	stmt := selectUsers.Where(colEmail.Eq(email)).Limit(1)

	row := &userRow{}
	if err := entity.GetRecord(ctx, row, repos.db, stmt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrUserNotFound
		}
		return nil, err
	}

	user, err := row.toDomain()
	if err != nil {
		return nil, fmt.Errorf("retrieve row values, %w", err)
	}
	return user, nil
}

// Create 保存新用户
func (repos *UserDBRepository) Create(ctx context.Context, user *domain.User) error {
	row := newUserRow()
	if err := row.Set(user); err != nil {
		return fmt.Errorf("set row values, %w", err)
	}

	_, err := entity.Insert(ctx, row, repos.db)
	if err == nil {
		user.ID = row.ID.Bytes
	}
	return err
}

// Save 更新用户数据
func (repos *UserDBRepository) Save(ctx context.Context, user *domain.User) error {
	row := newUserRow()
	if err := row.Set(user); err != nil {
		return fmt.Errorf("set row values, %w", err)
	}

	return entity.Update(ctx, row, repos.db)
}
