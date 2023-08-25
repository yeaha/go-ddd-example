package infra

import (
	"context"
	"database/sql"
	"errors"

	"ddd-example/internal/domain"

	"github.com/doug-martin/goqu/v9"
	"github.com/joyparty/entity"
	uuid "github.com/satori/go.uuid"
)

// OauthDBRepository 三方账号，数据库存储
type OauthDBRepository struct {
	db entity.DB
}

// NewOauthDBRepository 构造函数
func NewOauthDBRepository(db entity.DB) *OauthDBRepository {
	return &OauthDBRepository{db: db}
}

// Find 查询关联用户ID
func (repos *OauthDBRepository) Find(ctx context.Context, vendor, vendorUID string) (uuid.UUID, error) {
	stmt := selectOauth.
		Select(colAccountID).
		Where(
			colVendor.Eq(vendor),
			colVendorUID.Eq(vendorUID),
		).
		Limit(1)

	var accountID uuid.UUID
	if err := entity.GetRecord(ctx, &accountID, repos.db, stmt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return uuid.Nil, domain.ErrAccountNotFound
		}
		return uuid.Nil, err
	}
	return accountID, nil
}

// Bind 账号绑定
func (repos *OauthDBRepository) Bind(ctx context.Context, accountID uuid.UUID, vendor, vendorUID string) error {
	stmt := insertOauth.
		Rows(goqu.Record{
			"account_id": accountID,
			"vendor":     vendor,
			"vendor_uid": vendorUID,
			"create_at":  goqu.L(`now()`),
			"update_at":  goqu.L(`now()`),
		}).
		OnConflict(goqu.DoUpdate("account_id, vendor", goqu.Record{
			"vendor_uid": vendorUID,
			"update_at":  goqu.L(`now()`),
		}))

	_, err := entity.ExecInsert(ctx, repos.db, stmt)
	return err
}
