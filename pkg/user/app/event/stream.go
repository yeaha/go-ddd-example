package event

import (
	"context"
	"ddd-example/pkg/utils/events"
	"reflect"

	"github.com/reactivex/rxgo/v2"
	"golang.org/x/exp/slog"
)

var (
	// stream 事件流
	stream = events.NewStream()
	// Stream 事件观察对象
	Stream = stream.Observable(rxgo.WithErrorStrategy(rxgo.ContinueOnError))
)

// publish 发布领域事件
func publish(event any) {
	if slog.Default().Enabled(context.Background(), slog.LevelDebug) {
		slog.Debug("deliver domain event",
			"type", reflect.TypeOf(event).Name(),
			"data", event,
		)
	}

	if err := stream.Publish(event); err != nil {
		slog.Error("deliver domain event",
			"type", reflect.TypeOf(event).Name(),
			"data", event,
			"error", err,
		)
	}
}

// Publish 批量发布领域事件
func Publish(events ...any) {
	for _, event := range events {
		publish(event)
	}
}

// CloseStream 关闭事件流
func CloseStream() {
	stream.Close()
}
