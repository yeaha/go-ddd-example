// Code generated by Wire. DO NOT EDIT.

//go:generate go run github.com/google/wire/cmd/wire
//go:build !wireinject
// +build !wireinject

package app

import (
	"github.com/google/wire"
	"github.com/joyparty/entity"
	"gitlab.haochang.tv/yangyi/examine-code/pkg/user/app/adapter"
	"gitlab.haochang.tv/yangyi/examine-code/pkg/user/app/handler"
	"gitlab.haochang.tv/yangyi/examine-code/pkg/user/app/service"
	"gitlab.haochang.tv/yangyi/examine-code/pkg/user/infra"
)

// Injectors from wire.go:

func initRepositories(db entity.DB) Repositories {
	userDBRepository := infra.NewUserDBRepository(db)
	repositories := Repositories{
		Users: userDBRepository,
	}
	return repositories
}

func initHandlers(db entity.DB) Handlers {
	userDBRepository := infra.NewUserDBRepository(db)
	changePasswordHandler := &handler.ChangePasswordHandler{
		User: userDBRepository,
	}
	sessionTokenService := &service.SessionTokenService{
		Users: userDBRepository,
	}
	renewTokenHandler := &handler.RenewTokenHandler{
		Session: sessionTokenService,
	}
	loginWithEmailHandler := &handler.LoginWithEmailHandler{
		Users:   userDBRepository,
		Session: sessionTokenService,
	}
	logoutHandler := &handler.LogoutHandler{
		Session: sessionTokenService,
	}
	registerHandler := &handler.RegisterHandler{
		User:    userDBRepository,
		Session: sessionTokenService,
	}
	retrieveTokenHandler := &handler.RetrieveTokenHandler{
		Session: sessionTokenService,
	}
	handlers := Handlers{
		ChangePassword: changePasswordHandler,
		RenewToken:     renewTokenHandler,
		LoginWithEmail: loginWithEmailHandler,
		Logout:         logoutHandler,
		Register:       registerHandler,
		RetrieveToken:  retrieveTokenHandler,
	}
	return handlers
}

// wire.go:

var (
	repositoriesSet = wire.NewSet(wire.NewSet(infra.NewUserDBRepository, wire.Bind(new(adapter.UserRepository), new(*infra.UserDBRepository))),
	)

	serviceSet = wire.NewSet(wire.Struct(new(service.SessionTokenService), "*"))

	repositoriesProvider = wire.NewSet(
		repositoriesSet, wire.Struct(new(Repositories), "*"),
	)

	handlersProvider = wire.NewSet(
		repositoriesSet,
		serviceSet, wire.Struct(new(handler.LoginWithEmailHandler), "*"), wire.Struct(new(handler.ChangePasswordHandler), "*"), wire.Struct(new(handler.LogoutHandler), "*"), wire.Struct(new(handler.RegisterHandler), "*"), wire.Struct(new(handler.RenewTokenHandler), "*"), wire.Struct(new(handler.RetrieveTokenHandler), "*"), wire.Struct(new(Handlers), "*"),
	)
)
