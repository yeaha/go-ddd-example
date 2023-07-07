package httpapi

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"ddd-example/pkg/option"

	"golang.org/x/exp/slog"
)

// Server http服务
type Server struct {
	opt    *option.Options
	server *http.Server
}

// NewServer 构造http服务并启动
func NewServer(opt *option.Options) *Server {
	s := &Server{
		opt: opt,
		server: &http.Server{
			Addr:    fmt.Sprintf(":%d", opt.HTTP.Port),
			Handler: newRouter(opt),
		},
	}

	go func() {
		slog.Info("start server", "listen", s.server.Addr)
		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("start server", "error", err)
		}
	}()

	return s
}

// Close 关闭服务
func (s *Server) Close(wg *sync.WaitGroup) error {
	defer wg.Done()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := s.server.Shutdown(ctx); err != nil && err != http.ErrServerClosed {
		return err
	}
	return nil
}
