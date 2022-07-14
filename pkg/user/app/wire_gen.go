// Code generated by Wire. DO NOT EDIT.

//go:generate go run github.com/google/wire/cmd/wire
//go:build !wireinject
// +build !wireinject

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

// Injectors from wire.go:

func initRepositories(dbi entity.DB) Repositories {
	userDBRepository := infra.NewUserDBRepository(dbi)
	repositories := Repositories{
		Users: userDBRepository,
	}
	return repositories
}

func initHandlers(db *sqlx.DB, dbi entity.DB, cache adapter.Cacher) Handlers {
	userDBRepository := infra.NewUserDBRepository(dbi)
	changePasswordHandler := &handler.ChangePasswordHandler{
		User: userDBRepository,
	}
	sessionTokenService := &service.SessionTokenService{
		Users: userDBRepository,
	}
	userService := &service.UserService{
		Users: userDBRepository,
	}
	loginWithEmailHandler := &handler.LoginWithEmailHandler{
		Session: sessionTokenService,
		Users:   userService,
	}
	logoutHandler := &handler.LogoutHandler{
		Session: sessionTokenService,
	}
	registerHandler := &handler.RegisterHandler{
		Session: sessionTokenService,
		Users:   userService,
	}
	oauthTokenService := &service.OauthTokenService{
		Cache: cache,
	}
	registerWithOauthHandler := &handler.RegisterWithOauthHandler{
		DB:         db,
		Session:    sessionTokenService,
		OauthToken: oauthTokenService,
	}
	renewTokenHandler := &handler.RenewTokenHandler{
		Session: sessionTokenService,
	}
	retrieveTokenHandler := &handler.RetrieveTokenHandler{
		Session: sessionTokenService,
	}
	oauthDBRepository := infra.NewOauthDBRepository(dbi)
	oauthService := &service.OauthService{
		Users: userDBRepository,
		Oauth: oauthDBRepository,
	}
	verifyOauthHandler := &handler.VerifyOauthHandler{
		Oauth:      oauthService,
		OauthToken: oauthTokenService,
		Session:    sessionTokenService,
	}
	handlers := Handlers{
		ChangePassword:    changePasswordHandler,
		LoginWithEmail:    loginWithEmailHandler,
		Logout:            logoutHandler,
		Register:          registerHandler,
		RegisterWithOauth: registerWithOauthHandler,
		RenewToken:        renewTokenHandler,
		RetrieveToken:     retrieveTokenHandler,
		VerifyOauth:       verifyOauthHandler,
	}
	return handlers
}

// wire.go:

var (
	repositoriesSet = wire.NewSet(wire.NewSet(infra.NewUserDBRepository, wire.Bind(new(adapter.UserRepository), new(*infra.UserDBRepository))), wire.NewSet(infra.NewOauthDBRepository, wire.Bind(new(adapter.OauthRepository), new(*infra.OauthDBRepository))),
	)

	serviceSet = wire.NewSet(wire.Struct(new(service.OauthService), "*"), wire.Struct(new(service.OauthTokenService), "*"), wire.Struct(new(service.SessionTokenService), "*"), wire.Struct(new(service.UserService), "*"))

	repositoriesProvider = wire.NewSet(
		repositoriesSet, wire.Struct(new(Repositories), "*"),
	)

	handlersProvider = wire.NewSet(
		repositoriesSet,
		serviceSet, wire.Struct(new(handler.ChangePasswordHandler), "*"), wire.Struct(new(handler.LoginWithEmailHandler), "*"), wire.Struct(new(handler.LogoutHandler), "*"), wire.Struct(new(handler.RegisterHandler), "*"), wire.Struct(new(handler.RegisterWithOauthHandler), "*"), wire.Struct(new(handler.RenewTokenHandler), "*"), wire.Struct(new(handler.RetrieveTokenHandler), "*"), wire.Struct(new(handler.VerifyOauthHandler), "*"), wire.Struct(new(Handlers), "*"),
	)
)
