package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"syscall"
	"time"

	"github.com/cheggaaa/pb"
	"gopkg.in/alecthomas/kingpin.v2"
)

const StdIoFilename = "-"

var (
	serial = kingpin.Flag("serial",
		"Connect to device by serial number.").
		Short('s').
		String()

	shellCommand = kingpin.Command("shell",
		"Run a shell command on the device.")
	shellCommandArg = shellCommand.Arg("command",
		"Command to run on device.").
		Strings()

	devicesCommand = kingpin.Command("devices",
		"List devices.")
	devicesLongFlag = devicesCommand.Flag("long",
		"Include extra detail about devices.").
		Short('l').
		Bool()
	devicesWatchFlag = devicesCommand.Flag("watch",
		"Device watch for the online and offline devices").
		Short('w').
		Bool()

	pullCommand = kingpin.Command("pull",
		"Pull a file from the device.")
	pullProgressFlag = pullCommand.Flag("progress",
		"Show progress.").
		Short('p').
		Bool()
	pullRemoteArg = pullCommand.Arg("remote",
		"Path of source file on device.").
		Required().
		String()
	pullLocalArg = pullCommand.Arg("local",
		"Path of destination file. If -, will write to stdout.").
		String()

	pushCommand = kingpin.Command("push",
		"Push a file to the device.")
	pushProgressFlag = pushCommand.Flag("progress",
		"Show progress.").
		Short('p').
		Bool()
	pushLocalArg = pushCommand.Arg("local",
		"Path of source file. If -, will read from stdin.").
		Required().
		String()
	pushRemoteArg = pushCommand.Arg("remote",
		"Path of destination file on device.").
		Required().
		String()
)

var client *adb.Adb

func main() {
	var exitCode int

	var err error
	client, err = adb.NewWithConfig(adb.ServerConfig{})
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}

	switch kingpin.Parse() {
	case "devices":
		exitCode = listDevices(*devicesLongFlag, *devicesWatchFlag)
	case "shell":
		exitCode = runShellCommand(*shellCommandArg, parseDevice())
	case "pull":
		exitCode = pull(*pullProgressFlag, *pullRemoteArg, *pullLocalArg, parseDevice())
	case "push":
		exitCode = push(*pushProgressFlag, *pushLocalArg, *pushRemoteArg, parseDevice())
	}

	os.Exit(exitCode)
}

func parseDevice() adb.DeviceDescriptor {
	if *serial != "" {
		return adb.DeviceWithSerial(*serial)
	}

	return adb.AnyDevice()
}

func listDevices(long bool, watch bool) int {
	//client := adb.New(server)

	// Listen to the watch flag
	if watch {
		watcher := client.NewDeviceWatcher()
		for event := range watcher.C() {
			fmt.Printf("%+v\n", event)
		}
		if watcher.Err() != nil {
			fmt.Print(watcher.Err().Error())
		}
	} else {
		devices, err := client.ListDevices()
		if err != nil {
			fmt.Fprintln(os.Stderr, "error:", err)
		}
		for _, device := range devices {
			if long {
				if device.Usb == "" {
					fmt.Printf("%s\tproduct:%s model:%s device:%s\n",
						device.Serial, device.Product, device.Model, device.DeviceInfo)
				} else {
					fmt.Printf("%s\tusb:%s product:%s model:%s device:%s\n",
						device.Serial, device.Usb, device.Product, device.Model, device.DeviceInfo)
				}
			} else {
				fmt.Println(device.Serial)
			}
		}
	}

	return 0
}

func runShellCommand(commandAndArgs []string, device adb.DeviceDescriptor) int {
	if len(commandAndArgs) == 0 {
		fmt.Fprintln(os.Stderr, "error: no command")
		kingpin.Usage()
		return 1
	}

	command := commandAndArgs[0]
	var args []string

	if len(commandAndArgs) > 1 {
		args = commandAndArgs[1:]
	}

	client := client.Device(device)
	output, err := client.RunCommand(command, args...)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		return 1
	}

	fmt.Print(output)
	return 0
}

