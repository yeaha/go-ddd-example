package observer

import (
	"context"

	"ddd-example/pkg/user/app"
	"ddd-example/pkg/user/app/event"

	"github.com/reactivex/rxgo/v2"
	"github.com/sirupsen/logrus"
)

// 在用户注册成功之后发送邮件
type emailNotifier struct {
	App *app.Application
}

func (o *emailNotifier) Subscribe(ctx context.Context, events rxgo.Observable) rxgo.Disposed {
	logger := logrus.WithField("scope", "observer.emailNotifier")
	logger.Info("start")

	return events.
		Filter(func(item any) bool {
			_, ok := item.(event.Register)
			return ok
		}).
		ForEach(
			func(item any) {
				ev := item.(event.Register)

				// 这里就不具体实现邮件发送，打条日志意思一下
				logger.WithField("email", ev.User.Email).Info("send email to new user")
			},
			func(err error) {
				logger.WithError(err).Error("handle event")
			},
			func() {
				logger.Warn("complete")
			},

			rxgo.WithContext(ctx),
			rxgo.WithBufferedChannel(10),
		)
}
