package logx

import "time"

type Entry struct {
	Level      Level
	Time       time.Time
	LoggerName string
	Message    string
	Fields     map[string]any
	Caller     string
}
