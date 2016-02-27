package adb

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/zach-klippenstein/goadb/util"
	"github.com/zach-klippenstein/goadb/wire"
	"golang.org/x/sys/unix"
)

const (
	AdbExecutableName = "adb"

	// Default port the adb server listens on.
	AdbPort = 5037
)

type ServerConfig struct {
	// Path to the adb executable. If empty, the PATH environment variable will be searched.
	PathToAdb string

	// Host and port the adb server is listening on.
	// If not specified, will use the default port on localhost.
	Host string
	Port int

	// Dialer used to connect to the adb server.
	Dialer
}

// Server knows how to start the adb server and connect to it.
type Server interface {
	Start() error
	Dial() (*wire.Conn, error)
}

func roundTripSingleResponse(s Server, req string) ([]byte, error) {
	conn, err := s.Dial()
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	return conn.RoundTripSingleResponse([]byte(req))
}

type realServer struct {
	config ServerConfig
	fs     *filesystem

	// Caches Host:Port so they don't have to be concatenated for every dial.
	address string
}

// NewServer creates a new Server instance.
func NewServer(config ServerConfig) (Server, error) {
	return newServer(config, localFilesystem)
}

func newServer(config ServerConfig, fs *filesystem) (Server, error) {
	if config.Dialer == nil {
		config.Dialer = tcpDialer{}
	}

	if config.Host == "" {
		config.Host = "localhost"
	}
	if config.Port == 0 {
		config.Port = AdbPort
	}

	if config.PathToAdb == "" {
		path, err := fs.LookPath(AdbExecutableName)
		if err != nil {
			return nil, util.WrapErrorf(err, util.ServerNotAvailable, "could not find %s in PATH", AdbExecutableName)
		}
		config.PathToAdb = path
	}
	if err := fs.IsExecutableFile(config.PathToAdb); err != nil {
		return nil, util.WrapErrorf(err, util.ServerNotAvailable, "invalid adb executable: %s", config.PathToAdb)
	}

	return &realServer{
		config:  config,
		fs:      fs,
		address: fmt.Sprintf("%s:%d", config.Host, config.Port),
	}, nil
}

// Dial tries to connect to the server. If the first attempt fails, tries starting the server before
// retrying. If the second attempt fails, returns the error.
func (s *realServer) Dial() (*wire.Conn, error) {
	conn, err := s.config.Dial(s.address)
	if err != nil {
		// Attempt to start the server and try again.
		if err = s.Start(); err != nil {
			return nil, util.WrapErrorf(err, util.ServerNotAvailable, "error starting server for dial")
		}

		conn, err = s.config.Dial(s.address)
		if err != nil {
			return nil, err
		}
	}
	return conn, nil
}

// StartServer ensures there is a server running.
func (s *realServer) Start() error {
	output, err := s.fs.CmdCombinedOutput(s.config.PathToAdb, "start-server")
	outputStr := strings.TrimSpace(string(output))
	return util.WrapErrorf(err, util.ServerNotAvailable, "error starting server: %s\noutput:\n%s", err, outputStr)
}

// filesystem abstracts interactions with the local filesystem for testability.
type filesystem struct {
	// Wraps exec.LookPath.
	LookPath func(string) (string, error)

	// Returns nil if path is a regular file and executable by the current user.
	IsExecutableFile func(path string) error

	// Wraps exec.Command().CombinedOutput()
	CmdCombinedOutput func(name string, arg ...string) ([]byte, error)
}

var localFilesystem = &filesystem{
	LookPath: exec.LookPath,
	IsExecutableFile: func(path string) error {
		info, err := os.Stat(path)
		if err != nil {
			return err
		}
		if !info.Mode().IsRegular() {
			return errors.New("not a regular file")
		}
		return unix.Access(path, unix.X_OK)
	},
	CmdCombinedOutput: func(name string, arg ...string) ([]byte, error) {
		return exec.Command(name, arg...).CombinedOutput()
	},
}
