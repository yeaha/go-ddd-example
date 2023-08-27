package migrate

import (
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"os"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

// FS 数据库迁移文件
//
//go:embed scripts/*
var FS embed.FS

// Execute 执行数据库升级
func Execute(dir fs.FS, path string, dsn string) error {
	drv, err := iofs.New(dir, path)
	if err != nil {
		return fmt.Errorf("read dir, %w", err)
	}

	m, err := migrate.NewWithSourceInstance("iofs", drv, dsn)
	if err != nil {
		return fmt.Errorf("new migrate, %w", err)
	}
	defer m.Close()

	if err := m.Up(); err != nil &&
		!errors.Is(err, migrate.ErrNoChange) &&
		!errors.Is(err, database.ErrLocked) &&
		!errors.Is(err, os.ErrNotExist) {
		return err
	}
	return nil
}
