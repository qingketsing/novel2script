package observability

import (
	"context"
	"fmt"
	"log/slog"
	"sync/atomic"
	"time"
)

type contextKey string

const (
	loggerContextKey    contextKey = "logger"
	requestIDContextKey contextKey = "request_id"
)

var requestCounter atomic.Uint64

func NewRequestID() string {
	return fmt.Sprintf("req_%d_%06d", time.Now().UnixMilli(), requestCounter.Add(1))
}

func WithLogger(ctx context.Context, logger *slog.Logger) context.Context {
	if logger == nil {
		logger = slog.Default()
	}
	return context.WithValue(ctx, loggerContextKey, logger)
}

func Logger(ctx context.Context) *slog.Logger {
	logger, ok := ctx.Value(loggerContextKey).(*slog.Logger)
	if !ok || logger == nil {
		return slog.Default()
	}
	return logger
}

func WithRequestID(ctx context.Context, requestID string) context.Context {
	if requestID == "" {
		requestID = NewRequestID()
	}
	return context.WithValue(ctx, requestIDContextKey, requestID)
}

func RequestID(ctx context.Context) string {
	requestID, _ := ctx.Value(requestIDContextKey).(string)
	return requestID
}
