package httpapi

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	"gitlab.haochang.tv/yangyi/examine-code/pkg/option"
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
		logrus.WithField("listen", s.server.Addr).Info("start server")
		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logrus.WithError(err).Fatal("start server")
		}
	}()

	return s
}

// Close 关闭服务
func (s *Server) Close(wg *sync.WaitGroup) {
	defer wg.Done()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := s.server.Shutdown(ctx); err != nil && err != http.ErrServerClosed {
		logrus.WithError(err).Error("shutdown server")
		return
	}
	logrus.Info("shutdown server")
}
