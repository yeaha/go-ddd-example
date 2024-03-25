//go:build wireinject
// +build wireinject

package infra

import (
	"ddd-example/internal/app/adapter"

	"github.com/google/wire"
)

// ProviderSet is infra providers.
var ProviderSet = wire.NewSet(
	wire.NewSet(
		NewAccountDBRepository,
		wire.Bind(new(adapter.AccountRepository), new(*AccountDBRepository)),
	),
	wire.NewSet(
		NewOauthDBRepository,
		wire.Bind(new(adapter.OauthRepository), new(*OauthDBRepository)),
	),
)
