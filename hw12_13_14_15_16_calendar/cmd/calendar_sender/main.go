package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/maymaewa/home_work/hw12_13_14_15_calendar/internal/config"
	"github.com/maymaewa/home_work/hw12_13_14_15_calendar/internal/logger"
	"github.com/maymaewa/home_work/hw12_13_14_15_calendar/internal/queue/rabbitmq"
	"github.com/maymaewa/home_work/hw12_13_14_15_calendar/internal/sender"
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

	rabbit, err := rabbitmq.New(rabbitmq.Config{
		Host:     cfg.Rabbit.Host,
		Port:     cfg.Rabbit.Port,
		Username: cfg.Rabbit.Username,
		Password: cfg.Rabbit.Password,
		Queue:    cfg.Rabbit.Queue,
	})
	if err != nil {
		logg.Error("failed to connect to rabbitmq: " + err.Error())
		os.Exit(1)
	}

	defer func() {
		if err := rabbit.Close(); err != nil {
			logg.Error("failed to close rabbitmq: " + err.Error())
		}
	}()

	s := sender.New(rabbit, logg)

	ctx, cancel := signal.NotifyContext(
		context.Background(),
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGHUP,
	)
	defer cancel()

	logg.Info("calendar sender is running...")

	if err := s.Start(ctx); err != nil {
		logg.Error("scheduler stopped with error: " + err.Error())
	}
}
