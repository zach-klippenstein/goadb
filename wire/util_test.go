package wire

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/zach-klippenstein/goadb/internal/errors"
)

func TestAdbServerError_NoRequest(t *testing.T) {
	err := adbServerError("", "fail")
	assert.Equal(t, errors.Err{
		Code:    errors.AdbError,
		Message: "server error: fail",
		Details: ErrorResponseDetails{
			Request:   "",
			ServerMsg: "fail",
		},
	}, *(err.(*errors.Err)))
}

func TestAdbServerError_WithRequest(t *testing.T) {
	err := adbServerError("polite", "fail")
	assert.Equal(t, errors.Err{
		Code:    errors.AdbError,
		Message: "server error for polite request: fail",
		Details: ErrorResponseDetails{
			Request:   "polite",
			ServerMsg: "fail",
		},
	}, *(err.(*errors.Err)))
}

func TestAdbServerError_DeviceNotFound(t *testing.T) {
	err := adbServerError("", "device not found")
	assert.Equal(t, errors.Err{
		Code:    errors.DeviceNotFound,
		Message: "server error: device not found",
		Details: ErrorResponseDetails{
			Request:   "",
			ServerMsg: "device not found",
		},
	}, *(err.(*errors.Err)))
}

func TestAdbServerError_DeviceSerialNotFound(t *testing.T) {
	err := adbServerError("", "device 'LGV4801c74eccd' not found")
	assert.Equal(t, errors.Err{
		Code:    errors.DeviceNotFound,
		Message: "server error: device 'LGV4801c74eccd' not found",
		Details: ErrorResponseDetails{
			Request:   "",
			ServerMsg: "device 'LGV4801c74eccd' not found",
		},
	}, *(err.(*errors.Err)))
}
