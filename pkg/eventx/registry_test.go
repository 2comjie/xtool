package eventx

import (
	"testing"
)

type TestEvent struct {
	UserID string `json:"user_id"`
	Action string `json:"action"`
}

func (e *TestEvent) Key() string { return e.UserID }

func TestRegisterAndParse(t *testing.T) {
	Register[*TestEvent]("test_event")

	// 模拟序列化后的 payload
	payload := `{"topic":"test_event","data":"{\"user_id\":\"123\",\"action\":\"login\"}"}`

	ev, err := Parse(payload)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	te, ok := ev.(*TestEvent)
	if !ok {
		t.Fatalf("expected *TestEvent, got %T", ev)
	}
	if te.UserID != "123" {
		t.Errorf("expected UserID=123, got %s", te.UserID)
	}
	if te.Action != "login" {
		t.Errorf("expected Action=login, got %s", te.Action)
	}
}

func TestParseUnregisteredTopic(t *testing.T) {
	payload := `{"topic":"unknown_topic","data":"{}"}`
	_, err := Parse(payload)
	if err != ErrNotRegistered {
		t.Errorf("expected ErrNotRegistered, got %v", err)
	}
}

func TestParseInvalidPayload(t *testing.T) {
	_, err := Parse("not json")
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}
