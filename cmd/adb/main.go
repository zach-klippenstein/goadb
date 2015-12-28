package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/cheggaaa/pb"
	"github.com/zach-klippenstein/goadb"
	"github.com/zach-klippenstein/goadb/util"
	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	serial = kingpin.Flag("serial", "Connect to device by serial number.").Short('s').String()

	shellCommand    = kingpin.Command("shell", "Run a shell command on the device.")
	shellCommandArg = shellCommand.Arg("command", "Command to run on device.").Strings()

	devicesCommand  = kingpin.Command("devices", "List devices.")
	devicesLongFlag = devicesCommand.Flag("long", "Include extra detail about devices.").Short('l').Bool()

	pullCommand      = kingpin.Command("pull", "Pull a file from the device.")
	pullProgressFlag = pullCommand.Flag("progress", "Show progress.").Short('p').Bool()
	pullRemoteArg    = pullCommand.Arg("remote", "Path of source file on device.").Required().String()
	pullLocalArg     = pullCommand.Arg("local", "Path of destination file.").String()

	pushCommand      = kingpin.Command("push", "Push a file to the device.").Hidden()
	pushProgressFlag = pushCommand.Flag("progress", "Show progress.").Short('p').Bool()
	pushLocalArg     = pushCommand.Arg("local", "Path of source file.").Required().String()
	pushRemoteArg    = pushCommand.Arg("remote", "Path of destination file on device.").Required().String()
)

func main() {
	var exitCode int

	switch kingpin.Parse() {
	case "devices":
		exitCode = listDevices(*devicesLongFlag)
	case "shell":
		exitCode = runShellCommand(*shellCommandArg, parseDevice())
	case "pull":
		exitCode = pull(*pullProgressFlag, *pullRemoteArg, *pullLocalArg, parseDevice())
	case "push":
		exitCode = push(*pushProgressFlag, *pushLocalArg, *pushRemoteArg, parseDevice())
	}

	os.Exit(exitCode)
}

func parseDevice() goadb.DeviceDescriptor {
	if *serial != "" {
		return goadb.DeviceWithSerial(*serial)
	}

	return goadb.AnyDevice()
}

func listDevices(long bool) int {
	client := goadb.NewHostClient(goadb.ClientConfig{})
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

	return 0
}

func runShellCommand(commandAndArgs []string, device goadb.DeviceDescriptor) int {
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

	client := goadb.NewDeviceClient(goadb.ClientConfig{}, device)
	output, err := client.RunCommand(command, args...)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		return 1
	}

	fmt.Print(output)
	return 0
}

func pull(showProgress bool, remotePath, localPath string, device goadb.DeviceDescriptor) int {
	if remotePath == "" {
		fmt.Fprintln(os.Stderr, "error: must specify remote file")
		kingpin.Usage()
		return 1
	}

	if localPath == "" {
		localPath = filepath.Base(remotePath)
	}

	client := goadb.NewDeviceClient(goadb.ClientConfig{}, device)

	info, err := client.Stat(remotePath)
	if util.HasErrCode(err, util.FileNoExistError) {
		fmt.Fprintln(os.Stderr, "remote file does not exist:", remotePath)
		return 1
	} else if err != nil {
		fmt.Fprintf(os.Stderr, "error reading remote file %s: %s\n", remotePath, err)
		return 1
	}

	remoteFile, err := client.OpenRead(remotePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error opening remote file %s: %s\n", remotePath, err)
		return 1
	}
	defer remoteFile.Close()

	localFile, err := os.Create(localPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error opening local file %s: %s\n", localPath, err)
		return 1
	}
	defer localFile.Close()

	var output io.Writer
	var updateProgress func()
	if showProgress {
		output, updateProgress = createProgressBarWriter(int(info.Size), localFile)
	} else {
		output = localFile
		updateProgress = func() {}
	}

	startTime := time.Now()
	copied, err := io.Copy(output, remoteFile)

	// Force progress update if the transfer was really fast.
	updateProgress()

	if err != nil {
		fmt.Fprintln(os.Stderr, "error pulling file:", err)
		return 1
	}

	duration := time.Now().Sub(startTime)
	rate := int64(float64(copied) / duration.Seconds())
	fmt.Printf("%d B/s (%d bytes in %s)\n", rate, copied, duration)
	return 0
}

func push(showProgress bool, localPath, remotePath string, device goadb.DeviceDescriptor) int {
	fmt.Fprintln(os.Stderr, "not implemented")
	return 1
}

func createProgressBarWriter(size int, w io.Writer) (progressWriter io.Writer, update func()) {
	progress := pb.New(size)
	progress.SetUnits(pb.U_BYTES)
	progress.SetRefreshRate(100 * time.Millisecond)
	progress.ShowSpeed = true
	progress.ShowPercent = true
	progress.ShowTimeLeft = true
	progress.Start()

	progressWriter = io.MultiWriter(w, progress)
	update = progress.Update
	return
}
