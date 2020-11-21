package adb

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseDeviceState(t *testing.T) {
	for _, test := range []struct {
		String    string
		WantState DeviceState
		WantName  string
		WantError error // Compared by Error() message.
	}{
		{"", StateDisconnected, "StateDisconnected", nil},
		{"offline", StateOffline, "StateOffline", nil},
		{"device", StateOnline, "StateOnline", nil},
		{"unauthorized", StateUnauthorized, "StateUnauthorized", nil},
		{"bad", StateInvalid, "StateInvalid", errors.New(`ParseError: invalid device state: "StateInvalid"`)},
	} {
		state, err := parseDeviceState(test.String)
		if test.WantError == nil {
			assert.NoError(t, err)
		} else {
			assert.EqualError(t, err, test.WantError.Error())
		}
		assert.Equal(t, test.WantState, state)
		assert.Equal(t, test.WantName, state.String())
	}
}
