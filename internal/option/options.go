package option

import (
	"errors"
	"fmt"
	"io"
	"os"

	"ddd-example/internal/migrate"
	"ddd-example/pkg/database"
	"ddd-example/pkg/oauth"

	"github.com/BurntSushi/toml"
	"github.com/jmoiron/sqlx"
)

var (
	errOptionNotPrepare = errors.New("options not prepare")
)

// Options 系统配置
type Options struct {
	// 服务启动参数
	ConfigFile string `toml:"-"`
	LogLevel   string `toml:"-"`
	LogPretty  bool   `toml:"-"`

	HTTP struct {
		Port int `toml:"port"`
	} `toml:"http"`
	Database database.Option          `toml:"database"`
	Oauth    map[string]oauth.Options `toml:"oauth"`

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
	if err := migrate.Execute(migrate.FS, "scripts", opt.Database.DSN); err != nil {
		return fmt.Errorf("database migrate, %w", err)
	}

	db, err := database.NewDB(opt.Database)
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

// GetDB 获取数据库连接
func (opt *Options) GetDB() *sqlx.DB {
	if v := opt.clients.database; v != nil {
		return v
	}

	panic(errOptionNotPrepare)
}

// GetOauthClient 获取三方登录客户端
func (opt *Options) GetOauthClient(name string) (oauth.Client, bool) {
	client, ok := opt.clients.oauth[name]
	return client, ok
}
