package errors

import (
	"fmt"
)

type Error struct {
	Code    string
	Message string
	Err     error
}

func (e *Error) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

func Wrap(err error, code, message string) error {
	return &Error{
		Code:    code,
		Message: message,
		Err:     err,
	}
}
