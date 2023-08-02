package database

import (
	"time"

	"github.com/jmoiron/sqlx"
)

type (
	// Option 数据库连接配置
	Option struct {
		Driver       string        `toml:"driver"`
		DSN          string        `toml:"dsn"`
		MaxOpenConns int           `toml:"maxOpenConns"`
		MaxIdleConns int           `toml:"maxIdleConns"`
		MaxLifetime  time.Duration `toml:"maxLifetime"`
		MaxIdleTime  time.Duration `toml:"maxIdleTime"`
	}
)

// NewDB 构造数据库连接
func NewDB(opt Option) (*sqlx.DB, error) {
	db, err := sqlx.Connect(opt.Driver, opt.DSN)
	if err != nil {
		return nil, err
	}

	if n := opt.MaxOpenConns; n > 0 {
		db.SetMaxOpenConns(n)
	}
	if n := opt.MaxIdleConns; n > 0 {
		db.SetMaxIdleConns(n)
	}

	if t := opt.MaxLifetime; t > 0 {
		db.SetConnMaxLifetime(t)
	}
	if t := opt.MaxIdleTime; t > 0 {
		db.SetConnMaxIdleTime(t)
	}

	return db, nil
}
