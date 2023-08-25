package app

import (
	"ddd-example/internal/app/adapter"
	"ddd-example/internal/app/handler"
	"ddd-example/internal/infra"
	"ddd-example/internal/option"
)

// Application 账号模块业务逻辑
type Application struct {
	AccountRepository adapter.AccountRepository

	Authorize         *handler.AuthorizeHandler
	ChangePassword    *handler.ChangePasswordHandler
	LoginWithEmail    *handler.LoginWithEmailHandler
	Logout            *handler.LogoutHandler
	Register          *handler.RegisterHandler
	RegisterWithOauth *handler.RegisterWithOauthHandler
	VerifyOauth       *handler.VerifyOauthHandler
}

// NewApplication 构造函数
func NewApplication(opt *option.Options) *Application {
	db := opt.GetDB()
	cache := infra.NewMemoryCache()

	return initApplication(db, db, cache)
}
