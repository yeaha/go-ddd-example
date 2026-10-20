package main

import (
	"context"
	"errors"
	"flag"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"ddd-example/internal/bootstrap"
	"ddd-example/internal/option"
	"ddd-example/pkg/logger"

	"github.com/joyparty/gokit"
)

func main() {
	opt := &option.Options{}
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

	if err := run(opt); err != nil {
		logger.Error(context.Background(), "run server", "error", err)
		os.Exit(1)
	}
}

func run(opt *option.Options) error {
	if opt.ConfigFile == "" {
		return errors.New("need config file")
	}
	if err := opt.LoadFile(opt.ConfigFile); err != nil {
		return err
	}

	app, err := bootstrap.New(opt)
	if err != nil {
		return err
	}

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	return app.Run(ctx)
}
