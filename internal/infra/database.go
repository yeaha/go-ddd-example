package infra

import (
	"github.com/doug-martin/goqu/v9"

	// sql dialect
	_ "github.com/doug-martin/goqu/v9/dialect/postgres"
)

var (
	pgsql = goqu.Dialect("postgres")

	tableAccounts  = goqu.S(`account`).Table(`accounts`)
	selectAccounts = pgsql.From(tableAccounts).Prepared(true)

	tableOauth  = goqu.S(`account`).Table(`oauth`)
	selectOauth = pgsql.From(tableOauth).Prepared(true)
	insertOauth = pgsql.Insert(tableOauth).Prepared(true)

	colAccountID = goqu.C("account_id")
	colCreateAt  = goqu.C("create_at")
	colEmail     = goqu.C("email")
	colID        = goqu.C("id")
	colUpdateAt  = goqu.C("update_at")
	colVendor    = goqu.C("vendor")
	colVendorUID = goqu.C("vendor_uid")
)
