package migrate

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"io/fs"

	"github.com/pressly/goose/v3"
)

// FS 数据库迁移文件
//
//go:embed scripts/*
var FS embed.FS

// Execute 执行数据库升级
func Execute(ctx context.Context, db *sql.DB, dir fs.FS, path string) error {
	sub, err := fs.Sub(dir, path)
	if err != nil {
		return fmt.Errorf("read dir, %w", err)
	}

	provider, err := goose.NewProvider(goose.DialectSQLite3, db, sub)
	if err != nil {
		return fmt.Errorf("new provider, %w", err)
	}

	if _, err := provider.Up(ctx); err != nil {
		return fmt.Errorf("database migrate, %w", err)
	}
	return nil
}
