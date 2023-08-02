package events

import (
	"context"
	"errors"
	"sync/atomic"
	"time"

	"github.com/reactivex/rxgo/v2"
)

// Stream 领域事件流
type Stream struct {
	events chan rxgo.Item
	close  uint32
}

// NewStream 构造函数
func NewStream() *Stream {
	return &Stream{
		events: make(chan rxgo.Item, 1),
	}
}

// Observable 事件流订阅对象
func (s *Stream) Observable(opts ...rxgo.Option) rxgo.Observable {
	if atomic.LoadUint32(&s.close) != 0 {
		return rxgo.Just(rxgo.Error(errors.New("stream closed")))()
	}

	return rxgo.FromEventSource(s.events, opts...)
}

// Publish 发布领域事件
func (s *Stream) Publish(item any) error {
	if atomic.LoadUint32(&s.close) != 0 {
		return errors.New("stream closed")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if !rxgo.Of(item).SendContext(ctx, s.events) {
		return errors.New("send item failed")
	}
	return nil
}

// Close 关闭
func (s *Stream) Close() {
	if atomic.CompareAndSwapUint32(&s.close, 0, 1) {
		close(s.events)
	}
}
