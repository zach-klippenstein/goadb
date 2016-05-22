package errors

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestErrorWithCauseChain(t *testing.T) {
	err := &Err{
		Message: "err1",
		Code:    AssertionError,
		Cause: &Err{
			Message: "err2",
			Code:    AssertionError,
			Cause:   errors.New("err3"),
		},
	}

	expected := `AssertionError: err1
caused by AssertionError: err2
caused by err3`

	assert.Equal(t, expected, ErrorWithCauseChain(err))

	assert.Equal(t, "<err=nil>", ErrorWithCauseChain(nil))
}

func TestCombineErrors(t *testing.T) {
	assert.NoError(t, CombineErrs("hello", AdbError))
	assert.NoError(t, CombineErrs("hello", AdbError, nil, nil))

	err1 := errors.New("lulz")
	err2 := errors.New("fail")

	err := CombineErrs("hello", AdbError, nil, err1, nil)
	assert.EqualError(t, err, "lulz")

	err = CombineErrs("hello", AdbError, err1, err2)
	assert.EqualError(t, err, "AdbError: hello")
	assert.Equal(t, `AdbError: hello
caused by 2 errors: [lulz ∪ fail]`, ErrorWithCauseChain(err))
}
