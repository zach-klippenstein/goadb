#goadb

[![Build Status](https://travis-ci.org/zach-klippenstein/goadb.svg?branch=master)](https://travis-ci.org/zach-klippenstein/goadb)
[![GoDoc](https://godoc.org/github.com/zach-klippenstein/goadb?status.svg)](https://godoc.org/github.com/zach-klippenstein/goadb)

A Golang library for interacting with the Android Debug Bridge (adb).

See [demo.go](cmd/demo/demo.go) for usage.

## Usage

go run cmd/adb/main.go

```
usage: main [<flags>] <command> [<args> ...]

Flags:
      --help           Show context-sensitive help (also try --help-long and --help-man).
  -s, --serial=SERIAL  Connect to device by serial number.

Commands:
  help [<command>...]
    Show help.

  shell [<command>...]
    Run a shell command on the device.

  devices [<flags>]
    List devices.

  pull [<flags>] <remote> [<local>]
    Pull a file from the device.

  push [<flags>] <local> <remote>
    Push a file to the device.

```

## Get List of devices

```
go run cmd/adb/main.go devices

Output:
ABCDEFGH

```

## Watch for the device list

```
go run cmd/adb/main.go devices --watch

Output:
{Serial:ABCDEFGH OldState:StateDisconnected NewState:StateOnline}
{Serial:ABCDEFGH OldState:StateOnline NewState:StateOffline}
{Serial:ABCDEFGH OldState:StateOffline NewState:StateDisconnected}
{Serial:ABCDEFGH OldState:StateDisconnected NewState:StateOffline}
{Serial:ABCDEFGH OldState:StateOffline NewState:StateAuthorizing}
{Serial:ABCDEFGH OldState:StateAuthorizing NewState:StateOffline}
{Serial:ABCDEFGH OldState:StateOffline NewState:StateOnline}

```