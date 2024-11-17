package middleware

import (
	"context"
	"go.uber.org/zap"
	"net/http"
)

type contextKey string

const loggerKey contextKey = "logger"

// Logging добавляет middleware передающий logger приложения в контекст запроса.
func Logging(logger *zap.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := context.WithValue(r.Context(), loggerKey, logger)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GetLogger получает logger переданный в контекст запроса.
func GetLogger(ctx context.Context) *zap.Logger {
	logger, ok := ctx.Value(loggerKey).(*zap.Logger)
	if !ok {
		panic("Не могу получить логгер из контекста")
		return nil
	}
	return logger
}
