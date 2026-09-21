package logger

import (
	"log/slog"
	"os"
)

const (
	levelDebug = "debug"
	levelInfo  = "info"
	levelWarn  = "warn"
	levelError = "error"
)

type Logger struct {
	logger *slog.Logger
}

func New(level string) *Logger {
	var logLevel slog.Level

	switch level {
	case levelDebug:
		logLevel = slog.LevelDebug
	case levelInfo:
		logLevel = slog.LevelInfo
	case levelWarn:
		logLevel = slog.LevelWarn
	case levelError:
		logLevel = slog.LevelError
	default:
		logLevel = slog.LevelInfo
	}

	handler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: logLevel,
	})

	return &Logger{
		logger: slog.New(handler),
	}
}

func (l *Logger) Debug(msg string) {
	l.logger.Debug(msg)
}

func (l *Logger) Info(msg string) {
	l.logger.Info(msg)
}

func (l *Logger) Warn(msg string) {
	l.logger.Warn(msg)
}

func (l *Logger) Error(msg string) {
	l.logger.Error(msg)
}

func (l *Logger) InfoContext(msg string, args ...any) {
	l.logger.Info(msg, args...)
}
