package infra

import (
	"github.com/doug-martin/goqu/v9"

	// sql dialect
	_ "github.com/doug-martin/goqu/v9/dialect/postgres"
)

var (
	pgsql = goqu.Dialect("postgres")

	tableUsers  = goqu.S(`users`).Table(`users`)
	selectUsers = pgsql.From(tableUsers).Prepared(true)

	tableOauth  = goqu.S(`users`).Table(`oauth`)
	selectOauth = pgsql.From(tableOauth).Prepared(true)
	insertOauth = pgsql.Insert(tableOauth).Prepared(true)

	colCreateAt  = goqu.C("create_at")
	colEmail     = goqu.C("email")
	colID        = goqu.C("id")
	colUpdateAt  = goqu.C("update_at")
	colUserID    = goqu.C("user_id")
	colVendor    = goqu.C("vendor")
	colVendorUID = goqu.C("vendor_uid")
)
