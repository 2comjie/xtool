package errorx

import "fmt"

type IErr interface {
	Msg() string
	Code() int32
}

type stringErr struct {
	msg  string
	code int32
}

func (s stringErr) Msg() string {
	return s.msg
}

func (s stringErr) Code() int32 {
	return s.code
}

func NewStringErr(err string, code int32) IErr {
	return &stringErr{
		msg:  err,
		code: code,
	}
}
func NewFormatErr(format string, vs ...any) IErr {
	errStr := fmt.Sprintf(format, vs)
	return &stringErr{
		msg:  errStr,
		code: 0,
	}
}
