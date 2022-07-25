package app

import (
	"ddd-example/pkg/option"
	"ddd-example/pkg/user/app/adapter"
	"ddd-example/pkg/user/app/handler"
	"ddd-example/pkg/user/infra"
)

// Application 账号模块业务逻辑
type Application struct {
	Repositories Repositories

	ChangePassword       *handler.ChangePasswordHandler
	LoginWithEmail       *handler.LoginWithEmailHandler
	Logout               *handler.LogoutHandler
	Register             *handler.RegisterHandler
	RegisterWithOauth    *handler.RegisterWithOauthHandler
	RenewSessionToken    *handler.RenewSessionTokenHandler
	RetrieveSessionToken *handler.RetrieveSessionTokenHandler
	VerifyOauth          *handler.VerifyOauthHandler
}

// Repositories 数据存储
type Repositories struct {
	Users adapter.UserRepository
}

// NewApplication 构造函数
func NewApplication(opt *option.Options) *Application {
	db := opt.GetDB()
	cache := infra.NewMemoryCache()

	return initApplication(db, db, cache)
}
