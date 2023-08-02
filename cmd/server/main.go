package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"ddd-example/internal/option"
	"ddd-example/internal/presentation/httpapi"
	"ddd-example/internal/presentation/observer"
	"ddd-example/pkg/logger"

	"github.com/joyparty/entity"
	"github.com/joyparty/entity/cache"
	"golang.org/x/exp/slog"

	// database driver
	_ "github.com/jackc/pgx/v4/stdlib"
)

var (
	// 系统配置
	opt = &option.Options{}
)

func init() {
	flag.BoolVar(&opt.LogPretty, "logPretty", false, "output pretty print log")
	flag.StringVar(&opt.ConfigFile, "config", "", "config file")
	flag.StringVar(&opt.LogLevel, "logLevel", "", "log level")
	flag.Parse()

	initLogger(opt)

	if opt.ConfigFile == "" {
		logger.ErrorAndExist(context.Background(), "need config file")
	} else if err := opt.LoadFile(opt.ConfigFile); err != nil {
		logger.ErrorAndExist(context.Background(), "load config file", "error", err)
	} else if err := opt.Prepare(); err != nil {
		logger.ErrorAndExist(context.Background(), "prepare resources", "error", err)
	}

	// 实体对象，默认使用本地内存缓存
	entity.DefaultCacher = cache.NewMemoryCache()
}

func initLogger(opt *option.Options) {
	levels := map[string]slog.Level{
		"debug": slog.LevelDebug,
		"info":  slog.LevelInfo,
		"warn":  slog.LevelWarn,
		"error": slog.LevelError,
	}

	level := slog.LevelInfo
	if v, ok := levels[opt.LogLevel]; ok {
		level = v
	}

	slog.SetDefault(slog.New(
		slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{
			Level: level,
		}),
	))
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	server := httpapi.NewServer(opt)

	// 领域事件
	observer.Start(ctx, opt)

	sc := make(chan os.Signal)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM)

	select {
	case s := <-sc:
		slog.Debug("receive signal", "signal", s)

		wg := &sync.WaitGroup{}

		wg.Add(1)
		if err := server.Close(wg); err != nil {
			slog.Error("shutdown server", "error", err)
		} else {
			slog.Info("shutdown server")
		}

		wg.Add(1)
		observer.Stop(wg)

		wg.Wait()
		os.Exit(0)
	}
}
