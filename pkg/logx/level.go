package logx

import "strings"

type Level int

const (
	TraceLevel Level = iota
	DebugLevel
	InfoLevel
	WarnLevel
	ErrorLevel
	FatalLevel
	Off
)

var levelNames = [...]string{
	"TRACE",
	"DEBUG",
	"INFO",
	"WARN",
	"ERROR",
	"FATAL",
	"OFF",
}

func (l Level) String() string {
	if l >= TraceLevel && l <= Off {
		return levelNames[l]
	}
	return "UNKNOWN"
}

func ParseLevel(s string) Level {
	switch strings.ToUpper(strings.TrimSpace(s)) {
	case "TRACE":
		return TraceLevel
	case "DEBUG":
		return DebugLevel
	case "INFO":
		return InfoLevel
	case "WARN":
		return WarnLevel
	case "ERROR":
		return ErrorLevel
	case "FATAL":
		return FatalLevel
	case "OFF":
		return Off
	default:
		return InfoLevel
	}
}
