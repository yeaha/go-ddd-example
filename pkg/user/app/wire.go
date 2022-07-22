//go:build wireinject
// +build wireinject

package app

import (
	"ddd-example/pkg/user/app/adapter"
	"ddd-example/pkg/user/app/handler"
	"ddd-example/pkg/user/app/service"
	"ddd-example/pkg/user/infra"

	"github.com/google/wire"
	"github.com/jmoiron/sqlx"
	"github.com/joyparty/entity"
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
		wire.Struct(new(service.OauthUserService), "*"),
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
		wire.Struct(new(handler.RenewSessionTokenHandler), "*"),
		wire.Struct(new(handler.RetrieveSessionTokenHandler), "*"),
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
