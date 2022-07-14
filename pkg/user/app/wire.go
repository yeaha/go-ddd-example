//go:build wireinject
// +build wireinject

package app

import (
	"github.com/google/wire"
	"github.com/jmoiron/sqlx"
	"github.com/joyparty/entity"
	"gitlab.haochang.tv/yangyi/examine-code/pkg/user/app/adapter"
	"gitlab.haochang.tv/yangyi/examine-code/pkg/user/app/handler"
	"gitlab.haochang.tv/yangyi/examine-code/pkg/user/app/service"
	"gitlab.haochang.tv/yangyi/examine-code/pkg/user/infra"
)

var (
	repositoriesSet = wire.NewSet(
		wire.NewSet(
			infra.NewUserDBRepository,
			wire.Bind(new(adapter.UserRepository), new(*infra.UserDBRepository)),
		),
		wire.NewSet(
			infra.NewOauthDBRepository,
			wire.Bind(new(adapter.OauthRepository), new(*infra.OauthDBRepository)),
		),
	)

	serviceSet = wire.NewSet(
		wire.Struct(new(service.OauthService), "*"),
		wire.Struct(new(service.OauthTokenService), "*"),
		wire.Struct(new(service.SessionTokenService), "*"),
		wire.Struct(new(service.UserService), "*"),
	)

	repositoriesProvider = wire.NewSet(
		repositoriesSet,
		wire.Struct(new(Repositories), "*"),
	)

	handlersProvider = wire.NewSet(
		repositoriesSet,
		serviceSet,

		wire.Struct(new(handler.ChangePasswordHandler), "*"),
		wire.Struct(new(handler.LoginWithEmailHandler), "*"),
		wire.Struct(new(handler.LogoutHandler), "*"),
		wire.Struct(new(handler.RegisterHandler), "*"),
		wire.Struct(new(handler.RegisterWithOauthHandler), "*"),
		wire.Struct(new(handler.RenewTokenHandler), "*"),
		wire.Struct(new(handler.RetrieveTokenHandler), "*"),
		wire.Struct(new(handler.VerifyOauthHandler), "*"),
		wire.Struct(new(Handlers), "*"),
	)
)

func initRepositories(dbi entity.DB) Repositories {
	wire.Build(repositoriesProvider)
	return Repositories{}
}

func initHandlers(db *sqlx.DB, dbi entity.DB, cache adapter.Cacher) Handlers {
	wire.Build(handlersProvider)
	return Handlers{}
}
