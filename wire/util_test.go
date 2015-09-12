package wire

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/zach-klippenstein/goadb/util"
)

func TestAdbServerError_NoRequest(t *testing.T) {
	err := adbServerError("", "fail")
	assert.Equal(t, util.Err{
		Code:    util.AdbError,
		Message: "server error: fail",
		Details: ErrorResponseDetails{
			Request:   "",
			ServerMsg: "fail",
		},
	}, *(err.(*util.Err)))
}

func TestAdbServerError_WithRequest(t *testing.T) {
	err := adbServerError("polite", "fail")
	assert.Equal(t, util.Err{
		Code:    util.AdbError,
		Message: "server error for polite request: fail",
		Details: ErrorResponseDetails{
			Request:   "polite",
			ServerMsg: "fail",
		},
	}, *(err.(*util.Err)))
}

func TestAdbServerError_DeviceNotFound(t *testing.T) {
	err := adbServerError("", "device not found")
	assert.Equal(t, util.Err{
		Code:    util.DeviceNotFound,
		Message: "server error: device not found",
		Details: ErrorResponseDetails{
			Request:   "",
			ServerMsg: "device not found",
		},
	}, *(err.(*util.Err)))
}

func TestAdbServerError_DeviceSerialNotFound(t *testing.T) {
	err := adbServerError("", "device 'LGV4801c74eccd' not found")
	assert.Equal(t, util.Err{
		Code:    util.DeviceNotFound,
		Message: "server error: device 'LGV4801c74eccd' not found",
		Details: ErrorResponseDetails{
			Request:   "",
			ServerMsg: "device 'LGV4801c74eccd' not found",
		},
	}, *(err.(*util.Err)))
}
