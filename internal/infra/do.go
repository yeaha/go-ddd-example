package infra

import "github.com/samber/do/v2"

// Providers 依赖注入配置
var Providers = do.Package(
	do.Eager(NewMemoryCache()),

	do.Lazy(AccountRepositoryProvider),
	do.Lazy(OauthRepositoryProvider),
)
