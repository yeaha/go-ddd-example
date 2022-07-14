package main

import (
	"flag"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/joyparty/entity"
	"github.com/joyparty/entity/cache"
	"github.com/joyparty/httpkit"
	"github.com/sirupsen/logrus"
	"gitlab.haochang.tv/yangyi/examine-code/pkg/option"
	"gitlab.haochang.tv/yangyi/examine-code/pkg/presentation/httpapi"

	// postgresql database driver
	_ "github.com/jackc/pgx/v4/stdlib"

	// sql dialect
	_ "github.com/doug-martin/goqu/v9/dialect/postgres"
)

var (
	// 系统配置
	opt = &option.Options{}
)

func init() {
	flag.BoolVar(&opt.DevMode, "dev", false, "development mode")
	flag.BoolVar(&opt.LogPretty, "logPretty", false, "output pretty print log")
	flag.StringVar(&opt.ConfigFile, "config", "", "config file")
	flag.StringVar(&opt.LogLevel, "logLevel", "", "log level")
	flag.StringVar(&opt.MigratePath, "migrate", "", "database migrate file path")
	flag.Parse()

	initLogger(opt)

	if opt.ConfigFile == "" {
		logrus.Fatal("need config file")
	} else if err := opt.LoadFile(opt.ConfigFile); err != nil {
		logrus.WithError(err).Fatal("load config file")
	} else if err := opt.Prepare(); err != nil {
		logrus.WithError(err).Fatal("prepare resources")
	} else if err := dbMigrate(opt); err != nil {
		logrus.WithError(err).Fatal("database migrate")
	}

	// 实体对象，默认使用本地内存缓存
	entity.DefaultCacher = cache.NewMemoryCache()

	httpkit.RequestDecoder.SetAliasTag("json")
}

func initLogger(opt *option.Options) {
	logrus.SetReportCaller(true)
	logrus.SetFormatter(&logrus.JSONFormatter{
		PrettyPrint: opt.LogPretty,
	})
	logrus.SetLevel(logrus.DebugLevel)
}

func main() {
	server := httpapi.NewServer(opt)

	sc := make(chan os.Signal)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM)

	select {
	case s := <-sc:
		logrus.WithField("signal", s).Debug("receive signal")

		wg := &sync.WaitGroup{}

		wg.Add(1)
		server.Close(wg)

		wg.Wait()
		os.Exit(0)
	}
}