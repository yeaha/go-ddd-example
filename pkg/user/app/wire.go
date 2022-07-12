//go:build wireinject
// +build wireinject

package app

import (
	"github.com/google/wire"
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
	)

	serviceSet = wire.NewSet(
		wire.Struct(new(service.SessionTokenService), "*"),
	)

	repositoriesProvider = wire.NewSet(
		repositoriesSet,
		wire.Struct(new(Repositories), "*"),
	)

	handersProvider = wire.NewSet(
		repositoriesSet,
		serviceSet,

		wire.Struct(new(handler.LoginWithEmailHandler), "*"),
		wire.Struct(new(handler.ChangePasswordHandler), "*"),
		wire.Struct(new(handler.LogoutHandler), "*"),
		wire.Struct(new(handler.RegisterHandler), "*"),
		wire.Struct(new(handler.RenewTokenHandler), "*"),
		wire.Struct(new(handler.RetrieveTokenHandler), "*"),
		wire.Struct(new(Handlers), "*"),
	)
)

func initRepositories(db entity.DB) Repositories {
	wire.Build(repositoriesProvider)
	return Repositories{}
}

func initHandlers(db entity.DB) Handlers {
	wire.Build(handersProvider)
	return Handlers{}
}
