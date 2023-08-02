package oauth

import (
	"errors"
	"net/http"
	"net/url"
	"time"
)

var (
	httpClient = &http.Client{
		Timeout: 10 * time.Second,
	}

	// ErrNotImplemented 未实现的站点
	ErrNotImplemented = errors.New("not implemented")
)

// Client 客户端
type Client interface {
	AuthorizeURL(redirectURI string) *url.URL
	Authorize(code string, redirectURI string) (*User, error)
	Vendor() string
}

// Options 配置
type Options struct {
	ClientID     string `toml:"client_id"`
	ClientSecret string `toml:"client_secret"`
}

// User 三方平台用户信息
type User struct {
	Vendor      string `json:"vendor"`
	AccessToken string `json:"access_token"`
	ID          string `json:"id"`
}

// NewClient 构造函数
func NewClient(site string, opt *Options) (Client, error) {
	switch site {
	case "facebook":
		return &facebook{opt: opt}, nil
	default:
		return nil, ErrNotImplemented
	}
}
