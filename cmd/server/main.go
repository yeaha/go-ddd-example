package main

import (
	"context"
	"flag"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"ddd-example/internal/app"
	"ddd-example/internal/infra"
	"ddd-example/internal/option"
	"ddd-example/internal/presentation/httpapi"
	"ddd-example/internal/presentation/observer"
	"ddd-example/pkg/logger"

	"github.com/joyparty/gokit"
	"github.com/samber/do/v2"
)

var (
	// 系统配置
	opt = &option.Options{}

	injector do.Injector
)

func init() {
	flag.StringVar(&opt.ConfigFile, "config", "", "config file")
	flag.StringVar(&opt.LogLevel, "logLevel", "", "log level")
	flag.StringVar(&opt.DBDir, "dbDir", "", "database dir")
	flag.Parse()

	slog.SetDefault(gokit.MustReturn(
		logger.New(logger.Option{
			Level:  opt.LogLevel,
			Format: "json",
		}),
	))

	if opt.ConfigFile == "" {
		logAndExist("need config file")
	} else if err := opt.LoadFile(opt.ConfigFile); err != nil {
		logAndExist("load config file", "error", err)
	} else if err := opt.Prepare(); err != nil {
		logAndExist("prepare resources", "error", err)
	}

	injector = do.New(
		opt.Providers(),
		infra.Providers,
		app.Providers,
		httpapi.Providers,
	)
}

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	server := do.MustInvoke[*httpapi.Server](injector)
	observer.Start(ctx, opt)

	<-ctx.Done()
	if err := server.Close(); err != nil {
		logger.Error(ctx, "shutdown server", "error", err)
	} else {
		logger.Info(ctx, "shutdown server")
	}
	observer.Stop()

	os.Exit(0)
}

func logAndExist(msg string, args ...any) {
	logger.Error(context.TODO(), msg, args...)
	os.Exit(1) // revive:disable-line
}
