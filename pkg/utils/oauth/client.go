package oauth

import (
	"fmt"
	"net/http"
	"net/url"
	"time"
)

var (
	httpClient = &http.Client{
		Timeout: 10 * time.Second,
	}
)

// Client 客户端
type Client interface {
	AuthorizeURL(redirectURI string) *url.URL
	Authorize(code string, redirectURI string) (*User, error)
}

// Options 配置
type Options struct {
	ClientID     string
	ClientSecret string
}

// Result http handler result
type Result struct {
	User    *User
	NextURL *url.URL
}

// User 三方平台用户信息
type User struct {
	AccessToken string `json:"-"`
	ID          string `json:"id"`
}

// NewClient 构造函数
func NewClient(site string, opt *Options) (Client, error) {
	switch site {
	case "facebook":
		return &facebook{opt: opt}, nil
	default:
		return nil, fmt.Errorf("unknown oauth site %q", site)
	}
}

// Handle oauth验证
// 第一步：code参数为空，把用户重定向至authorizeURL
// 第二步：用户在authorizeURL认证完毕后，会返回当前url，附带参数 ?code=xxxx
// 第三步：服务器端，用code对accessURL发起请求，获取access token
// 第四步：使用access token请求用户信息并返回
func Handle(c Client, r *http.Request) (*Result, error) {
	// 当前请求的url作为默认redirect_uri
	redirectURI := fmt.Sprintf("%s://%s%s", r.URL.Scheme, r.Host, r.URL.Path)

	query := r.URL.Query()
	// 前后端分离时，当前接口的url和前端的url可能是不一致的
	// 所以允许前端请求时指定redirect_uri
	if s := query.Get("redirect_uri"); s != "" {
		if l, err := url.Parse(s); err == nil && l.Host == r.URL.Host {
			redirectURI = l.String()
		}
	}

	code := query.Get("code")
	if code == "" { // step1
		return &Result{
			NextURL: c.AuthorizeURL(redirectURI),
		}, nil
	}

	// step3 ~ 4
	u, err := c.Authorize(code, redirectURI)
	if err != nil {
		return nil, fmt.Errorf("get access token, %w", err)
	}
	return &Result{User: u}, nil
}
