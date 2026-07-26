// Package domain defines shared business models and stable business error codes.
//
// Error templates are package variables for use with errors.Is. Treat them as
// immutable constants: never reassign them and never store a Cause on a template.
package domain

import (
	"errors"
	"fmt"
)

// Error is a coded business error. Code == 0 is reserved for nil or non-domain
// errors and is not used by success paths, which should return nil.
type Error struct {
	Code    int
	Message string
	Cause   error
}

func (e Error) Error() string {
	if e.Cause == nil {
		return fmt.Sprintf("[%06d] %s", e.Code, e.Message)
	}
	return fmt.Sprintf("[%06d] %s: %v", e.Code, e.Message, e.Cause)
}

func (e Error) Unwrap() error {
	return e.Cause
}

func (e Error) WithCause(err error) Error {
	next := e
	next.Cause = err
	return next
}

func (e Error) Is(target error) bool {
	switch t := target.(type) {
	case Error:
		return e.Code == t.Code
	case *Error:
		return t != nil && e.Code == t.Code
	default:
		return false
	}
}

func CodeOf(err error) int {
	if err == nil {
		return 0
	}
	var target Error
	if errors.As(err, &target) {
		return target.Code
	}
	return 0
}

func PublicMessage(err error) string {
	if err == nil {
		return ""
	}
	var target Error
	if errors.As(err, &target) {
		return fmt.Sprintf("[%06d] %s", target.Code, target.Message)
	}
	return "请求失败，请查看日志"
}
