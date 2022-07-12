package httpapi

import (
	"github.com/go-chi/chi/v5"
	"github.com/joyparty/httpkit"
	"github.com/sirupsen/logrus"
	"gitlab.haochang.tv/yangyi/examine-code/pkg/option"
)

func newRouter(opt *option.Options) chi.Router {
	router := chi.NewRouter()

	router.Use(httpkit.Recoverer(logrus.StandardLogger()))

	uc := newUserController(opt)
	router.Use(uc.Authorize)

	router.Post(`/login`, uc.LoginWithEmail())
	router.Post(`/register`, uc.Register())
	router.Delete(`/login`, uc.Logout())

	router.Group(func(router chi.Router) {
		router.Use(uc.DenyAnonymous)

		router.Get(`/my/identity`, uc.MyIdentity())
		router.Put(`/my/password`, uc.ChangePassword())
	})

	return router
}
