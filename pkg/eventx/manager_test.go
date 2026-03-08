package eventx

import (
	"testing"
)

func TestManagerRegisterNewChannel(t *testing.T) {
	m := NewEventManager()

	// 第一次注册应返回 true（新 channel）
	if !m.Register("listener1", "channel1") {
		t.Error("expected new channel on first register")
	}

	// 同一 channel 不同 listener，应返回 false
	if m.Register("listener2", "channel1") {
		t.Error("expected existing channel on second register")
	}

	// 新 channel 应返回 true
	if !m.Register("listener1", "channel2") {
		t.Error("expected new channel for channel2")
	}
}

func TestManagerUnregister(t *testing.T) {
	m := NewEventManager()
	m.Register("listener1", "channel1")
	m.Register("listener2", "channel1")

	// listener1 取消，channel1 仍有 listener2，应返回 false
	if m.Unregister("listener1", "channel1") {
		t.Error("expected channel still has listeners")
	}

	// listener2 取消，channel1 无人关注，应返回 true
	if !m.Unregister("listener2", "channel1") {
		t.Error("expected channel has no listeners")
	}
}

func TestManagerRemoveListener(t *testing.T) {
	m := NewEventManager()
	m.Register("listener1", "channel1")
	m.Register("listener1", "channel2")
	m.Register("listener2", "channel1")

	// 移除 listener1，channel2 应被清理（无人关注），channel1 仍有 listener2
	cleared := m.RemoveListener("listener1")
	if len(cleared) != 1 || cleared[0] != "channel2" {
		t.Errorf("expected [channel2] cleared, got %v", cleared)
	}

	// 再次移除不存在的 listener 应返回 nil
	cleared = m.RemoveListener("listener1")
	if cleared != nil {
		t.Errorf("expected nil, got %v", cleared)
	}
}
