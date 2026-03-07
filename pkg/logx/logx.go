package logx

import (
	"fmt"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var global *zap.SugaredLogger

func init() {
	cfg := zap.NewDevelopmentConfig()
	cfg.EncoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	cfg.EncoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout("2006-01-02 15:04:05.000")
	logger, _ := cfg.Build(zap.AddCallerSkip(1))
	global = logger.Sugar()
}

func SetGlobal(l *zap.SugaredLogger) {
	global = l
}

func RawLogger() *zap.SugaredLogger {
	return global
}

func Init(level string, filePath string) {
	lvl, err := zapcore.ParseLevel(level)
	if err != nil {
		lvl = zapcore.InfoLevel
	}

	encoderCfg := zap.NewDevelopmentEncoderConfig()
	encoderCfg.EncodeLevel = zapcore.CapitalLevelEncoder
	encoderCfg.EncodeTime = zapcore.TimeEncoderOfLayout("2006-01-02 15:04:05.000")
	encoder := zapcore.NewConsoleEncoder(encoderCfg)

	var cores []zapcore.Core
	// stdout
	cores = append(cores, zapcore.NewCore(encoder, zapcore.AddSync(os.Stdout), lvl))

	// file
	if filePath != "" {
		f, ferr := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if ferr != nil {
			fmt.Fprintf(os.Stderr, "logx: failed to open log file %s: %v\n", filePath, ferr)
		} else {
			cores = append(cores, zapcore.NewCore(encoder, zapcore.AddSync(f), lvl))
		}
	}

	core := zapcore.NewTee(cores...)
	logger := zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1))
	global = logger.Sugar()
}

func Sync() {
	_ = global.Sync()
}

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

func With(args ...any) *zap.SugaredLogger {
	return global.With(args...)
}
