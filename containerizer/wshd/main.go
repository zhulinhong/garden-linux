package main

import (
	"flag"
	"fmt"
	"os"
	"path"
	"path/filepath"

	"github.com/cloudfoundry-incubator/garden-linux/containerizer"
	"github.com/cloudfoundry-incubator/garden-linux/containerizer/system"
	"github.com/cloudfoundry/gunk/command_runner/linux_command_runner"
)

func missing(flagName string) {
	fmt.Fprintf(os.Stderr, "%s is required\n", flagName)
	flag.Usage()
	os.Exit(1)
}

// TODO: Catch the system errors and panic
func main() {
	runPath := flag.String("run", "./run", "Directory where server socket is placed")
	libPath := flag.String("lib", "./lib", "Directory containing hooks")
	rootFsPath := flag.String("root", "", "Directory that will become root in the new mount namespace")
	userNsFlag := flag.String("userns", "enabled", "If specified, use user namespacing")
	flag.String("title", "", "") // todo: potentially remove this if unused
	flag.Parse()

	if *rootFsPath == "" {
		missing("--root")
	}

	binPath, _ := filepath.Abs(filepath.Dir(os.Args[0]))

	socketPath := path.Join(*runPath, "wshd.sock")

	privileged := false
	if *userNsFlag == "" || *userNsFlag == "disabled" {
		privileged = true
	}

	containerReader, hostWriter, _ := os.Pipe()
	hostReader, containerWriter, _ := os.Pipe()

	sync := &containerizer.PipeSynchronizer{
		Reader: hostReader,
		Writer: hostWriter,
	}

	cz := containerizer.Containerizer{
		InitBinPath: path.Join(binPath, "initd"),
		InitArgs: []string{
			"--socket", socketPath,
			"--root", *rootFsPath,
			"--config", path.Join(*libPath, "../etc/config"),
		},
		Execer: &system.NamespacingExecer{
			CommandRunner: linux_command_runner.New(),
			ExtraFiles:    []*os.File{containerReader, containerWriter},
			Privileged:    privileged,
		},
		Signaller: sync,
		Waiter:    sync,
		// Temporary until we merge the hook scripts functionality in Golang
		CommandRunner: linux_command_runner.New(),
		LibPath:       *libPath,
	}

	err := cz.Create()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create container: %s", err)
		os.Exit(2)
	}
}