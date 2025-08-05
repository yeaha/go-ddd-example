package infra

import (
	"github.com/doug-martin/goqu/v9"

	// sql dialect
	_ "github.com/doug-martin/goqu/v9/dialect/sqlite3"
)

var (
	sqlite = goqu.Dialect("sqlite3")

	tableAccounts  = goqu.T((accountRow{}).TableName())
	selectAccounts = sqlite.From(tableAccounts).Prepared(true)

	tableOauth  = goqu.T((oauthRow{}).TableName())
	selectOauth = sqlite.From(tableOauth).Prepared(true)
	insertOauth = sqlite.Insert(tableOauth).Prepared(true)

	colAccountID = goqu.C("account_id")
	colCreateAt  = goqu.C("create_at")
	colEmail     = goqu.C("email")
	colID        = goqu.C("id")
	colUpdateAt  = goqu.C("update_at")
	colVendor    = goqu.C("vendor")
	colVendorUID = goqu.C("vendor_uid")
)
