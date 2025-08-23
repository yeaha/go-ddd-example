package option

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"time"

	"ddd-example/internal/migrate"
	"ddd-example/pkg/database"
	"ddd-example/pkg/oauth"

	"github.com/BurntSushi/toml"
	"github.com/jmoiron/sqlx"
	"github.com/samber/do/v2"

	// database driver
	_ "github.com/golang-migrate/migrate/v4/database/sqlite3"
	_ "github.com/mattn/go-sqlite3"
)

var errNotPrepare = errors.New("options not prepare")

// Options 系统配置
type Options struct {
	// 服务启动参数
	ConfigFile string `toml:"-"`
	DBDir      string `toml:"-"`
	LogLevel   string `toml:"-"`

	HTTP struct {
		Port int `toml:"port"`
	} `toml:"http"`
	Oauth map[string]oauth.Options `toml:"oauth"`

	clients struct {
		database *sqlx.DB
		oauth    map[string]oauth.Client
	}
}

// LoadFile 载入配置文件
func (opt *Options) LoadFile(file string) error {
	fh, err := os.Open(file)
	if err != nil {
		return err
	}
	defer fh.Close()

	data, err := io.ReadAll(fh)
	if err != nil {
		return err
	}
	return toml.Unmarshal(data, opt)
}

// Prepare 初始化资源
func (opt *Options) Prepare() error {
	if opt.DBDir == "" {
		return errors.New("need database dir")
	}

	if err := migrate.Execute(
		migrate.FS,
		"scripts",
		fmt.Sprintf("sqlite3://%s", opt.getDBDSN()),
	); err != nil {
		return fmt.Errorf("database migrate, %w", err)
	}

	db, err := database.NewDB(database.Option{
		Driver:       "sqlite3",
		DSN:          opt.getDBDSN(),
		MaxIdleConns: 3,
		MaxIdleTime:  10 * time.Minute,
	})
	if err != nil {
		return fmt.Errorf("connect database, %w", err)
	}
	opt.clients.database = db.Unsafe()

	opt.clients.oauth = make(map[string]oauth.Client)
	for name, options := range opt.Oauth {
		client, err := oauth.NewClient(name, &options)
		if err != nil {
			return fmt.Errorf("create oauth client, %q, %w", name, err)
		}
		opt.clients.oauth[name] = client
	}

	return nil
}

// Providers 提供依赖注入
func (opt *Options) Providers() func(do.Injector) {
	return do.Package(
		do.Eager(opt),
		do.Eager(opt.GetDB()),
	)
}

// getMainDSN 主数据库DSN
func (opt *Options) getDBDSN() string {
	file := filepath.Join(opt.DBDir, "main.db")
	return fmt.Sprintf("%s?mode=rwc&_timeout=5000&_journal=WAL&_sync=NORMAL&_encoding=UTF-8&_fk=1", file)
}

// GetDB 获取数据库连接
func (opt *Options) GetDB() *sqlx.DB {
	return mustNotNil(opt.clients.database)
}

// GetOauthClient 获取三方登录客户端
func (opt *Options) GetOauthClient(name string) (oauth.Client, bool) {
	client, ok := opt.clients.oauth[name]
	return client, ok
}

func mustNotNil[T any](v T) T {
	if reflect.ValueOf(v).IsNil() {
		panic(errNotPrepare)
	}

	return v
}
