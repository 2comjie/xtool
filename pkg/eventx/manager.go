package eventx

import (
	"sync"
	"time"
)

// EventManager 管理 listener 与 event channel 的双向映射
type EventManager struct {
	mu               sync.RWMutex
	listenerToEvents map[string]map[string]time.Time // listener -> {channel -> expireTime}
	eventToListeners map[string]map[string]time.Time // channel  -> {listener -> expireTime}
	ttl              time.Duration
	stopCh           chan struct{}
}

type ManagerOption func(*EventManager)

// WithTTL 设置注册项的过期时间，并启动后台清理协程
func WithTTL(ttl time.Duration, checkInterval time.Duration) ManagerOption {
	return func(m *EventManager) {
		m.ttl = ttl
		go m.cleanup(checkInterval)
	}
}

func NewEventManager(opts ...ManagerOption) *EventManager {
	m := &EventManager{
		listenerToEvents: make(map[string]map[string]time.Time),
		eventToListeners: make(map[string]map[string]time.Time),
		stopCh:           make(chan struct{}),
	}
	for _, opt := range opts {
		opt(m)
	}
	return m
}

// Register 注册 listener 对 channel 的关注，返回 channel 是否为新增
func (m *EventManager) Register(listener, channel string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	var expireTime time.Time
	if m.ttl > 0 {
		expireTime = time.Now().Add(m.ttl)
	}

	if _, ok := m.listenerToEvents[listener]; !ok {
		m.listenerToEvents[listener] = make(map[string]time.Time)
	}
	m.listenerToEvents[listener][channel] = expireTime

	newChannel := false
	if _, ok := m.eventToListeners[channel]; !ok {
		newChannel = true
		m.eventToListeners[channel] = make(map[string]time.Time)
	}
	m.eventToListeners[channel][listener] = expireTime
	return newChannel
}

// Unregister 取消 listener 对 channel 的关注，返回 channel 是否已无人关注
func (m *EventManager) Unregister(listener, channel string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	if eventMap, ok := m.listenerToEvents[listener]; ok {
		delete(eventMap, channel)
		if len(eventMap) == 0 {
			delete(m.listenerToEvents, listener)
		}
	}
	if listenerMap, ok := m.eventToListeners[channel]; ok {
		delete(listenerMap, listener)
		if len(listenerMap) == 0 {
			delete(m.eventToListeners, channel)
			return true
		}
	}
	return false
}

// RemoveListener 移除整个 listener，返回已无人关注的 channel 列表
func (m *EventManager) RemoveListener(listener string) []string {
	m.mu.Lock()
	defer m.mu.Unlock()

	eventMap, ok := m.listenerToEvents[listener]
	if !ok {
		return nil
	}

	var clearChannels []string
	for channel := range eventMap {
		if listenerMap, ok := m.eventToListeners[channel]; ok {
			delete(listenerMap, listener)
			if len(listenerMap) == 0 {
				clearChannels = append(clearChannels, channel)
				delete(m.eventToListeners, channel)
			}
		}
	}
	delete(m.listenerToEvents, listener)
	return clearChannels
}

// Stop 停止后台清理协程
func (m *EventManager) Stop() {
	select {
	case <-m.stopCh:
	default:
		close(m.stopCh)
	}
}

func (m *EventManager) cleanup(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-m.stopCh:
			return
		case now := <-ticker.C:
			m.cleanOnce(now)
		}
	}
}

func (m *EventManager) cleanOnce(now time.Time) {
	m.mu.Lock()
	defer m.mu.Unlock()

	for listener, eventMap := range m.listenerToEvents {
		for channel, expireTime := range eventMap {
			if !expireTime.IsZero() && expireTime.Before(now) {
				delete(eventMap, channel)
			}
		}
		if len(eventMap) == 0 {
			delete(m.listenerToEvents, listener)
		}
	}

	for channel, listenerMap := range m.eventToListeners {
		for listener, expireTime := range listenerMap {
			if !expireTime.IsZero() && expireTime.Before(now) {
				delete(listenerMap, listener)
			}
		}
		if len(listenerMap) == 0 {
			delete(m.eventToListeners, channel)
		}
	}
}
