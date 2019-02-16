// Package failure provides an error represented as error code and
// extensible error interface with wrappers.
package failure

import (
	"fmt"
	"strings"
)

// Failure represents an error with error code.
type Failure interface {
	error
	GetCode() Code
}

// CodeOf extracts an error Code from the error.
func CodeOf(err error) (Code, bool) {
	if err == nil {
		return nil, false
	}

	i := NewIterator(err)
	for i.Next() {
		err := i.Error()
		if f, ok := err.(Failure); ok {
			return f.GetCode(), true
		}
	}

	return nil, false
}

// New creates a error from error Code.
func New(code Code, wrappers ...Wrapper) error {
	return Custom(Custom(newFailure(code), wrappers...), WithFormatter(), WithCallStackSkip(1))
	return Custom(Custom(NewFailure(code), wrappers...), WithFormatter(), WithCallStackSkip(1))
}

// Translate translates err to an error with given code.
// It wraps the error with given wrappers, and automatically
// add call stack and formatter.
func Translate(err error, code Code, wrappers ...Wrapper) error {
	return Custom(Custom(Custom(err, WithCode(code)), wrappers...), WithFormatter(), WithCallStackSkip(1))
}

// Wrap wraps err with given wrappers, and automatically add
// call stack and formatter.
func Wrap(err error, wrappers ...Wrapper) error {
	return Custom(Custom(err, wrappers...), WithFormatter(), WithCallStackSkip(1))
}

// Custom is the general error wrapping constructor.
// It just wraps err with given wrappers.
func Custom(err error, wrappers ...Wrapper) error {
	if err == nil {
		return nil
	}
	// To process from left to right, iterate from the last one.
	// Custom(errors.New("foo"), Message("aaa"), Message("bbb")) should be "aaa: bbb: foo".
	for i := len(wrappers) - 1; i >= 0; i-- {
		err = wrappers[i].WrapError(err)
	}
	return err
}

func NewFailure(code Code) Failure {
	return &withCode{code: code}
}

func WithCode(code Code) Wrapper {
	return WrapperFunc(func(err error) error {
		return &withCode{code, err}
	})
}

type withCode struct {
	code       Code
	underlying error
}

var _ interface {
	Failure
	Unwrapper
} = (*withCode)(nil)

func (f *withCode) UnwrapError() error {
	return f.underlying
}

func (f *withCode) GetCode() Code {
	return f.code
}

func (f *withCode) Error() string {
	msg := fmt.Sprintf("code(%s)", f.code.ErrorCode())
	if f.underlying != nil {
		msg = strings.Join([]string{msg, f.underlying.Error()}, ": ")
	}
	return msg
}
