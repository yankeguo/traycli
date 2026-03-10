//go:build !windows

package main

import (
	"os/exec"
	"syscall"
	"time"
)

func setNoWindow(cmd *exec.Cmd) {}

func gracefulStop(cmd *exec.Cmd, done <-chan struct{}) {
	if cmd == nil || cmd.Process == nil {
		return
	}
	cmd.Process.Signal(syscall.SIGTERM)
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		cmd.Process.Kill()
		<-done
	}
}
