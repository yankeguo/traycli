//go:build windows

package main

import (
	"os/exec"
	"syscall"
)

func setNoWindow(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
}

func gracefulStop(cmd *exec.Cmd, done <-chan struct{}) {
	if cmd == nil || cmd.Process == nil {
		return
	}
	cmd.Process.Kill()
	<-done
}
