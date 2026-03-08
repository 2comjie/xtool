package eventx

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
)

func setupRedis(t *testing.T) redis.UniversalClient {
	rc := redis.NewClient(&redis.Options{
		Addr: "127.0.0.1:6379",
	})
	if err := rc.Ping(context.Background()).Err(); err != nil {
		t.Skipf("跳过集成测试，Redis 不可用: %v", err)
	}
	return rc
}

type UserLoginEvent struct {
	UserID  string `json:"user_id"`
	LoginAt int64  `json:"login_at"`
}

func (e *UserLoginEvent) Key() string { return e.UserID }

func TestIntegrationPubSub(t *testing.T) {
	rc := setupRedis(t)
	defer rc.Close()

	Register[*UserLoginEvent]("user_login")

	channel := fmt.Sprintf("test:eventx:%d", time.Now().UnixNano())

	var wg sync.WaitGroup
	wg.Add(1)

	var received *UserLoginEvent

	listener := NewListener(context.Background(), rc, func(ev IEvent) {
		if e, ok := ev.(*UserLoginEvent); ok {
			received = e
			wg.Done()
		}
	})
	defer listener.Close()

	// 订阅
	err := listener.Subscribe("test-listener", channel)
	if err != nil {
		t.Fatalf("Subscribe failed: %v", err)
	}

	// 等待订阅生效
	time.Sleep(time.Millisecond * 200)

	// 发布
	now := time.Now().Unix()
	n, err := Publish(rc, channel, &UserLoginEvent{
		UserID:  "user_456",
		LoginAt: now,
	})
	if err != nil {
		t.Fatalf("Publish failed: %v", err)
	}
	t.Logf("Publish 成功，%d 个订阅者收到消息", n)

	// 等待接收（最多 3 秒）
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		t.Logf("收到事件: UserID=%s, LoginAt=%d", received.UserID, received.LoginAt)
		if received.UserID != "user_456" {
			t.Errorf("expected UserID=user_456, got %s", received.UserID)
		}
		if received.LoginAt != now {
			t.Errorf("expected LoginAt=%d, got %d", now, received.LoginAt)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("超时：3秒内未收到事件")
	}
}

func TestIntegrationUnsubscribe(t *testing.T) {
	rc := setupRedis(t)
	defer rc.Close()

	Register[*UserLoginEvent]("user_login")

	channel := fmt.Sprintf("test:eventx:unsub:%d", time.Now().UnixNano())

	receivedCount := 0
	var mu sync.Mutex

	listener := NewListener(context.Background(), rc, func(ev IEvent) {
		mu.Lock()
		receivedCount++
		mu.Unlock()
	})
	defer listener.Close()

	// 订阅
	listener.Subscribe("test-listener", channel)
	time.Sleep(time.Millisecond * 200)

	// 发布第一条
	Publish(rc, channel, &UserLoginEvent{UserID: "1"})
	time.Sleep(time.Millisecond * 500)

	mu.Lock()
	count1 := receivedCount
	mu.Unlock()
	t.Logf("取消订阅前收到 %d 条消息", count1)

	// 取消订阅
	listener.RemoveListener("test-listener")
	time.Sleep(time.Millisecond * 200)

	// 发布第二条（不应收到）
	Publish(rc, channel, &UserLoginEvent{UserID: "2"})
	time.Sleep(time.Millisecond * 500)

	mu.Lock()
	count2 := receivedCount
	mu.Unlock()
	t.Logf("取消订阅后收到 %d 条消息", count2)

	if count1 != 1 {
		t.Errorf("expected 1 message before unsubscribe, got %d", count1)
	}
	if count2 != 1 {
		t.Errorf("expected still 1 message after unsubscribe, got %d", count2)
	}
}
