package logger

import (
	"context"
	"log/slog"
)

type contextKey string

const loggerKey = contextKey("logger")

func WithLogger(ctx context.Context, logger *slog.Logger) context.Context {
	return context.WithValue(ctx, loggerKey, logger)
}

func FromContext(ctx context.Context) *slog.Logger {
	if logger, ok := ctx.Value(loggerKey).(*slog.Logger); ok {
		return logger
	}
	return NewLogger()
}
