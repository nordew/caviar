// Package apperror provides a structured, chainable error type
// with machine-readable codes, human messages, stack traces, and
// HTTP status mapping.
package apperror

import (
	"fmt"
	"io"
	"net/http"
	"runtime"
	"strings"
)

type Code string

const (
	CodeUnknown      Code = "UNKNOWN"
	CodeNotFound     Code = "NOT_FOUND"
	CodeInvalidInput Code = "INVALID_INPUT"
	CodeUnauthorized Code = "UNAUTHORIZED"
	CodeInternal     Code = "INTERNAL_ERROR"
	// extend as needed…
)

type AppError struct {
	Code       Code
	Message    string
	Err        error
	stackTrace string
}

func New(code Code, message string) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		stackTrace: captureStack(),
	}
}

func Wrap(err error, code Code, message string) *AppError {
	if err == nil {
		return nil
	}
	return &AppError{
		Code:       code,
		Message:    message,
		Err:        err,
		stackTrace: captureStack(),
	}
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s: %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func (e *AppError) Unwrap() error {
	return e.Err
}

func (e *AppError) HTTPStatus() int {
	switch e.Code {
	case CodeNotFound:
		return http.StatusNotFound
	case CodeInvalidInput:
		return http.StatusBadRequest
	case CodeUnauthorized:
		return http.StatusUnauthorized
	default:
		return http.StatusInternalServerError
	}
}

func (e *AppError) Format(f fmt.State, c rune) {
	switch c {
	case 'v':
		if f.Flag('+') {
			io.WriteString(f, e.Error())
			io.WriteString(f, "\nStack trace:\n"+e.stackTrace)
			return
		}
		f.Write([]byte(e.Error()))
	default:
		f.Write([]byte(e.Error()))
	}
}

func captureStack() string {
	const depth = 32
	pcs := make([]uintptr, depth)
	n := runtime.Callers(3, pcs)
	frames := runtime.CallersFrames(pcs[:n])

	var sb strings.Builder
	for {
		frame, more := frames.Next()
		sb.WriteString(fmt.Sprintf("%s\n\t%s:%d\n", frame.Function, frame.File, frame.Line))
		if !more {
			break
		}
	}
	return sb.String()
}
