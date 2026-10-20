package bootstrap

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"path/filepath"
	"time"

	"ddd-example/internal/app"
	"ddd-example/internal/infra"
	"ddd-example/internal/migrate"
	"ddd-example/internal/option"
	"ddd-example/internal/presentation/httpapi"
	"ddd-example/internal/presentation/observer"
	"ddd-example/pkg/database"
	"ddd-example/pkg/logger"
	"ddd-example/pkg/oauth"

	"github.com/jmoiron/sqlx"
	"github.com/samber/do/v2"

	// database driver
	_ "github.com/mattn/go-sqlite3"
)

// http服务超时配置
const (
	httpReadHeaderTimeout = 5 * time.Second
	httpReadTimeout       = 30 * time.Second
	httpWriteTimeout      = 30 * time.Second
	httpIdleTimeout       = 60 * time.Second
)

// App 应用实例
type App struct {
	db     *sqlx.DB
	server *http.Server
	addr   string
}

// New 构建应用
func New(opt *option.Options) (_ *App, err error) {
	if opt.DBDir == "" {
		return nil, errors.New("need database dir")
	}
	if opt.HTTP.Addr == "" {
		return nil, errors.New("need http addr")
	}

	db, err := database.NewDB(database.Option{
		Driver:       "sqlite3",
		DSN:          getDBDSN(opt.DBDir),
		MaxIdleConns: 3,
		MaxIdleTime:  10 * time.Minute,
	})
	if err != nil {
		return nil, fmt.Errorf("connect database, %w", err)
	}
	defer func() {
		if err != nil {
			_ = db.Close()
		}
	}()

	if err := migrate.Execute(context.Background(), db.DB, migrate.FS, "scripts"); err != nil {
		return nil, fmt.Errorf("database migrate, %w", err)
	}
	// 保持原有行为：infra 收到的是 sqlx 的 unsafe 版本
	db = db.Unsafe()

	clients, err := newOauthClients(opt.Oauth)
	if err != nil {
		return nil, err
	}

	injector := do.New(
		do.Eager(db),
		do.Eager(clients),
		infra.Providers,
		app.Providers,
		httpapi.Providers,
	)

	return &App{
		db:   db,
		addr: opt.HTTP.Addr,
		server: &http.Server{
			Handler:           httpapi.NewHandler(injector),
			ReadHeaderTimeout: httpReadHeaderTimeout,
			ReadTimeout:       httpReadTimeout,
			WriteTimeout:      httpWriteTimeout,
			IdleTimeout:       httpIdleTimeout,
		},
	}, nil
}

// Start 启动应用，返回实际监听地址
// ctx 仅用于事件观察者和日志，服务生命周期由Shutdown结束
func (a *App) Start(ctx context.Context) (_ string, err error) {
	defer func() {
		if err != nil {
			_ = a.db.Close()
		}
	}()

	l, err := net.Listen("tcp", a.addr)
	if err != nil {
		return "", fmt.Errorf("listen, %w", err)
	}
	a.addr = l.Addr().String()

	observer.Start(ctx)

	go func() {
		logger.Info(ctx, "start server", "listen", a.addr)
		if err := a.server.Serve(l); err != nil && !errors.Is(err, http.ErrServerClosed) {
			panic(fmt.Errorf("serve http, %w", err))
		}
	}()

	return a.addr, nil
}

// Shutdown 关闭应用
func (a *App) Shutdown() error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	err := a.server.Shutdown(ctx)
	if errors.Is(err, http.ErrServerClosed) {
		err = nil
	}
	if err == nil {
		logger.Info(context.Background(), "shutdown server")
	}

	observer.Stop()

	return errors.Join(err, a.db.Close())
}

// Run 启动应用并等待ctx结束
func (a *App) Run(ctx context.Context) error {
	if _, err := a.Start(ctx); err != nil {
		return err
	}

	<-ctx.Done()
	return a.Shutdown()
}

func getDBDSN(dir string) string {
	file := filepath.Join(dir, "main.db")
	return fmt.Sprintf("%s?mode=rwc&_timeout=5000&_journal=WAL&_sync=NORMAL&_encoding=UTF-8&_fk=1", file)
}

func newOauthClients(options map[string]oauth.Options) (oauth.Clients, error) {
	clients := make(oauth.Clients, len(options))
	for name, opt := range options {
		client, err := oauth.NewClient(name, &opt)
		if err != nil {
			return nil, fmt.Errorf("create oauth client, %q, %w", name, err)
		}
		clients[name] = client
	}
	return clients, nil
}
