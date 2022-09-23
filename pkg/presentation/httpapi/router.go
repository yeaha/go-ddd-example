package httpapi

import (
	"ddd-example/pkg/option"

	"github.com/go-chi/chi/v5"
	"github.com/sirupsen/logrus"
)

func newRouter(opt *option.Options) chi.Router {
	router := chi.NewRouter()

	router.Use(recoverer(logrus.StandardLogger()))

	uc := newUserController(opt)
	router.Use(uc.Authorize)

	router.Post(`/login`, uc.LoginWithEmail())
	router.Post(`/register`, uc.Register())
	router.Delete(`/login`, uc.Logout())
	router.Get(`/login/oauth/{site}`, uc.LoginWithOauth())
	router.Post(`/login/oauth/{site}`, uc.VerifyOauth())
	router.Post(`/register/oauth`, uc.RegisterWithOauth())

	router.Group(func(router chi.Router) {
		router.Use(uc.DenyAnonymous)

		router.Get(`/my/identity`, uc.MyIdentity())
		router.Put(`/my/password`, uc.ChangePassword())
	})

	return router
}
