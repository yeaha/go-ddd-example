package option

import (
	"io"
	"os"

	"ddd-example/pkg/oauth"

	"github.com/BurntSushi/toml"
)

// Options 系统配置
type Options struct {
	// 服务启动参数
	ConfigFile string `toml:"-"`
	DBDir      string `toml:"-"`
	LogLevel   string `toml:"-"`

	HTTP struct {
		Addr string `toml:"addr"`
	} `toml:"http"`
	Oauth map[string]oauth.Options `toml:"oauth"`
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
