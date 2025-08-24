package logger

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"log/syslog"
	"os"
)

type loggerKey struct{}

// Option is the logger option.
type Option struct {
	Level  string // debug, info(default), warn, error
	Output string // stdout, stderr(default), syslog
	Format string // text(default), json
}

// New creates a new logger.
func New(opt Option) (*slog.Logger, error) {
	var lvl slog.Level
	switch opt.Level {
	case "debug":
		lvl = slog.LevelDebug
	case "warn":
		lvl = slog.LevelWarn
	case "error":
		lvl = slog.LevelError
	default:
		lvl = slog.LevelInfo
	}

	var (
		output io.Writer
		err    error
	)
	switch opt.Output {
	case "syslog":
		output, err = syslog.New(syslog.LOG_DEBUG, "lightshow")
		if err != nil {
			return nil, fmt.Errorf("create syslog writer, %w", err)
		}
	case "stdout":
		output = os.Stdout
	default:
		output = os.Stderr
	}

	var handler slog.Handler
	switch opt.Format {
	case "json":
		handler = slog.NewJSONHandler(output, &slog.HandlerOptions{Level: lvl})
	default:
		handler = slog.NewTextHandler(output, &slog.HandlerOptions{Level: lvl})
	}

	return slog.New(handler), nil
}

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
