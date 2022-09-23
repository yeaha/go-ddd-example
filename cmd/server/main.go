package main

import (
	"context"
	"embed"
	"flag"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"ddd-example/pkg/option"
	"ddd-example/pkg/presentation/httpapi"
	"ddd-example/pkg/presentation/observer"
	"ddd-example/pkg/utils/database"

	"github.com/joyparty/entity"
	"github.com/joyparty/entity/cache"
	"github.com/sirupsen/logrus"

	// database driver
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/jackc/pgx/v4/stdlib"

	// sql dialect
	_ "github.com/doug-martin/goqu/v9/dialect/postgres"
)

var (
	// 系统配置
	opt = &option.Options{}

	//go:embed migrate/*
	migrateFiles embed.FS
)

func init() {
	flag.BoolVar(&opt.LogPretty, "logPretty", false, "output pretty print log")
	flag.StringVar(&opt.ConfigFile, "config", "", "config file")
	flag.StringVar(&opt.LogLevel, "logLevel", "", "log level")
	flag.Parse()

	initLogger(opt)

	if opt.ConfigFile == "" {
		logrus.Fatal("need config file")
	} else if err := opt.LoadFile(opt.ConfigFile); err != nil {
		logrus.WithError(err).Fatal("load config file")
	} else if err := opt.Prepare(); err != nil {
		logrus.WithError(err).Fatal("prepare resources")
	} else if err := database.Migrate(migrateFiles, "migrate", opt.Database.DSN); err != nil {
		logrus.WithError(err).Fatal("database migrate")
	}

	// 实体对象，默认使用本地内存缓存
	entity.DefaultCacher = cache.NewMemoryCache()
}

func initLogger(opt *option.Options) {
	logrus.SetReportCaller(true)
	logrus.SetFormatter(&logrus.JSONFormatter{
		PrettyPrint: opt.LogPretty,
	})

	if v := opt.LogLevel; v != "" {
		if lvl, err := logrus.ParseLevel(v); err != nil {
			logrus.WithError(err).Fatal("parse logLevel")
		} else {
			logrus.SetLevel(lvl)
		}
	}
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
		logrus.WithField("signal", s).Debug("receive signal")

		wg := &sync.WaitGroup{}

		wg.Add(1)
		if err := server.Close(wg); err != nil {
			logrus.WithError(err).Error("shutdown server")
		} else {
			logrus.Info("shutdown server")
		}

		wg.Add(1)
		observer.Stop(wg)

		wg.Wait()
		os.Exit(0)
	}
}
