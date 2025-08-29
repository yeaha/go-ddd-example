package infra

import (
	"context"
	"fmt"
	"time"

	"ddd-example/pkg/database"

	"github.com/doug-martin/goqu/v9"
	"github.com/google/uuid"
	"github.com/jackc/pgtype"
	"github.com/joyparty/entity"
	"github.com/joyparty/entity/cache"

	// sql dialect
	_ "github.com/doug-martin/goqu/v9/dialect/sqlite3"
)

func init() {
	entity.DefaultCacher = cache.NewMemoryCache()
}

var (
	sqlite = goqu.Dialect("sqlite3")

	tableAccounts  = goqu.T((accountRow{}).TableName())
	selectAccounts = sqlite.From(tableAccounts).Prepared(true)

	tableOauth  = goqu.T((oauthRow{}).TableName())
	selectOauth = sqlite.From(tableOauth).Prepared(true)

	colAccountID = goqu.C("account_id")
	colEmail     = goqu.C("email")
	colVendor    = goqu.C("vendor")
	colVendorUID = goqu.C("vendor_uid")
)

type baseRow struct {
	ID       pgtype.UUID `db:"id,primaryKey"`
	CreateAt int64       `db:"create_at,refuseUpdate"`
	UpdateAt int64       `db:"update_at"`
}

func (base *baseRow) BeforeInsert(_ context.Context) error {
	if base.ID.Status != pgtype.Present {
		if id, err := uuid.NewV7(); err != nil {
			return fmt.Errorf("create id, %w", err)
		} else if err := database.SetUUID(&base.ID, id); err != nil {
			return fmt.Errorf("set id, %w", err)
		}
	}

	now := time.Now().Unix()
	base.CreateAt = now
	base.UpdateAt = now

	return nil
}

func (base *baseRow) BeforeUpdate(_ context.Context) error {
	base.UpdateAt = time.Now().Unix()

	return nil
}

func (base *baseRow) GetID() uuid.UUID {
	return base.ID.Bytes
}

func (base *baseRow) SetID(id uuid.UUID) error {
	return database.SetUUID(&base.ID, id)
}
