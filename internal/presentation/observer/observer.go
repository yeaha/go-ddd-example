package observer

import (
	"context"
	"sync"

	userApp "ddd-example/internal/app"
	"ddd-example/internal/app/event"
	"ddd-example/internal/option"
)

// Start 启动领域事件观察者
func Start(ctx context.Context, opt *option.Options) {
	(&emailNotifier{
		App: userApp.NewApplication(opt),
	}).Subscribe(ctx, event.Stream)
}

// Stop 关闭领域事件流
func Stop(wg *sync.WaitGroup) {
	defer wg.Done()

	event.CloseStream()
}
