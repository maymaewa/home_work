package internalhttp

import (
	"log/slog"
	"net"
	"net/http"
	"time"
)

func loggingMiddleware(next http.Handler, logger Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		clientIP, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			clientIP = r.RemoteAddr
		}

		rw := &responseWriter{
			ResponseWriter: w,
		}

		next.ServeHTTP(rw, r)

		statusCode := rw.statusCode
		if statusCode == 0 {
			statusCode = http.StatusOK
		}

		logger.InfoContext(
			"HTTP request",
			slog.String("client_ip", clientIP),
			slog.String("time", start.Format(time.RFC3339)),
			slog.String("method", r.Method),
			slog.String("path", r.URL.Path),
			slog.String("http_version", r.Proto),
			slog.Int("status", statusCode),
			slog.Duration("latency", time.Since(start)),
			slog.String("user_agent", r.UserAgent()),
		)
	})
}

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (w *responseWriter) WriteHeader(statusCode int) {
	if w.statusCode != 0 {
		return
	}

	w.statusCode = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

func (w *responseWriter) Write(body []byte) (int, error) {
	if w.statusCode == 0 {
		w.WriteHeader(http.StatusOK)
	}

	return w.ResponseWriter.Write(body)
}
