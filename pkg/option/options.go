package option

import (
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/BurntSushi/toml"
	"github.com/jmoiron/sqlx"
	"gitlab.haochang.tv/yangyi/examine-code/pkg/utils/database"
)

var (
	errOptionNotPrepare = errors.New("options not prepare")
)

// Options 系统配置
type Options struct {
	// 服务启动参数
	ConfigFile  string `toml:"-"`
	DevMode     bool   `toml:"-"`
	LogLevel    string `toml:"-"`
	LogPretty   bool   `toml:"-"`
	MigratePath string `toml:"-"`

	HTTP struct {
		Port int `toml:"port"`
	} `toml:"http"`
	Database database.Option `toml:"database"`

	clients struct {
		database *sqlx.DB
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
	db, err := database.NewDB(opt.Database)
	if err != nil {
		return fmt.Errorf("connect database, %w", err)
	}
	opt.clients.database = db.Unsafe()

	return nil
}

// GetDB 获取数据库连接
func (opt *Options) GetDB() *sqlx.DB {
	if v := opt.clients.database; v != nil {
		return v
	}

	panic(errOptionNotPrepare)
}
