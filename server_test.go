package goadb

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/zach-klippenstein/goadb/wire"
)

func TestNewServer_ZeroConfig(t *testing.T) {
	config := ServerConfig{}
	fs := &filesystem{
		LookPath: func(name string) (string, error) {
			if name == AdbExecutableName {
				return "/bin/adb", nil
			}
			return "", fmt.Errorf("invalid name: %s", name)
		},
		IsExecutableFile: func(path string) error {
			if path == "/bin/adb" {
				return nil
			}
			return fmt.Errorf("wrong path: %s", path)
		},
	}

	serverIf, err := newServer(config, fs)
	server := serverIf.(*realServer)
	assert.NoError(t, err)
	assert.IsType(t, tcpDialer{}, server.config.Dialer)
	assert.Equal(t, "localhost", server.config.Host)
	assert.Equal(t, AdbPort, server.config.Port)
	assert.Equal(t, fmt.Sprintf("localhost:%d", AdbPort), server.address)
	assert.Equal(t, "/bin/adb", server.config.PathToAdb)
}

type MockDialer struct{}

func (d MockDialer) Dial(address string) (*wire.Conn, error) {
	return nil, nil
}

func TestNewServer_CustomConfig(t *testing.T) {
	config := ServerConfig{
		Dialer:    MockDialer{},
		Host:      "foobar",
		Port:      1,
		PathToAdb: "/bin/adb",
	}
	fs := &filesystem{
		IsExecutableFile: func(path string) error {
			if path == "/bin/adb" {
				return nil
			}
			return fmt.Errorf("wrong path: %s", path)
		},
	}

	serverIf, err := newServer(config, fs)
	server := serverIf.(*realServer)
	assert.NoError(t, err)
	assert.IsType(t, MockDialer{}, server.config.Dialer)
	assert.Equal(t, "foobar", server.config.Host)
	assert.Equal(t, 1, server.config.Port)
	assert.Equal(t, fmt.Sprintf("foobar:1"), server.address)
	assert.Equal(t, "/bin/adb", server.config.PathToAdb)
}

func TestNewServer_AdbNotFound(t *testing.T) {
	config := ServerConfig{}
	fs := &filesystem{
		LookPath: func(name string) (string, error) {
			return "", fmt.Errorf("executable not found: %s", name)
		},
	}

	_, err := newServer(config, fs)
	assert.EqualError(t, err, "ServerNotAvailable: could not find adb in PATH")
}
