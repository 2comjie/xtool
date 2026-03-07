package logx

import "sync"

var (
	global    Logger = NewLogger("root")
	loggersMu sync.RWMutex
	loggers   = make(map[string]Logger)
)

func SetGlobal(l Logger) {
	global = l
}

func GetLogger(name string) Logger {
	loggersMu.RLock()
	if l, ok := loggers[name]; ok {
		loggersMu.RUnlock()
		return l
	}
	loggersMu.RUnlock()

	loggersMu.Lock()
	defer loggersMu.Unlock()
	if l, ok := loggers[name]; ok {
		return l
	}
	l := NewLogger(name)
	loggers[name] = l
	return l
}

func SetLevel(level Level) {
	global.SetLevel(level)
}

func AddHook(hook Hook) {
	global.AddHook(hook)
}

func Trace(msg string)                  { global.Trace(msg) }
func Tracef(format string, args ...any) { global.Tracef(format, args...) }
func Debug(msg string)                  { global.Debug(msg) }
func Debugf(format string, args ...any) { global.Debugf(format, args...) }
func Info(msg string)                   { global.Info(msg) }
func Infof(format string, args ...any)  { global.Infof(format, args...) }
func Warn(msg string)                   { global.Warn(msg) }
func Warnf(format string, args ...any)  { global.Warnf(format, args...) }
func Error(msg string)                  { global.Error(msg) }
func Errorf(format string, args ...any) { global.Errorf(format, args...) }
func Fatal(msg string)                  { global.Fatal(msg) }
func Fatalf(format string, args ...any) { global.Fatalf(format, args...) }

func WithField(key string, value any) Logger  { return global.WithField(key, value) }
func WithFields(fields map[string]any) Logger { return global.WithFields(fields) }
