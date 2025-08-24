package observer

import (
	"context"

	"ddd-example/internal/app/event"
	"ddd-example/internal/option"
)

// Start 启动领域事件观察者
func Start(ctx context.Context, opt *option.Options) {
	(&emailNotifier{}).Subscribe(ctx, event.Stream)
}

// Stop 关闭领域事件流
func Stop() {
	event.CloseStream()
}
