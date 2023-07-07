package event

import (
	"ddd-example/pkg/utils/events"
	"reflect"

	"github.com/reactivex/rxgo/v2"
	"github.com/sirupsen/logrus"
)

var (
	// stream 事件流
	stream = events.NewStream()
	// Stream 事件观察对象
	Stream = stream.Observable(rxgo.WithErrorStrategy(rxgo.ContinueOnError))
)

// publish 发布领域事件
func publish(event any) {
	if logrus.StandardLogger().Level == logrus.TraceLevel {
		logrus.WithFields(logrus.Fields{
			"type": reflect.TypeOf(event).Name(),
			"data": event,
		}).Trace("deliver domain event")
	}

	if err := stream.Publish(event); err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"type": reflect.TypeOf(event).Name(),
			"data": event,
		}).Error("deliver domain event")
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
