package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/maymaewa/home_work/hw12_13_14_15_calendar/internal/app"
	"github.com/maymaewa/home_work/hw12_13_14_15_calendar/internal/config"
	"github.com/maymaewa/home_work/hw12_13_14_15_calendar/internal/logger"
	"github.com/maymaewa/home_work/hw12_13_14_15_calendar/internal/queue/rabbitmq"
	"github.com/maymaewa/home_work/hw12_13_14_15_calendar/internal/scheduler"
	sqlstorage "github.com/maymaewa/home_work/hw12_13_14_15_calendar/internal/storage/sql"
)

var configFile string

func init() {
	flag.StringVar(
		&configFile,
		"config",
		"/etc/calendar/config.toml",
		"Path to configuration file",
	)
}

func main() {
	flag.Parse()

	cfg, err := config.New(configFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load config: %v\n", err)
		os.Exit(1)
	}

	logg := logger.New(cfg.Logger.Level)

	sqlStore := sqlstorage.New(sqlstorage.Config{
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

	defer func() {
		if err := sqlStore.Close(context.Background()); err != nil {
			logg.Error("failed to close database: " + err.Error())
		}
	}()

	rabbit, err := rabbitmq.New(rabbitmq.Config{
		Host:     cfg.Rabbit.Host,
		Port:     cfg.Rabbit.Port,
		Username: cfg.Rabbit.Username,
		Password: cfg.Rabbit.Password,
		Queue:    cfg.Rabbit.Queue,
	})
	if err != nil {
		logg.Error("failed to connect to rabbitmq: " + err.Error())
		return
	}

	defer func() {
		if err := rabbit.Close(); err != nil {
			logg.Error("failed to close rabbitmq: " + err.Error())
		}
	}()

	calendar := app.New(logg, sqlStore)

	s := scheduler.New(
		calendar,
		rabbit,
		logg,
		cfg.Scheduler.Interval,
	)

	ctx, cancel := signal.NotifyContext(
		context.Background(),
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGHUP,
	)
	defer cancel()

	logg.Info("calendar scheduler is running...")

	if err := s.Start(ctx); err != nil {
		logg.Error("scheduler stopped with error: " + err.Error())
	}
}
