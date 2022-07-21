package domain

import (
	"reflect"

	"github.com/reactivex/rxgo/v2"
	"github.com/sirupsen/logrus"
	"gitlab.haochang.tv/yangyi/examine-code/pkg/utils/events"
)

var (
	// Stream 事件流
	Stream = events.NewEventStream()
	// Events 事件观察对象
	Events = Stream.Observable(rxgo.WithErrorStrategy(rxgo.ContinueOnError))
)

// PublishEvent 发布领域事件
func PublishEvent(event any) {
	if logrus.StandardLogger().Level == logrus.TraceLevel {
		logrus.WithFields(logrus.Fields{
			"type": reflect.TypeOf(event).Name(),
			"data": event,
		}).Trace("deliver domain event")
	}

	if err := Stream.Publish(event); err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"type": reflect.TypeOf(event).Name(),
			"data": event,
		}).Error("deliver domain event")
	}
}

// PublishEvents 批量发布领域事件
func PublishEvents(events ...any) {
	for _, event := range events {
		PublishEvent(event)
	}
}

// EventLogin 账号登录
type EventLogin struct {
	User *User
}

// EventRegister 账号注册
type EventRegister struct {
	User *User
}
