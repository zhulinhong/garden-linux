package main

import (
	"encoding/gob"
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"time"

	"github.com/kr/pty"
)

func spawn(socketPath string, path string, argv []string, timeout time.Duration, withTty bool) {
	err := os.MkdirAll(filepath.Dir(socketPath), 0755)
	if err != nil {
		fatal(err)
	}

	listener, err := net.Listen("unix", socketPath)
	if err != nil {
		fatal(err)
	}

	bin, err := exec.LookPath(path)
	if err != nil {
		fatal(err)
	}

	cmd := &exec.Cmd{
		Path: bin,
		Args: argv,
	}

	var stdinW, stdoutR, stderrR *os.File
	var stdinR, stdoutW, stderrW *os.File

	if withTty {
		pty, tty, err := pty.Open()
		if err != nil {
			fatal(err)
		}

		stdinW = pty
		stdoutR = pty
		stderrR = pty

		stdinR = tty
		stdoutW = tty
		stderrW = tty

		cmd.SysProcAttr = &syscall.SysProcAttr{Setctty: true, Setsid: true}
	} else {
		stdinR, stdinW, err = os.Pipe()
		if err != nil {
			fatal(err)
		}

		stdoutR, stdoutW, err = os.Pipe()
		if err != nil {
			fatal(err)
		}

		stderrR, stderrW, err = os.Pipe()
		if err != nil {
			fatal(err)
		}
	}

	cmd.Stdin = stdinR
	cmd.Stdout = stdoutW
	cmd.Stderr = stderrW

	statusR, statusW, err := os.Pipe()
	if err != nil {
		fatal(err)
	}

	fmt.Println("ready")

	started := false

	for {
		conn, err := listener.Accept()
		if err != nil {
			fatal(err)
			break
		}

		rights := syscall.UnixRights(
			int(stdoutR.Fd()),
			int(stderrR.Fd()),
			int(statusR.Fd()),
		)

		_, _, err = conn.(*net.UnixConn).WriteMsgUnix([]byte{}, rights, nil)
		if err != nil {
			log.Println("ERROR WRITING UNIX:", err)
			break
		}

		if !started {
			err := cmd.Start()
			if err != nil {
				fatal(err)
			}

			// close no longer relevant pipe ends
			// this closes tty 3 times but that's OK
			stdinR.Close()
			stdoutW.Close()
			stderrW.Close()

			go func() {
				cmd.Wait()

				if cmd.ProcessState != nil {
					fmt.Fprintf(statusW, "%d\n", cmd.ProcessState.Sys().(syscall.WaitStatus).ExitStatus())
				} else {
					fmt.Fprintf(statusW, "255\n")
				}

				os.Exit(0)
			}()

			fmt.Println("pid:", cmd.Process.Pid)

			// detach from parent process
			os.Stdin.Close()
			os.Stdout.Close()
			os.Stderr.Close()

			started = true
		}

		decoder := gob.NewDecoder(conn)

		for {
			var input Input
			err := decoder.Decode(&input)
			if err != nil {
				break
			}

			if input.EOF {
				err := stdinW.Close()
				if err != nil {
					conn.Close()
					break
				}
			} else {
				_, err := stdinW.Write(input.Data)
				if err != nil {
					conn.Close()
					break
				}
			}
		}
	}
}

func fatal(err error) {
	println("fatal: " + err.Error())
	os.Exit(1)
}
