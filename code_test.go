package failure_test

import (
	"errors"
	"testing"

	"io"

	"github.com/morikuni/failure"
)

type CustomCode string

func (c CustomCode) ErrorCode() string {
	return string(c)
}

func TestCode(t *testing.T) {
	const (
		s failure.StringCode = "123"
		c CustomCode         = "123"

		s2 failure.StringCode = "123"
		c2 CustomCode         = "123"
	)

	shouldEqual(t, s.ErrorCode(), "123")
	shouldEqual(t, c.ErrorCode(), "123")

	// 用了反射的DeepEqual， 类型相同的且值相同才是真正的一样
	shouldEqual(t, s, s2)
	shouldEqual(t, c, c2)

	shouldDiffer(t, s, c)
}

func TestIs(t *testing.T) {
	const (
		A failure.StringCode = "A"
		B failure.StringCode = "B"
	)

	// 创建一个错误
	errA := failure.New(A)
	// translate the err to an error with given code
	errB := failure.Translate(errA, B)
	//
	errC := failure.Wrap(errB)

	// 实际上的错误码，但是暴露出来的都是error类型的值
	// failure.Is语义的errA的错误码是A，是的概念有点像是几次的关系
	shouldEqual(t, failure.Is(errA, A), true)
	shouldEqual(t, failure.Is(errB, B), true)
	shouldEqual(t, failure.Is(errC, B), true)

	shouldEqual(t, failure.Is(errA, A, B), true)
	// 既是A，又是B
	shouldEqual(t, failure.Is(errB, A, B), true)
	shouldEqual(t, failure.Is(errC, A, B), true)

	shouldEqual(t, failure.Is(errA, B), false)
	shouldEqual(t, failure.Is(errB, A), false)
	shouldEqual(t, failure.Is(errC, A), false)

	shouldEqual(t, failure.Is(nil, A, B), false)
	shouldEqual(t, failure.Is(io.EOF, A, B), false)
	shouldEqual(t, failure.Is(errA), false)

	shouldEqual(t, failure.Is(nil, nil), true)
	shouldEqual(t, failure.Is(errors.New("error"), nil), true)
}
