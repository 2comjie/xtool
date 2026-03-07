package logx

import (
	"os"
	"testing"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

func newObservedLogger() (*zap.SugaredLogger, *observer.ObservedLogs) {
	core, logs := observer.New(zapcore.DebugLevel)
	logger := zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1))
	return logger.Sugar(), logs
}

func TestInfof(t *testing.T) {
	sugar, logs := newObservedLogger()
	old := global
	global = sugar
	defer func() { global = old }()

	Infof("hello %s", "world")

	entries := logs.All()
	if len(entries) != 1 {
		t.Fatalf("expected 1 log entry, got %d", len(entries))
	}
	e := entries[0]
	if e.Level != zapcore.InfoLevel {
		t.Errorf("expected INFO, got %s", e.Level)
	}
	if e.Message != "hello world" {
		t.Errorf("expected 'hello world', got %q", e.Message)
	}
	// caller 应该指向本测试文件
	if e.Caller.File == "" {
		t.Error("expected caller info, got empty")
	}
	t.Logf("caller: %s", e.Caller)
}

func TestLevels(t *testing.T) {
	sugar, logs := newObservedLogger()
	old := global
	global = sugar
	defer func() { global = old }()

	Debug("d")
	Debugf("d%s", "f")
	Info("i")
	Infof("i%s", "f")
	Warn("w")
	Warnf("w%s", "f")
	Error("e")
	Errorf("e%s", "f")

	entries := logs.All()
	if len(entries) != 8 {
		t.Fatalf("expected 8 log entries, got %d", len(entries))
	}
	expected := []zapcore.Level{
		zapcore.DebugLevel, zapcore.DebugLevel,
		zapcore.InfoLevel, zapcore.InfoLevel,
		zapcore.WarnLevel, zapcore.WarnLevel,
		zapcore.ErrorLevel, zapcore.ErrorLevel,
	}
	for i, e := range entries {
		if e.Level != expected[i] {
			t.Errorf("entry[%d]: expected %s, got %s", i, expected[i], e.Level)
		}
	}
}

func TestWith(t *testing.T) {
	sugar, logs := newObservedLogger()
	old := global
	global = sugar
	defer func() { global = old }()

	l := With("requestId", "abc-123")
	l.Info("with fields")

	entries := logs.All()
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	ctx := entries[0].ContextMap()
	if ctx["requestId"] != "abc-123" {
		t.Errorf("expected requestId=abc-123, got %v", ctx["requestId"])
	}
}

func TestCallerPointsToCallSite(t *testing.T) {
	sugar, logs := newObservedLogger()
	old := global
	global = sugar
	defer func() { global = old }()

	Info("caller test") // 这一行就是 caller 应该指向的位置

	entries := logs.All()
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	caller := entries[0].Caller
	t.Logf("caller: %s", caller)
	if caller.File == "" {
		t.Fatal("caller file is empty")
	}
	// 不应该指向 logx.go
	if caller.File == "logx.go" {
		t.Error("caller should not point to logx.go")
	}
}

func TestInit(t *testing.T) {
	tmpFile := t.TempDir() + "/test.log"
	Init("debug", tmpFile)
	defer Sync()

	Infof("file output test")

	data, err := os.ReadFile(tmpFile)
	if err != nil {
		t.Fatalf("failed to read log file: %v", err)
	}
	if len(data) == 0 {
		t.Error("log file should not be empty")
	}
	t.Logf("file content: %s", string(data))
}
