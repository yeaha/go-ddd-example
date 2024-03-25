//go:build wireinject
// +build wireinject

package app

import (
	"ddd-example/internal/app/adapter"
	"ddd-example/internal/app/handler"
	"ddd-example/internal/app/service"
	"ddd-example/internal/infra"

	"github.com/google/wire"
	"github.com/jmoiron/sqlx"
	"github.com/joyparty/entity"
)

func initApplication(db *sqlx.DB, dbi entity.DB, cache adapter.Cacher) *Application {
	wire.Build(wire.NewSet(
		infra.ProviderSet,
		service.ProviderSet,

		wire.Struct(new(handler.AuthorizeHandler), "*"),
		wire.Struct(new(handler.ChangePasswordHandler), "*"),
		wire.Struct(new(handler.LoginWithEmailHandler), "*"),
		wire.Struct(new(handler.LogoutHandler), "*"),
		wire.Struct(new(handler.RegisterHandler), "*"),
		wire.Struct(new(handler.RegisterWithOauthHandler), "*"),
		wire.Struct(new(handler.VerifyOauthHandler), "*"),

		wire.Struct(new(Application), "*"),
	))
	return &Application{}
}
