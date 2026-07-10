package core

import "fmt"

type Error struct {
	Code    string
	Message string
	Data    any
}

func (e Error) Error() string {
	return e.Message
}

func NewError(code string, format string, args ...any) Error {
	return Error{Code: code, Message: fmt.Sprintf(format, args...)}
}
