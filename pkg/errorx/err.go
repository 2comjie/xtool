package errorx

import (
	"fmt"
)

type IErr interface {
	Msg() string
	Code() int32
	Err() error
}

type stringErr struct {
	msg  string
	code int32
	err  error
}

func (s stringErr) Msg() string {
	return s.msg
}

func (s stringErr) Code() int32 {
	return s.code
}

func (s stringErr) Err() error {
	return s.err
}

func NewStringErr(err string, code int32) IErr {
	return &stringErr{
		msg:  err,
		code: code,
		err:  fmt.Errorf("msg %s code %d", err, code),
	}
}
func NewFormatErr(format string, vs ...any) IErr {
	errStr := fmt.Sprintf(format, vs)
	return &stringErr{
		msg:  errStr,
		code: 0,
		err:  fmt.Errorf("msg %s code 0", errStr),
	}
}

type ConstStringErr string

func (e ConstStringErr) Error() string { return string(e) }
