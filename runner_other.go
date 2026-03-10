//go:build !windows

package main

import (
	"os/exec"
	"syscall"
	"time"
)

func setNoWindow(cmd *exec.Cmd) {}

func terminateProcess(cmd *exec.Cmd) {
	if cmd == nil || cmd.Process == nil {
		return
	}
	cmd.Process.Signal(syscall.SIGTERM)
	done := make(chan struct{})
	go func() {
		cmd.Process.Wait()
		close(done)
	}()
	select {
	case <-done:
		return
	case <-time.After(5 * time.Second):
		cmd.Process.Kill()
	}
}
