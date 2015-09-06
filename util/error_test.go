package util

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
}
