package goadb

import (
	"log"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/zach-klippenstein/goadb/util"
	"github.com/zach-klippenstein/goadb/wire"
)

func TestParseDeviceStatesSingle(t *testing.T) {
	states, err := parseDeviceStates(`192.168.56.101:5555	emulator-state
`)

	assert.NoError(t, err)
	assert.Len(t, states, 1)
	assert.Equal(t, "emulator-state", states["192.168.56.101:5555"])
}

func TestParseDeviceStatesMultiple(t *testing.T) {
	states, err := parseDeviceStates(`192.168.56.101:5555	emulator-state
0x0x0x0x	usb-state
`)

	assert.NoError(t, err)
	assert.Len(t, states, 2)
	assert.Equal(t, "emulator-state", states["192.168.56.101:5555"])
	assert.Equal(t, "usb-state", states["0x0x0x0x"])
}

func TestParseDeviceStatesMalformed(t *testing.T) {
	_, err := parseDeviceStates(`192.168.56.101:5555	emulator-state
0x0x0x0x
`)

	assert.True(t, util.HasErrCode(err, util.ParseError))
	assert.Equal(t, "invalid device state line 1: 0x0x0x0x", err.(*util.Err).Message)
}

func TestCalculateStateDiffsUnchangedEmpty(t *testing.T) {
	oldStates := map[string]string{}
	newStates := map[string]string{}

	diffs := calculateStateDiffs(oldStates, newStates)

	assert.Empty(t, diffs)
}

func TestCalculateStateDiffsUnchangedNonEmpty(t *testing.T) {
	oldStates := map[string]string{
		"1": "device",
		"2": "device",
	}
	newStates := map[string]string{
		"1": "device",
		"2": "device",
	}

	diffs := calculateStateDiffs(oldStates, newStates)

	assert.Empty(t, diffs)
}

func TestCalculateStateDiffsOneAdded(t *testing.T) {
	oldStates := map[string]string{}
	newStates := map[string]string{
		"serial": "added",
	}

	diffs := calculateStateDiffs(oldStates, newStates)

	assert.Equal(t, []DeviceStateChangedEvent{
		DeviceStateChangedEvent{"serial", "", "added"},
	}, diffs)
}

func TestCalculateStateDiffsOneRemoved(t *testing.T) {
	oldStates := map[string]string{
		"serial": "removed",
	}
	newStates := map[string]string{}

	diffs := calculateStateDiffs(oldStates, newStates)

	assert.Equal(t, []DeviceStateChangedEvent{
		DeviceStateChangedEvent{"serial", "removed", ""},
	}, diffs)
}

func TestCalculateStateDiffsOneAddedOneUnchanged(t *testing.T) {
	oldStates := map[string]string{
		"1": "device",
	}
	newStates := map[string]string{
		"1": "device",
		"2": "added",
	}

	diffs := calculateStateDiffs(oldStates, newStates)

	assert.Equal(t, []DeviceStateChangedEvent{
		DeviceStateChangedEvent{"2", "", "added"},
	}, diffs)
}

func TestCalculateStateDiffsOneRemovedOneUnchanged(t *testing.T) {
	oldStates := map[string]string{
		"1": "removed",
		"2": "device",
	}
	newStates := map[string]string{
		"2": "device",
	}

	diffs := calculateStateDiffs(oldStates, newStates)

	assert.Equal(t, []DeviceStateChangedEvent{
		DeviceStateChangedEvent{"1", "removed", ""},
	}, diffs)
}

func TestCalculateStateDiffsOneAddedOneRemoved(t *testing.T) {
	oldStates := map[string]string{
		"1": "removed",
	}
	newStates := map[string]string{
		"2": "added",
	}

	diffs := calculateStateDiffs(oldStates, newStates)

	assert.Equal(t, []DeviceStateChangedEvent{
		DeviceStateChangedEvent{"1", "removed", ""},
		DeviceStateChangedEvent{"2", "", "added"},
	}, diffs)
}

func TestCalculateStateDiffsOneChangedOneUnchanged(t *testing.T) {
	oldStates := map[string]string{
		"1": "oldState",
		"2": "device",
	}
	newStates := map[string]string{
		"1": "newState",
		"2": "device",
	}

	diffs := calculateStateDiffs(oldStates, newStates)

	assert.Equal(t, []DeviceStateChangedEvent{
		DeviceStateChangedEvent{"1", "oldState", "newState"},
	}, diffs)
}

func TestCalculateStateDiffsMultipleChangedMultipleUnchanged(t *testing.T) {
	oldStates := map[string]string{
		"1": "oldState",
		"2": "oldState",
	}
	newStates := map[string]string{
		"1": "newState",
		"2": "newState",
	}

	diffs := calculateStateDiffs(oldStates, newStates)

	assert.True(t, reflect.DeepEqual([]DeviceStateChangedEvent{
		DeviceStateChangedEvent{"1", "oldState", "newState"},
		DeviceStateChangedEvent{"2", "oldState", "newState"},
	}, diffs))
}

func TestCalculateStateDiffsOneAddedOneRemovedOneChanged(t *testing.T) {
	oldStates := map[string]string{
		"1": "oldState",
		"2": "removed",
	}
	newStates := map[string]string{
		"1": "newState",
		"3": "added",
	}

	diffs := calculateStateDiffs(oldStates, newStates)

	assert.True(t, reflect.DeepEqual([]DeviceStateChangedEvent{
		DeviceStateChangedEvent{"1", "oldState", "newState"},
		DeviceStateChangedEvent{"2", "removed", ""},
		DeviceStateChangedEvent{"3", "", "added"},
	}, diffs))
}

func TestPublishDevicesRestartsServer(t *testing.T) {
	starter := &MockServerStarter{}
	dialer := &MockServer{
		Status: wire.StatusSuccess,
		Errs: []error{
			nil, nil, nil, // Successful dial.
			util.Errorf(util.ConnectionResetError, "failed first read"),
			util.Errorf(util.ServerNotAvailable, "failed redial"),
		},
	}
	watcher := deviceWatcherImpl{
		config:      ClientConfig{dialer},
		eventChan:   make(chan DeviceStateChangedEvent),
		startServer: starter.StartServer,
	}

	publishDevices(&watcher)

	assert.Empty(t, dialer.Errs)
	assert.Equal(t, []string{"host:track-devices"}, dialer.Requests)
	assert.Equal(t, []string{"Dial", "SendMessage", "ReadStatus", "ReadMessage", "Dial"}, dialer.Trace)
	err := watcher.err.Load().(*util.Err)
	assert.Equal(t, util.ServerNotAvailable, err.Code)
	assert.Equal(t, 1, starter.startCount)
}

type MockServerStarter struct {
	startCount int
	err        error
}

func (s *MockServerStarter) StartServer() error {
	log.Printf("Starting mock server")
	if s.err == nil {
		s.startCount += 1
		return nil
	} else {
		return s.err
	}
}
