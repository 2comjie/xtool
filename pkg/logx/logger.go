package logx

import (
	"fmt"
	"io"
	"maps"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"
)

// logxFuncPrefix 用于识别 logx 包自身的函数帧（基于 frame.Function）。
// projectRoot 用于裁剪文件路径为项目相对路径。
var (
	logxFuncPrefix string
	projectRoot    string
)

func init() {
	// 通过当前函数的 PC 获取完整函数名，推断 logx 包的函数前缀
	pc, file, _, ok := runtime.Caller(0)
	if !ok {
		return
	}
	// funcName 类似 "github.com/2comjie/xtool/pkg/logx.init"
	funcName := runtime.FuncForPC(pc).Name()
	if idx := strings.LastIndex(funcName, "/logx."); idx >= 0 {
		logxFuncPrefix = funcName[:idx+len("/logx.")]
	}
	// projectRoot: /Users/.../project/xtool/
	const marker = "/pkg/"
	if idx := strings.Index(file, marker); idx >= 0 {
		projectRoot = file[:idx+1]
	}
}

type Logger interface {
	Name() string

	Trace(msg string)
	Tracef(format string, args ...any)
	Debug(msg string)
	Debugf(format string, args ...any)
	Info(msg string)
	Infof(format string, args ...any)
	Warn(msg string)
	Warnf(format string, args ...any)
	Error(msg string)
	Errorf(format string, args ...any)
	Fatal(msg string)
	Fatalf(format string, args ...any)

	// WithField 返回携带额外字段的新 Logger（类似 SLF4J 的 MDC）。
	WithField(key string, value any) Logger
	// WithFields 返回携带多个额外字段的新 Logger。
	WithFields(fields map[string]any) Logger

	// SetLevel 设置日志级别。
	SetLevel(level Level)
	// AddHook 添加钩子。
	AddHook(hook Hook)
	// Close 关闭 Logger（释放文件资源等）。
	Close() error
}

// LoggerOption 用于配置 Logger 的选项。
type LoggerOption func(l *defaultLogger)

// WithOutput 自定义输出目标，替换默认的 stdout。
func WithOutput(w io.Writer) LoggerOption {
	return func(l *defaultLogger) {
		l.output = w
	}
}

// WithStdout 输出到 stdout（默认行为，可与 WithFileOutput 组合使用）。
func WithStdout() LoggerOption {
	return func(l *defaultLogger) {
		l.stdout = true
	}
}

// WithFileOutput 输出到指定文件路径（追加模式）。
// 可与 WithStdout 组合，同时输出到 stdout 和文件。
func WithFileOutput(path string) LoggerOption {
	return func(l *defaultLogger) {
		l.filePath = path
	}
}

// WithLogLevel 设置 Logger 的初始日志级别。
func WithLogLevel(level Level) LoggerOption {
	return func(l *defaultLogger) {
		l.level = level
	}
}

// defaultLogger 默认的日志实现。
type defaultLogger struct {
	name     string
	level    Level
	hooks    LevelHooks
	fields   map[string]any
	output   io.Writer
	file     *os.File
	filePath string
	stdout   bool
	mu       sync.RWMutex
}

func NewLogger(name string, opts ...LoggerOption) Logger {
	l := &defaultLogger{
		name:   name,
		level:  InfoLevel,
		hooks:  make(LevelHooks),
		fields: make(map[string]any),
		stdout: true,
	}
	for _, opt := range opts {
		opt(l)
	}
	l.output = l.buildWriter()
	return l
}

func (l *defaultLogger) buildWriter() io.Writer {
	var writers []io.Writer

	if l.stdout {
		writers = append(writers, os.Stdout)
	}

	if l.filePath != "" {
		f, err := os.OpenFile(l.filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err == nil {
			l.file = f
			writers = append(writers, f)
		} else {
			// 文件打开失败，回退到 stdout
			_, _ = fmt.Fprintf(os.Stderr, "logx: failed to open log file %s: %v, fallback to stdout\n", l.filePath, err)
			if !l.stdout {
				writers = append(writers, os.Stdout)
			}
		}
	}

	if len(writers) == 0 {
		return os.Stdout
	}
	if len(writers) == 1 {
		return writers[0]
	}
	return io.MultiWriter(writers...)
}

func (l *defaultLogger) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.file != nil {
		return l.file.Close()
	}
	return nil
}

func (l *defaultLogger) Name() string {
	return l.name
}

func (l *defaultLogger) SetLevel(level Level) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.level = level
}

func (l *defaultLogger) AddHook(hook Hook) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.hooks.Add(hook)
}

