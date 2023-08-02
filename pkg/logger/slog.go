package logger

import (
	"context"
	"os"

	"golang.org/x/exp/slog"
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

// LogAndExist 日志并退出
func LogAndExist(ctx context.Context, msg string, args ...any) {
	logAndExist(ctx, slog.LevelInfo, msg, args...)
}

// ErrorAndExist 记录错误并退出
func ErrorAndExist(ctx context.Context, msg string, args ...any) {
	logAndExist(ctx, slog.LevelError, msg, args...)
}

func logAndExist(ctx context.Context, level slog.Level, msg string, args ...any) {
	FromContext(ctx).Log(ctx, level, msg, args...)
	os.Exit(1) // revive:disable-line
}
