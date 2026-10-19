package httpapi

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/samber/do/v2"
)

// NewHandler 创建http handler
func NewHandler(injector do.Injector) http.Handler {
	return newRouter(do.MustInvoke[*authController](injector))
}

func newRouter(ac *authController) chi.Router {
	router := chi.NewRouter()

	router.Use(recoverer)

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
