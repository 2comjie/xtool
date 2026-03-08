package eventx

import (
	"context"

	"github.com/2comjie/xtool/pkg/logx"
	"github.com/2comjie/xtool/pkg/safex"
	"github.com/redis/go-redis/v9"
)

// Listener 封装 Redis PubSub 订阅，接收消息后通过 Processor 分发
type Listener struct {
	pubsub    *redis.PubSub
	mgr       *EventManager
	processor Processor
	ctx       context.Context
	cancel    context.CancelFunc
}

// NewListener 创建并启动一个监听器
func NewListener(ctx context.Context, rc redis.UniversalClient, processor Processor, opts ...ManagerOption) *Listener {
	ctx, cancel := context.WithCancel(ctx)
	pubsub := rc.Subscribe(ctx)
	l := &Listener{
		pubsub:    pubsub,
		mgr:       NewEventManager(opts...),
		processor: processor,
		ctx:       ctx,
		cancel:    cancel,
	}
	go l.run()
	return l
}

func (l *Listener) run() {
	defer l.cancel()
	ch := l.pubsub.Channel()
	for {
		select {
		case <-l.ctx.Done():
			return
		case msg, ok := <-ch:
			if !ok {
				logx.Error("eventx: pubsub channel closed")
				return
			}
			if msg == nil {
				continue
			}
			ev, err := Parse(msg.Payload)
			if err != nil {
				logx.Errorf("eventx: parse message error: %v", err)
				continue
			}
			safex.Run(func() {
				l.processor(ev)
			})
		}
	}
}

// Subscribe 为指定 listener 订阅 channels
func (l *Listener) Subscribe(listener string, channels ...string) error {
	var toAdd []string
	for _, c := range channels {
		if l.mgr.Register(listener, c) {
			toAdd = append(toAdd, c)
		}
	}
	if len(toAdd) > 0 {
		return l.pubsub.Subscribe(l.ctx, toAdd...)
	}
	return nil
}

// Unsubscribe 为指定 listener 取消订阅 channels
func (l *Listener) Unsubscribe(listener string, channels ...string) error {
	var toRemove []string
	for _, c := range channels {
		if l.mgr.Unregister(listener, c) {
			toRemove = append(toRemove, c)
		}
	}
	if len(toRemove) > 0 {
		return l.pubsub.Unsubscribe(l.ctx, toRemove...)
	}
	return nil
}

// RemoveListener 移除整个 listener 的所有订阅
func (l *Listener) RemoveListener(listener string) error {
	clearChannels := l.mgr.RemoveListener(listener)
	if len(clearChannels) > 0 {
		return l.pubsub.Unsubscribe(l.ctx, clearChannels...)
	}
	return nil
}

// Close 关闭监听器
func (l *Listener) Close() error {
	l.cancel()
	l.mgr.Stop()
	return l.pubsub.Close()
}
