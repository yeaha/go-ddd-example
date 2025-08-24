package httpapi

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"ddd-example/internal/option"
	"ddd-example/pkg/logger"

	"github.com/go-chi/chi/v5"
	"github.com/samber/do/v2"
)

// Server http服务
type Server struct {
	server *http.Server

	auth *authController
}

// ServerProvider 提供Server实例
func ServerProvider(injector do.Injector) (*Server, error) {
	s := &Server{
		auth: do.MustInvoke[*authController](injector),
	}

	opt := do.MustInvoke[*option.Options](injector)
	router := s.newRouter()
	s.server = &http.Server{
		Addr:    fmt.Sprintf(":%d", opt.HTTP.Port),
		Handler: router,
	}

	go func() {
		logger.Info(context.Background(), "start server", "listen", s.server.Addr)
		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error(context.Background(), "start server", "error", err)
		}
	}()

	return s, nil
}

// Close 关闭服务
func (s *Server) Close() error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := s.server.Shutdown(ctx); err != nil && err != http.ErrServerClosed {
		return err
	}
	return nil
}

func (s *Server) newRouter() chi.Router {
	router := chi.NewRouter()

	router.Use(recoverer)

	ac := s.auth
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
