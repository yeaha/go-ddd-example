package events

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/reactivex/rxgo/v2"
)

// EventStream 领域事件流
type EventStream struct {
	events chan rxgo.Item
	m      sync.RWMutex
	closed bool
}

// NewEventStream 构造函数
func NewEventStream() *EventStream {
	return &EventStream{
		events: make(chan rxgo.Item, 1),
	}
}

// Observable 事件流订阅对象
func (es *EventStream) Observable(opts ...rxgo.Option) rxgo.Observable {
	es.m.RLock()
	closed := es.closed
	es.m.RUnlock()

	if closed {
		return rxgo.Just(rxgo.Error(errors.New("stream closed")))()
	}
	return rxgo.FromEventSource(es.events, opts...)
}

// Publish 发布领域事件
func (es *EventStream) Publish(item any) error {
	es.m.RLock()
	closed := es.closed
	es.m.RUnlock()

	if closed {
		return errors.New("event stream closed")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if !rxgo.Of(item).SendContext(ctx, es.events) {
		return errors.New("send item failed")
	}
	return nil
}

// Close 关闭
func (es *EventStream) Close() {
	es.m.Lock()
	defer es.m.Unlock()

	if !es.closed {
		es.closed = true
		close(es.events)
	}
}
