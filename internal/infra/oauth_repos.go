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

// oauthDBRepository 三方账号，数据库存储
type oauthDBRepository struct {
	db entity.DB
}

// OauthRepositoryProvider provide oauth repository
func OauthRepositoryProvider(injector do.Injector) (adapter.OauthRepository, error) {
	return &oauthDBRepository{
		db: do.MustInvoke[*sqlx.DB](injector),
	}, nil
}

// NewOauthRepositoryTx returns oauth repository with transaction.
func NewOauthRepositoryTx(tx *sqlx.Tx) adapter.OauthRepository {
	return &oauthDBRepository{db: tx}
}

// Find 查询关联用户ID
func (r *oauthDBRepository) Find(ctx context.Context, vendor, vendorUID string) (uuid.UUID, error) {
	stmt := selectOauth.
		Select(colAccountID).
		Where(
			colVendor.Eq(vendor),
			colVendorUID.Eq(vendorUID),
		).
		Limit(1)

	var accountID uuid.UUID
	if err := entity.GetRecord(ctx, &accountID, r.db, stmt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return uuid.Nil, domain.ErrAccountNotFound
		}
		return uuid.Nil, err
	}
	return accountID, nil
}

// Bind 账号绑定
func (r *oauthDBRepository) Bind(ctx context.Context, accountID uuid.UUID, vendor, vendorUID string) error {
	row := &oauthRow{}
	if err := row.SetID(oauthID{AccountID: accountID, Vendor: vendor}); err != nil {
		return fmt.Errorf("set oauth id: %w", err)
	} else if err := database.SetText(&row.VendorID, vendorUID); err != nil {
		return fmt.Errorf("set vendor_uid: %w", err)
	}

	return entity.Upsert(ctx, row, r.db)
}

type oauthID struct {
	AccountID uuid.UUID
	Vendor    string
}

type oauthRow struct {
	AccountID pgtype.UUID `db:"account_id,primaryKey"`
	Vendor    pgtype.Text `db:"vendor,primaryKey"`
	VendorID  pgtype.Text `db:"vendor_uid"`
	CreateAt  pgtype.Int4 `db:"create_at,refuseUpdate"`
	UpdateAt  pgtype.Int4 `db:"update_at"`
}

func (row oauthRow) TableName() string {
	return "oauth_accounts"
}

func (row *oauthRow) BeforeInsert(_ context.Context) error {
	now := time.Now().Unix()

	if err := row.CreateAt.Set(now); err != nil {
		return fmt.Errorf("set create_at: %w", err)
	} else if err := row.UpdateAt.Set(now); err != nil {
		return fmt.Errorf("set update_at: %w", err)
	}
	return nil
}

func (row *oauthRow) BeforeUpdate(_ context.Context) error {
	now := time.Now().Unix()

	if err := row.UpdateAt.Set(now); err != nil {
		return fmt.Errorf("set update_at: %w", err)
	}
	return nil
}

func (row *oauthRow) SetID(id oauthID) error {
	if err := database.SetUUID(&row.AccountID, id.AccountID); err != nil {
		return fmt.Errorf("set account_id: %w", err)
	} else if err := database.SetText(&row.Vendor, id.Vendor); err != nil {
		return fmt.Errorf("set vendor: %w", err)
	}
	return nil
}
