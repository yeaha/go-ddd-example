package logger

import (
	"context"
	"log/slog"
)

type loggerKey struct{}

// NewContext 返回包含了slog.Logger的context
func NewContext(ctx context.Context, l *slog.Logger) context.Context {
	return context.WithValue(ctx, loggerKey{}, l)
}

// FromContext 从context中获取可能存在的slog.Logger，如果不存在，返回slog default logger
func FromContext(ctx context.Context) *slog.Logger {
	if l, ok := ctx.Value(loggerKey{}).(*slog.Logger); ok {
		return l
	}
	return slog.Default()
}

func log(ctx context.Context, level slog.Level, msg string, args ...any) {
	FromContext(ctx).Log(ctx, level, msg, args...)
}

// Debug debug级别日志
func Debug(ctx context.Context, msg string, args ...any) {
	log(ctx, slog.LevelDebug, msg, args...)
}

// Info info级别日志
func Info(ctx context.Context, msg string, args ...any) {
	log(ctx, slog.LevelInfo, msg, args...)
}

// Warn warn级别日志
func Warn(ctx context.Context, msg string, args ...any) {
	log(ctx, slog.LevelWarn, msg, args...)
}

// Error error级别日志
func Error(ctx context.Context, msg string, args ...any) {
	log(ctx, slog.LevelError, msg, args...)
}
