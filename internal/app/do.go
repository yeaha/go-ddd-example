package app

import (
	"ddd-example/internal/app/handler"
	"ddd-example/internal/app/service"

	"github.com/samber/do/v2"
)

// Providers 依赖注入配置
var Providers = do.Package(
	do.Lazy(do.InvokeStruct[*service.AccountService]),
	do.Lazy(do.InvokeStruct[*service.OauthTokenService]),
	do.Lazy(do.InvokeStruct[*service.SessionTokenService]),

	do.Lazy(do.InvokeStruct[*handler.AuthorizeHandler]),
	do.Lazy(do.InvokeStruct[*handler.ChangePasswordHandler]),
	do.Lazy(do.InvokeStruct[*handler.LoginWithEmailHandler]),
	do.Lazy(do.InvokeStruct[*handler.LogoutHandler]),
	do.Lazy(do.InvokeStruct[*handler.RegisterHandler]),
	do.Lazy(do.InvokeStruct[*handler.RegisterWithOauthHandler]),
	do.Lazy(do.InvokeStruct[*handler.VerifyOauthHandler]),
)
