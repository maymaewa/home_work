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

	config, err := NewConfig(configFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load config: %v\n", err)
		os.Exit(1)
	}

	logg := logger.New(config.Logger.Level)

	var storage app.Storage
	var sqlStore *sqlstorage.Storage

	switch config.Storage.Type {
	case "memory":
		storage = memorystorage.New()

	case "sql":
		sqlStore = sqlstorage.New(sqlstorage.Config{
			Host:     config.Storage.SQL.Host,
			Port:     config.Storage.SQL.Port,
			Database: config.Storage.SQL.Database,
			Username: config.Storage.SQL.Username,
			Password: config.Storage.SQL.Password,
		})

		if err := sqlStore.Connect(context.Background()); err != nil {
			logg.Error("failed to connect to database: " + err.Error())
			os.Exit(1)
		}

		storage = sqlStore

	default:
		logg.Error("unknown storage type: " + config.Storage.Type)
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
			Host: config.HTTP.Host,
			Port: config.HTTP.Port,
		},
	)

	grpcServer := internalgrpc.NewServer(
		logg,
		calendar,
		internalgrpc.Config{
			Host: config.GRPC.Host,
			Port: config.GRPC.Port,
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
