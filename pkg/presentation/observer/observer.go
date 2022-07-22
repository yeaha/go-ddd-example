package observer

import (
	"context"
	"sync"

	"ddd-example/pkg/option"
	userApp "ddd-example/pkg/user/app"
	user "ddd-example/pkg/user/domain"
)

// Start 启动领域事件观察者
func Start(ctx context.Context, opt *option.Options) {
	(&emailNotifier{
		App: userApp.NewApplication(opt),
	}).Subscribe(ctx, user.Events)
}

// Stop 关闭领域事件流
func Stop(wg *sync.WaitGroup) {
	defer wg.Done()

	user.Stream.Close()
}
