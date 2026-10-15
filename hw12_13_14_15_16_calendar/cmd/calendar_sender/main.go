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

	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "failed to load config: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.New(configFile)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
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
		return fmt.Errorf("failed to connect to rabbitmq: %w", err)
	}

	defer func() {
		if err := rabbit.Close(); err != nil {
			logg.Error("failed to close rabbitmq: " + err.Error())
		}
	}()

	statusRabbit, err := rabbitmq.New(rabbitmq.Config{
		Host:     cfg.Rabbit.Host,
		Port:     cfg.Rabbit.Port,
		Username: cfg.Rabbit.Username,
		Password: cfg.Rabbit.Password,
		Queue:    cfg.Rabbit.StatusQueue,
	})
	if err != nil {
		return fmt.Errorf("failed to connect to rabbitmq status queue: %w", err)
	}

	defer func() {
		if err := statusRabbit.Close(); err != nil {
			logg.Error("failed to close rabbitmq status queue: " + err.Error())
		}
	}()

	s := sender.New(rabbit, statusRabbit, logg)

	ctx, cancel := signal.NotifyContext(
		context.Background(),
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGHUP,
	)
	defer cancel()

	logg.Info("calendar sender is running...")

	if err := s.Start(ctx); err != nil {
		return fmt.Errorf("calendar sender stopped with error: %w", err)
	}

	return nil
}
