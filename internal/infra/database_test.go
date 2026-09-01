//go:build dbtest

package infra

import (
	"context"
	"errors"
	"fmt"

	"ddd-example/internal/migrate"
	"ddd-example/pkg/database"

	"github.com/jmoiron/sqlx"
	"github.com/joyparty/entity"
	"github.com/joyparty/entity/cache"

	// database driver
	_ "github.com/mattn/go-sqlite3"
)

var (
	errRollbackTest = errors.New("rollback test")

	testDB *sqlx.DB
)

func init() {
	entity.DefaultCacher = cache.NewMemoryCache()

	db, err := database.NewDB(database.Option{
		Driver: "sqlite3",
		DSN:    ":memory:",
	})
	if err != nil {
		panic(fmt.Errorf("connect test main db, %w", err))
	}

	if err := migrate.Execute(context.Background(), db.DB, migrate.FS, "scripts"); err != nil {
		panic(fmt.Errorf("migrate test main db, %w", err))
	}
	testDB = db
}

type testTask struct {
	Name string
	Func func() error
}

type testTable []testTask

func (r testTable) Execute() error {
	for _, fn := range r {
		if err := fn.Func(); err != nil {
			return fmt.Errorf("%s: %w", fn.Name, err)
		}
	}
	return nil
}
