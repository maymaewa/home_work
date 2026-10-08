package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/maymaewa/home_work/hw12_13_14_15_calendar/internal/app"
	"github.com/maymaewa/home_work/hw12_13_14_15_calendar/internal/config"
	"github.com/maymaewa/home_work/hw12_13_14_15_calendar/internal/logger"
	internalgrpc "github.com/maymaewa/home_work/hw12_13_14_15_calendar/internal/server/grpc"
	internalhttp "github.com/maymaewa/home_work/hw12_13_14_15_calendar/internal/server/http"
	memorystorage "github.com/maymaewa/home_work/hw12_13_14_15_calendar/internal/storage/memory"
	sqlstorage "github.com/maymaewa/home_work/hw12_13_14_15_calendar/internal/storage/sql"
)

var configFile string

func init() {
	flag.StringVar(&configFile, "config", "/etc/calendar/config.toml", "Path to configuration file")
}

func main() {
	flag.Parse()

	if flag.Arg(0) == "version" {
		printVersion()
		return
	}

	cfg, err := config.New(configFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load config: %v\n", err)
		os.Exit(1)
	}

	logg := logger.New(cfg.Logger.Level)

	var storage app.Storage
	var sqlStore *sqlstorage.Storage

	switch cfg.Storage.Type {
	case "memory":
		storage = memorystorage.New()

	case "sql":
		sqlStore = sqlstorage.New(sqlstorage.Config{
			Host:     cfg.Storage.SQL.Host,
			Port:     cfg.Storage.SQL.Port,
			Database: cfg.Storage.SQL.Database,
			Username: cfg.Storage.SQL.Username,
			Password: cfg.Storage.SQL.Password,
		})

		if err := sqlStore.Connect(context.Background()); err != nil {
			logg.Error("failed to connect to database: " + err.Error())
			os.Exit(1)
		}

		storage = sqlStore

	default:
		logg.Error("unknown storage type: " + cfg.Storage.Type)
		os.Exit(1)
	}

	if sqlStore != nil {
		defer func() {
			if err := sqlStore.Close(context.Background()); err != nil {
				logg.Error("failed to close database: " + err.Error())
			}
		}()
	}

	calendar := app.New(logg, storage)

	httpServer := internalhttp.NewServer(
		logg,
		calendar,
		internalhttp.Config{
			Host: cfg.HTTP.Host,
			Port: cfg.HTTP.Port,
		},
	)

	grpcServer := internalgrpc.NewServer(
		logg,
		calendar,
		internalgrpc.Config{
			Host: cfg.GRPC.Host,
			Port: cfg.GRPC.Port,
		},
	)

	ctx, cancel := signal.NotifyContext(
		context.Background(),
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGHUP,
	)
	defer cancel()

	go func() {
		<-ctx.Done()

		shutdownCtx, shutdownCancel := context.WithTimeout(
			context.Background(),
			3*time.Second,
		)
		defer shutdownCancel()

		if err := httpServer.Stop(shutdownCtx); err != nil {
			logg.Error("failed to stop http server: " + err.Error())
		}
	}()

	logg.Info("calendar is running...")

	errCh := make(chan error, 2)

	go func() {
		errCh <- httpServer.Start(ctx)
	}()

	go func() {
		errCh <- grpcServer.Start(ctx)
	}()

	if err := <-errCh; err != nil {
		logg.Error("server stopped with error: " + err.Error())
		cancel()
		os.Exit(1) //nolint:gocritic
	}
}
