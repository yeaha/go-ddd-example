//go:build wireinject
// +build wireinject

package service

import "github.com/google/wire"

// ProviderSet is service providers.
var ProviderSet = wire.NewSet(
	wire.Struct(new(OauthTokenService), "*"),
	wire.Struct(new(SessionTokenService), "*"),
	wire.Struct(new(AccountService), "*"),
)
