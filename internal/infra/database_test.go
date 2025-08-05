//go:build dbtest

package infra

import (
	"ddd-example/internal/migrate"
	"ddd-example/pkg/database"
	"fmt"
	"io/fs"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/joyparty/entity"
	"github.com/joyparty/entity/cache"

	// database driver
	_ "github.com/mattn/go-sqlite3"
)

var (
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
	} else if err := dbMigrate(migrate.FS, "scripts", db); err != nil {
		panic(fmt.Errorf("migrate test main db, %w", err))
	}
	testDB = db
}

func dbMigrate(source fs.FS, dir string, db *sqlx.DB) error {
	return fs.WalkDir(source, dir, func(path string, d fs.DirEntry, err error) error {
		if !d.IsDir() && strings.HasSuffix(d.Name(), ".up.sql") {
			content, err := migrate.FS.ReadFile(path)
			if err != nil {
				return fmt.Errorf("read migrate file, %w", err)
			}

			if _, err := db.Exec(string(content)); err != nil {
				return fmt.Errorf("exec migrate file, %s, %w", path, err)
			}
		}
		return nil
	})
}
