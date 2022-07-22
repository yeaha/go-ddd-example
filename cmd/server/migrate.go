package main

import (
	"errors"
	"fmt"
	"os"

	"ddd-example/pkg/option"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database"
	"github.com/sirupsen/logrus"

	// database migrate
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

// 升级数据库结构，需要确保只能有一个实例执行
func dbMigrate(opt *option.Options) error {
	if opt.MigratePath == "" {
		return nil
	}

	logrus.Info("migrate database schema")

	m, err := migrate.New(
		fmt.Sprintf("file://%s", opt.MigratePath),
		opt.Database.DSN,
	)
	if err != nil {
		return err
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
