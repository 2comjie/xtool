package eventx

// IEvent 是所有事件必须实现的接口
type IEvent interface {
	Key() string
}

// RedisChannelMessage 是 Redis Pub/Sub 传输的消息包装
type RedisChannelMessage struct {
	Topic string `json:"topic"`
	Data  string `json:"data"`
}

// Processor 是事件处理回调函数类型
type Processor func(ev IEvent)