func pull(showProgress bool, remotePath, localPath string, device adb.DeviceDescriptor) int {
	if remotePath == "" {
		fmt.Fprintln(os.Stderr, "error: must specify remote file")
		kingpin.Usage()
		return 1
	}

	if localPath == "" {
		localPath = filepath.Base(remotePath)
	}

	client := client.Device(device)

	info, err := client.Stat(remotePath)
	if adb.HasErrCode(err, adb.ErrCode(adb.FileNoExistError)) {
		fmt.Fprintln(os.Stderr, "remote file does not exist:", remotePath)
		return 1
	} else if err != nil {
		fmt.Fprintf(os.Stderr, "error reading remote file %s: %s\n", remotePath, err)
		return 1
	}

	remoteFile, err := client.OpenRead(remotePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error opening remote file %s: %s\n", remotePath, adb.ErrorWithCauseChain(err))
		return 1
	}
	defer remoteFile.Close()

	var localFile io.WriteCloser
	if localPath == StdIoFilename {
		localFile = os.Stdout
	} else {
		localFile, err = os.Create(localPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error opening local file %s: %s\n", localPath, err)
			return 1
		}
	}
	defer localFile.Close()

	if err := copyWithProgressAndStats(localFile, remoteFile, int(info.Size), showProgress); err != nil {
		fmt.Fprintln(os.Stderr, "error pulling file:", err)
		return 1
	}
	return 0
}

func push(showProgress bool, localPath, remotePath string, device adb.DeviceDescriptor) int {
	if remotePath == "" {
		fmt.Fprintln(os.Stderr, "error: must specify remote file")
		kingpin.Usage()
		return 1
	}

	var (
		localFile io.ReadCloser
		size      int
		perms     os.FileMode
		mtime     time.Time
	)
	if localPath == "" || localPath == StdIoFilename {
		localFile = os.Stdin
		// 0 size will hide the progress bar.
		perms = os.FileMode(0660)
		mtime = adb.MtimeOfClose
	} else {
		var err error
		localFile, err = os.Open(localPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error opening local file %s: %s\n", localPath, err)
			return 1
		}
		info, err := os.Stat(localPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error reading local file %s: %s\n", localPath, err)
			return 1
		}
		size = int(info.Size())
		perms = info.Mode().Perm()
		mtime = info.ModTime()
	}
	defer localFile.Close()

	client := client.Device(device)
	writer, err := client.OpenWrite(remotePath, perms, mtime)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error opening remote file %s: %s\n", remotePath, err)
		return 1
	}
	defer writer.Close()

	if err := copyWithProgressAndStats(writer, localFile, size, showProgress); err != nil {
		fmt.Fprintln(os.Stderr, "error pushing file:", err)
		return 1
	}
	return 0
}

// copyWithProgressAndStats copies src to dst.
// If showProgress is true and size is positive, a progress bar is shown.
// After copying, final stats about the transfer speed and size are shown.
// Progress and stats are printed to stderr.
func copyWithProgressAndStats(dst io.Writer, src io.Reader, size int, showProgress bool) error {
	var progress *pb.ProgressBar
	if showProgress && size > 0 {
		progress = pb.New(size)
		// Write to stderr in case dst is stdout.
		progress.Output = os.Stderr
		progress.ShowSpeed = true
		progress.ShowPercent = true
		progress.ShowTimeLeft = true
		progress.SetUnits(pb.U_BYTES)
		progress.Start()
		dst = io.MultiWriter(dst, progress)
	}

	startTime := time.Now()
	copied, err := io.Copy(dst, src)

	if progress != nil {
		progress.Finish()
	}

	if pathErr, ok := err.(*os.PathError); ok {
		if errno, ok := pathErr.Err.(syscall.Errno); ok && errno == syscall.EPIPE {
			// Pipe closed. Handle this like an EOF.
			err = nil
		}
	}
	if err != nil {
		return err
	}

	duration := time.Now().Sub(startTime)
	rate := int64(float64(copied) / duration.Seconds())
	fmt.Fprintf(os.Stderr, "%d B/s (%d bytes in %s)\n", rate, copied, duration)

	return nil
}
