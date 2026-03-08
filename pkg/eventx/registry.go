package eventx

import (
	"context"
	"encoding/json"
	"reflect"

	"github.com/2comjie/xtool/pkg/logx"
	"github.com/2comjie/xtool/pkg/syncx"
	"github.com/redis/go-redis/v9"
)

var (
	topicRegistry = syncx.NewSafeMap[string, *registryInfo]()
	typeRegistry  = syncx.NewSafeMap[reflect.Type, *registryInfo]()
)

type registryInfo struct {
	Topic    string
	DataType reflect.Type
}

// Register 注册事件类型到全局注册表
// T 必须实现 IEvent 接口，topic 是消息路由标识
func Register[T IEvent](topic string) {
	t := reflect.TypeFor[T]()
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	info := &registryInfo{Topic: topic, DataType: t}
	topicRegistry.Store(topic, info)
	typeRegistry.Store(t, info)
}

// Publish 将事件发布到指定 Redis channel
func Publish(rc redis.UniversalClient, channel string, data IEvent) (int64, error) {
	t := reflect.TypeOf(data)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	info, ok := typeRegistry.Load(t)
	if !ok {
		logx.Errorf("eventx: publish failed, type not registered: channel=%s data=%+v", channel, data)
		return 0, ErrNotRegistered
	}
	dataBs, err := json.Marshal(data)
	if err != nil {
		return 0, err
	}
	msg := RedisChannelMessage{Topic: info.Topic, Data: string(dataBs)}
	bs, err := json.Marshal(msg)
	if err != nil {
		return 0, err
	}
	return rc.Publish(context.Background(), channel, string(bs)).Result()
}

// Parse 将 Redis 消息 payload 解析为具体的 IEvent 实例
func Parse(payload string) (IEvent, error) {
	var msg RedisChannelMessage
	if err := json.Unmarshal([]byte(payload), &msg); err != nil {
		return nil, err
	}
	info, ok := topicRegistry.Load(msg.Topic)
	if !ok {
		return nil, ErrNotRegistered
	}
	ret := reflect.New(info.DataType).Interface()
	if err := json.Unmarshal([]byte(msg.Data), ret); err != nil {
		return nil, err
	}
	return ret.(IEvent), nil
}
