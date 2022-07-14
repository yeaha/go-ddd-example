package app

import (
	"github.com/joyparty/entity/cache"
	"gitlab.haochang.tv/yangyi/examine-code/pkg/option"
	"gitlab.haochang.tv/yangyi/examine-code/pkg/user/app/adapter"
	"gitlab.haochang.tv/yangyi/examine-code/pkg/user/app/handler"
)

// Application 账号模块业务逻辑
type Application struct {
	Repositories Repositories
	Handlers     Handlers
}

// Handlers 业务命令
type Handlers struct {
	ChangePassword    *handler.ChangePasswordHandler
	LoginWithEmail    *handler.LoginWithEmailHandler
	Logout            *handler.LogoutHandler
	Register          *handler.RegisterHandler
	RegisterWithOauth *handler.RegisterWithOauthHandler
	RenewToken        *handler.RenewTokenHandler
	RetrieveToken     *handler.RetrieveTokenHandler
	VerifyOauth       *handler.VerifyOauthHandler
}

// Repositories 数据存储
type Repositories struct {
	Users adapter.UserRepository
}

// NewApplication 构造函数
func NewApplication(opt *option.Options) *Application {
	db := opt.GetDB()

	return &Application{
		Repositories: initRepositories(db),
		Handlers:     initHandlers(db, cache.NewMemoryCache()),
	}
}
