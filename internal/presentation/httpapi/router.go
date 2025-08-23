package httpapi

import (
	"log/slog"

	"github.com/go-chi/chi/v5"
	"github.com/samber/do/v2"
)

func routerProvider(injector do.Injector) (chi.Router, error) {
	router := chi.NewRouter()

	router.Use(recoverer(slog.Default()))

	ac := do.MustInvokeStruct[*authController](injector)
	router.Use(ac.Authorize)

	router.Post(`/session`, ac.LoginWithEmail())
	router.Post(`/register`, ac.Register())
	router.Delete(`/session`, ac.Logout())
	router.Get(`/login/oauth/{site}`, ac.LoginWithOauth())
	router.Post(`/login/oauth/{site}`, ac.VerifyOauth())
	router.Post(`/register/oauth`, ac.RegisterWithOauth())

	router.Group(func(router chi.Router) {
		router.Use(ac.DenyAnonymous)

		router.Get(`/session`, ac.MyIdentity())
		router.Put(`/my/password`, ac.ChangePassword())
	})

	return router, nil
}
