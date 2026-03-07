package logx

import (
	"bytes"
	"strings"
	"sync"
	"testing"
)

func newTestLogger(name string) (*defaultLogger, *bytes.Buffer) {
	buf := &bytes.Buffer{}
	l := &defaultLogger{
		name:   name,
		level:  TraceLevel,
		hooks:  make(LevelHooks),
		fields: make(map[string]any),
		output: buf,
	}
	return l, buf
}

func TestLogLevels(t *testing.T) {
	l, buf := newTestLogger("test")

	l.Trace("trace msg")
	l.Debug("debug msg")
	l.Info("info msg")
	l.Warn("warn msg")
	l.Error("error msg")

	output := buf.String()
	for _, expected := range []string{"TRACE", "DEBUG", "INFO", "WARN", "ERROR"} {
		if !strings.Contains(output, expected) {
			t.Errorf("expected output to contain %q, got:\n%s", expected, output)
		}
	}
}

func TestLogFormat(t *testing.T) {
	l, buf := newTestLogger("myApp")

	l.Infof("hello %s, port %d", "world", 8080)

	output := buf.String()
	if !strings.Contains(output, "hello world, port 8080") {
		t.Errorf("expected formatted message, got: %s", output)
	}
	if !strings.Contains(output, "[INFO ]") {
		t.Errorf("expected level [INFO ], got: %s", output)
	}
	if !strings.Contains(output, "[myApp]") {
		t.Errorf("expected logger name [myApp], got: %s", output)
	}
}

func TestCallerInfo(t *testing.T) {
	l, buf := newTestLogger("caller-test")

	l.Info("check caller")

	output := buf.String()
	t.Logf("caller output: %s", output)
	// 应该包含相对路径 logx_test.go，而不是绝对路径
	if !strings.Contains(output, "logx_test.go:") {
		t.Errorf("expected caller to contain logx_test.go:line, got: %s", output)
	}
	// 不应包含绝对路径前缀
	if strings.Contains(output, "/Users/") {
		t.Errorf("expected relative path, but got absolute path: %s", output)
	}
}

func TestCallerInfoPackageLevel(t *testing.T) {
	// 测试通过包级别函数调用时，caller 指向测试文件而不是 logx.go
	buf := &bytes.Buffer{}
	old := global
	global = &defaultLogger{
		name:   "pkg-test",
		level:  TraceLevel,
		hooks:  make(LevelHooks),
		fields: make(map[string]any),
		output: buf,
	}
	defer func() { global = old }()

	Info("package level call")

	output := buf.String()
	t.Logf("package-level caller output: %s", output)
	// 应该指向 logx_test.go，而不是 logx.go
	if !strings.Contains(output, "logx_test.go:") {
		t.Errorf("expected caller to point to logx_test.go, got: %s", output)
	}
	if strings.Contains(output, "logx.go:") {
		t.Errorf("caller should NOT point to logx.go wrapper, got: %s", output)
	}
}

func TestLevelFilter(t *testing.T) {
	l, buf := newTestLogger("filter-test")
	l.level = WarnLevel

	l.Debug("should not appear")
	l.Info("should not appear")
	l.Warn("should appear")
	l.Error("should appear")

	output := buf.String()
	if strings.Contains(output, "should not appear") {
		t.Errorf("debug/info should be filtered, got: %s", output)
	}
	if !strings.Contains(output, "WARN") || !strings.Contains(output, "ERROR") {
		t.Errorf("warn/error should appear, got: %s", output)
	}
}

func TestWithFields(t *testing.T) {
	l, buf := newTestLogger("fields-test")

	l2 := l.WithField("requestId", "abc-123").(*defaultLogger)
	l2.output = buf
	l2.Info("with field")

	output := buf.String()
	if !strings.Contains(output, "requestId") || !strings.Contains(output, "abc-123") {
		t.Errorf("expected fields in output, got: %s", output)
	}
}

func TestWithFieldsMultiple(t *testing.T) {
	l, buf := newTestLogger("fields-test")

	l2 := l.WithFields(map[string]any{"uid": 42, "action": "login"}).(*defaultLogger)
	l2.output = buf
	l2.Info("multi fields")

	output := buf.String()
	if !strings.Contains(output, "uid") || !strings.Contains(output, "action") {
		t.Errorf("expected multiple fields in output, got: %s", output)
	}
}

// testHook 用于测试的钩子实现
type testHook struct {
	mu      sync.Mutex
	entries []*Entry
	levels  []Level
}

func (h *testHook) Levels() []Level {
	return h.levels
}

func (h *testHook) Fire(entry *Entry) error {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.entries = append(h.entries, entry)
	return nil
}

func TestHook(t *testing.T) {
	l, _ := newTestLogger("hook-test")

	hook := &testHook{levels: []Level{ErrorLevel, WarnLevel}}
	l.AddHook(hook)

	l.Info("info - no hook")
	l.Warn("warn - hook fires")
	l.Error("error - hook fires")
	l.Debug("debug - no hook")

	hook.mu.Lock()
	defer hook.mu.Unlock()
	if len(hook.entries) != 2 {
		t.Errorf("expected 2 hook fires, got %d", len(hook.entries))
	}
	if hook.entries[0].Level != WarnLevel {
		t.Errorf("expected first hook entry to be WARN, got %s", hook.entries[0].Level)
	}
	if hook.entries[1].Level != ErrorLevel {
		t.Errorf("expected second hook entry to be ERROR, got %s", hook.entries[1].Level)
	}
}

func TestGetLogger(t *testing.T) {
	l1 := GetLogger("service-a")
	l2 := GetLogger("service-a")
	l3 := GetLogger("service-b")

	if l1 != l2 {
		t.Error("same name should return same logger instance")
	}
	if l1 == l3 {
		t.Error("different name should return different logger instance")
	}
}

func TestParseLevel(t *testing.T) {
	tests := []struct {
		input    string
		expected Level
	}{
		{"TRACE", TraceLevel},
		{"debug", DebugLevel},
		{"Info", InfoLevel},
		{" WARN ", WarnLevel},
		{"ERROR", ErrorLevel},
		{"fatal", FatalLevel},
		{"off", Off},
		{"unknown", InfoLevel},
	}
	for _, tt := range tests {
		got := ParseLevel(tt.input)
		if got != tt.expected {
			t.Errorf("ParseLevel(%q) = %v, want %v", tt.input, got, tt.expected)
		}
	}
}