func (l *defaultLogger) WithField(key string, value any) Logger {
	fields := make(map[string]any, len(l.fields)+1)
	maps.Copy(fields, l.fields)
	fields[key] = value
	return &defaultLogger{
		name:   l.name,
		level:  l.level,
		hooks:  l.hooks,
		fields: fields,
		output: l.output,
		file:   l.file,
		stdout: l.stdout,
	}
}

func (l *defaultLogger) WithFields(fields map[string]any) Logger {
	merged := make(map[string]any, len(l.fields)+len(fields))
	maps.Copy(merged, l.fields)
	maps.Copy(merged, fields)
	return &defaultLogger{
		name:   l.name,
		level:  l.level,
		hooks:  l.hooks,
		fields: merged,
		output: l.output,
		file:   l.file,
		stdout: l.stdout,
	}
}

func (l *defaultLogger) Trace(msg string)                  { l.log(TraceLevel, msg) }
func (l *defaultLogger) Tracef(format string, args ...any) { l.logf(TraceLevel, format, args...) }
func (l *defaultLogger) Debug(msg string)                  { l.log(DebugLevel, msg) }
func (l *defaultLogger) Debugf(format string, args ...any) { l.logf(DebugLevel, format, args...) }
func (l *defaultLogger) Info(msg string)                   { l.log(InfoLevel, msg) }
func (l *defaultLogger) Infof(format string, args ...any)  { l.logf(InfoLevel, format, args...) }
func (l *defaultLogger) Warn(msg string)                   { l.log(WarnLevel, msg) }
func (l *defaultLogger) Warnf(format string, args ...any)  { l.logf(WarnLevel, format, args...) }
func (l *defaultLogger) Error(msg string)                  { l.log(ErrorLevel, msg) }
func (l *defaultLogger) Errorf(format string, args ...any) { l.logf(ErrorLevel, format, args...) }
func (l *defaultLogger) Fatal(msg string)                  { l.log(FatalLevel, msg); os.Exit(1) }
func (l *defaultLogger) Fatalf(format string, args ...any) {
	l.logf(FatalLevel, format, args...)
	os.Exit(1)
}

func (l *defaultLogger) isEnabled(level Level) bool {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return level >= l.level
}

func (l *defaultLogger) log(level Level, msg string) {
	if !l.isEnabled(level) {
		return
	}
	entry := l.newEntry(level, msg)
	l.mu.RLock()
	l.hooks.Fire(entry)
	l.mu.RUnlock()
	l.write(entry)
}

func (l *defaultLogger) logf(level Level, format string, args ...any) {
	if !l.isEnabled(level) {
		return
	}
	entry := l.newEntry(level, fmt.Sprintf(format, args...))
	l.mu.RLock()
	l.hooks.Fire(entry)
	l.mu.RUnlock()
	l.write(entry)
}

func (l *defaultLogger) newEntry(level Level, msg string) *Entry {
	entry := &Entry{
		Level:      level,
		Time:       time.Now(),
		LoggerName: l.name,
		Message:    msg,
	}
	if len(l.fields) > 0 {
		entry.Fields = make(map[string]any, len(l.fields))
		maps.Copy(entry.Fields, l.fields)
	}
	entry.Caller = callerLocation()
	return entry
}

// callerLocation 遍历调用栈，找到第一个不在 logx 包内的调用方。
// 不依赖固定的 skip 层数，避免编译器内联导致定位错误。
func callerLocation() string {
	var pcs [16]uintptr
	// skip 0=Callers, 1=callerLocation, 2=newEntry, 从 3 开始看
	n := runtime.Callers(3, pcs[:])
	frames := runtime.CallersFrames(pcs[:n])
	for {
		frame, more := frames.Next()
		// 跳过 logx 包自身的帧
		if logxFuncPrefix != "" && strings.HasPrefix(frame.Function, logxFuncPrefix) && !strings.HasSuffix(frame.File, "_test.go") {
			if !more {
				break
			}
			continue
		}
		file := frame.File
		if projectRoot != "" {
			file = strings.TrimPrefix(file, projectRoot)
		}
		return fmt.Sprintf("%s:%d", file, frame.Line)
	}
	return "unknown"
}

func (l *defaultLogger) write(entry *Entry) {
	buf := formatEntry(entry)
	l.mu.RLock()
	defer l.mu.RUnlock()
	_, _ = fmt.Fprintln(l.output, buf)
}

func formatEntry(entry *Entry) string {
	ts := entry.Time.Format("2006-01-02 15:04:05.000")
	base := fmt.Sprintf("%s [%-5s] [%s] [%s] %s", ts, entry.Level, entry.LoggerName, entry.Caller, entry.Message)
	if len(entry.Fields) > 0 {
		base += fmt.Sprintf(" %v", entry.Fields)
	}
	return base
}
