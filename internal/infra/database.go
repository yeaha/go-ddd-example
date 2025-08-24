package infra

import (
	"github.com/doug-martin/goqu/v9"
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
