package logger

import (
	"bytes"
	"context"
	"log/slog"
	"strings"
	"testing"
)

func TestNew_DefaultLevel(t *testing.T) {
	logger := New("unknown")

	innerLogger := logger.logger

	if innerLogger == nil {
		t.Fatal("inner logger is nil")
	}

	if !innerLogger.Enabled(context.TODO(), slog.LevelInfo) {
		t.Fatal("expected info level to be enabled")
	}

	if innerLogger.Enabled(context.TODO(), slog.LevelDebug) {
		t.Fatal("expected debug level to be disabled")
	}
}

func TestNew_Levels(t *testing.T) {
	tests := []struct {
		name        string
		level       string
		debugEnable bool
		infoEnable  bool
		warnEnable  bool
		errorEnable bool
	}{
		{
			name:        levelDebug,
			level:       levelDebug,
			debugEnable: true,
			infoEnable:  true,
			warnEnable:  true,
			errorEnable: true,
		},
		{
			name:        levelInfo,
			level:       levelInfo,
			debugEnable: false,
			infoEnable:  true,
			warnEnable:  true,
			errorEnable: true,
		},
		{
			name:        levelWarn,
			level:       levelWarn,
			debugEnable: false,
			infoEnable:  false,
			warnEnable:  true,
			errorEnable: true,
		},
		{
			name:        levelError,
			level:       levelError,
			debugEnable: false,
			infoEnable:  false,
			warnEnable:  false,
			errorEnable: true,
		},
		{
			name:        "unknown",
			level:       "unknown",
			debugEnable: false,
			infoEnable:  true,
			warnEnable:  true,
			errorEnable: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := New(tt.level)

			if got := logger.logger.Enabled(context.TODO(), slog.LevelDebug); got != tt.debugEnable {
				t.Errorf("Debug enabled = %v, want %v", got, tt.debugEnable)
			}

			if got := logger.logger.Enabled(context.TODO(), slog.LevelInfo); got != tt.infoEnable {
				t.Errorf("Info enabled = %v, want %v", got, tt.infoEnable)
			}

			if got := logger.logger.Enabled(context.TODO(), slog.LevelWarn); got != tt.warnEnable {
				t.Errorf("Warn enabled = %v, want %v", got, tt.warnEnable)
			}

			if got := logger.logger.Enabled(context.TODO(), slog.LevelError); got != tt.errorEnable {
				t.Errorf("Error enabled = %v, want %v", got, tt.errorEnable)
			}
		})
	}
}

func TestLogger_Levels(t *testing.T) {
	tests := []struct {
		name        string
		configLevel string
		log         func(*Logger)
		want        string
	}{
		{
			name:        levelDebug,
			configLevel: levelDebug,
			log: func(logger *Logger) {
				logger.Debug("debug message")
			},
			want: "debug message",
		},
		{
			name:        levelInfo,
			configLevel: levelInfo,
			log: func(logger *Logger) {
				logger.Info("info message")
			},
			want: "info message",
		},
		{
			name:        levelWarn,
			configLevel: levelWarn,
			log: func(logger *Logger) {
				logger.Warn("warn message")
			},
			want: "warn message",
		},
		{
			name:        levelError,
			configLevel: levelError,
			log: func(logger *Logger) {
				logger.Error("error message")
			},
			want: "error message",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer

			handler := slog.NewTextHandler(&buf, &slog.HandlerOptions{
				Level: slog.LevelDebug,
			})

			logger := &Logger{
				logger: slog.New(handler),
			}

			tt.log(logger)

			output := buf.String()

			if !strings.Contains(output, tt.want) {
				t.Fatalf(
					"expected log to contain %q, got %q",
					tt.want,
					output,
				)
			}
		})
	}
}

func TestLogger_InfoContext(t *testing.T) {
	var buf bytes.Buffer

	handler := slog.NewTextHandler(&buf, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})

	logger := &Logger{
		logger: slog.New(handler),
	}

	logger.InfoContext(
		"HTTP request",
		slog.String("method", "GET"),
		slog.String("path", "/hello"),
		slog.Int("status", 200),
	)

	output := buf.String()

	for _, expected := range []string{
		"HTTP request",
		"method=GET",
		"path=/hello",
		"status=200",
	} {
		if !strings.Contains(output, expected) {
			t.Errorf(
				"expected log to contain %q, got %q",
				expected,
				output,
			)
		}
	}
}
