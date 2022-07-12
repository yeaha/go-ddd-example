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

	colEmail = goqu.C("email")
	colID    = goqu.C("id")
)
