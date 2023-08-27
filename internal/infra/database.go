package infra

import (
	"github.com/doug-martin/goqu/v9"

	// sql dialect
	_ "github.com/doug-martin/goqu/v9/dialect/sqlite3"
)

func init() {
	goqu.SetDefaultPrepared(true)
}

var (
	sqlite = goqu.Dialect("sqlite3")

	tableAccounts  = goqu.T(`accounts`)
	selectAccounts = sqlite.From(tableAccounts)

	tableOauth  = goqu.T(`oauth_accounts`)
	selectOauth = sqlite.From(tableOauth)
	insertOauth = sqlite.Insert(tableOauth)

	colAccountID = goqu.C("account_id")
	colCreateAt  = goqu.C("create_at")
	colEmail     = goqu.C("email")
	colID        = goqu.C("id")
	colUpdateAt  = goqu.C("update_at")
	colVendor    = goqu.C("vendor")
	colVendorUID = goqu.C("vendor_uid")
)
