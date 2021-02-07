package main

import (
	"fmt"
	"os"
	"runtime"

	"github.com/morvencao/minicni/pkg/args"
	"github.com/morvencao/minicni/pkg/handler"
)

const (
	IPStore = "/tmp/reserved_ips"
)

func init() {
	// this ensures that main runs only on main thread (thread group leader).
	// since namespace ops (unshare, setns) are done for a single thread, we
	// must ensure that the goroutine does not jump from OS thread to thread
	runtime.LockOSThread()
}

func main() {
	cmd, cmdArgs, err := args.GetArgsFromEnv()
	if err != nil {
		fmt.Fprintf(os.Stderr, "getting cmd arguments with error: %v", err)
	}

	fh := handler.NewFileHandler(IPStore)

	switch cmd {
	case "ADD":
		err = fh.HandleAdd(cmdArgs)
	case "DEL":
		err = fh.HandleDel(cmdArgs)
	case "CHECK":
		err = fh.HandleCheck(cmdArgs)
	case "VERSION":
		err = fh.HandleVersion(cmdArgs)
	default:
		err = fmt.Errorf("unknown CNI_COMMAND: %s", cmd)
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to handle CNI_COMMAND %q: %v", cmd, err)
		os.Exit(1)
	}
}
