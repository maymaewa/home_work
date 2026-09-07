package internalhttp

import (
	"context"
	"errors"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/maymaewa/home_work/hw12_13_14_15_calendar/internal/storage"
)

type Config struct {
	Host string
	Port int
}

type Server struct {
	logger Logger
	app    Application
	server *http.Server
}

type Logger interface {
	Debug(msg string)
	Info(msg string)
	Warn(msg string)
	Error(msg string)
	InfoContext(msg string, args ...any)
}

type Application interface {
	CreateEvent(ctx context.Context, event storage.Event) error
	UpdateEvent(ctx context.Context, event storage.Event) error
	DeleteEvent(ctx context.Context, id string) error

	ListEventsForDay(ctx context.Context, date time.Time) ([]storage.Event, error)
	ListEventsForWeek(ctx context.Context, date time.Time) ([]storage.Event, error)
	ListEventsForMonth(ctx context.Context, date time.Time) ([]storage.Event, error)
}

func NewServer(logger Logger, app Application, config Config) *Server {
	mux := http.NewServeMux()

	mux.HandleFunc("/hello", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("Hello, world!"))
	})

	return &Server{
		logger: logger,
		app:    app,
		server: &http.Server{
			Addr:              net.JoinHostPort(config.Host, strconv.Itoa(config.Port)),
			Handler:           loggingMiddleware(mux, logger),
			ReadHeaderTimeout: 5 * time.Second,
		},
	}
}

func (s *Server) Start(ctx context.Context) error {
	go func() {
		<-ctx.Done()

		if err := s.Stop(context.Background()); err != nil {
			s.logger.Error("failed to stop http server: " + err.Error())
		}
	}()

	err := s.server.ListenAndServe()
	if errors.Is(err, http.ErrServerClosed) {
		return nil
	}

	return err
}

func (s *Server) Stop(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}
