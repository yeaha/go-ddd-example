package httpapi

import (
	"ddd-example/internal/option"
	"log/slog"

	"github.com/go-chi/chi/v5"
)

func newRouter(opt *option.Options) chi.Router {
	router := chi.NewRouter()

	router.Use(recoverer(slog.Default()))

	ac := newAuthController(opt)
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

	return router
}
