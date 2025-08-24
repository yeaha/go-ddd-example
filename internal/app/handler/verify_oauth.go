package handler

import (
	"context"
	"errors"
	"fmt"
	"net/url"

	"ddd-example/internal/app/adapter"
	"ddd-example/internal/app/internal/service"
	"ddd-example/internal/domain"
	"ddd-example/pkg/oauth"
)

// VerifyOauth 三方登录验证，参数
type VerifyOauth struct {
	Client      oauth.Client `json:"-"`
	RedirectURI string       `json:"redirect_uri" validate:"http_url"`
	// 从三方验证完毕重定向回来时，附带的url query
	RawQuery string     `json:"query" validate:"required"`
	Query    url.Values `json:"-"`
}

// VerifyOauthResult 三方登录验证结果
type VerifyOauthResult struct {
	// 三方验证完毕后，找到的对应系统账号
	// 在没有绑定关系的情况下，会是空值
	Account *domain.Account
	// 会话凭证
	SessionToken string
	// 三方账号的缓存凭证，后续使用这个凭证可以注册新账号或者绑定已有账号
	OauthToken string
}

// VerifyOauthHandler 三方登录验证
type VerifyOauthHandler struct {
	OauthToken *service.OauthTokenService   `do:""`
	Session    *service.SessionTokenService `do:""`
	Oauth      adapter.OauthRepository      `do:""`
	Accounts   adapter.AccountRepository    `do:""`
}

// Handle 验证三方登录
func (h *VerifyOauthHandler) Handle(ctx context.Context, args VerifyOauth) (result VerifyOauthResult, err error) {
	code := args.Query.Get("code")
	if code == "" {
		err = errors.New("empty oauth code")
		return
	}

	vendorUser, err := args.Client.Authorize(code, args.RedirectURI)
	if err != nil {
		err = fmt.Errorf("verify by vendor, %w", err)
		return
	}
	vendorUser.Vendor = args.Client.Vendor()

	account, err := h.findAccount(ctx, vendorUser)
	if errors.Is(err, domain.ErrAccountNotFound) {
		var oauthToken string

		oauthToken, err = h.OauthToken.Save(ctx, vendorUser)
		if err != nil {
			err = fmt.Errorf("cache vendor user, %w", err)
			return
		}
		result.OauthToken = oauthToken
		return
	} else if err != nil {
		err = fmt.Errorf("find account by vendor uid, %w", err)
		return
	}

	sessionToken, err := h.Session.Generate(ctx, account)
	if err != nil {
		err = fmt.Errorf("generate session token, %w", err)
		return
	}

	result.Account = account
	result.SessionToken = sessionToken
	return
}

func (h *VerifyOauthHandler) findAccount(ctx context.Context, vendorUser *oauth.User) (*domain.Account, error) {
	accountID, err := h.Oauth.Find(ctx, vendorUser.Vendor, vendorUser.ID)
	if err != nil {
		return nil, fmt.Errorf("find account id, %w", err)
	}

	return h.Accounts.Find(ctx, accountID)
}
