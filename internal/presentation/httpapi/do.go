package httpapi

import "github.com/samber/do/v2"

// Providers 依赖注入配置
var Providers = do.Package(
	do.Lazy(do.InvokeStruct[*authController]),

	do.Lazy(ServerProvider),
)
