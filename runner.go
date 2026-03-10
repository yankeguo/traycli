package main

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"

	"runtime"
)

// Runner executes a command with stdout/stderr redirection and restart on exit.
type Runner struct {
	cfg       *Config
	command   string
	ctx       context.Context
	cancel    context.CancelFunc
	restarts  atomic.Uint64
	startedAt time.Time
	mu        sync.Mutex
	cmd       *exec.Cmd
}

// NewRunner creates a runner for the given command.
func NewRunner(cfg *Config, command string) *Runner {
	ctx, cancel := context.WithCancel(context.Background())
	return &Runner{
		cfg:     cfg,
		command: command,
		ctx:     ctx,
		cancel:  cancel,
	}
}

// Restarts returns the number of restarts so far.
func (r *Runner) Restarts() uint64 {
	return r.restarts.Load()
}

// Uptime returns the duration since the current process started.
func (r *Runner) Uptime() time.Duration {
	r.mu.Lock()
	t := r.startedAt
	r.mu.Unlock()
	if t.IsZero() {
		return 0
	}
	return time.Since(t)
}

// Run executes the command in a loop, restarting 5 seconds after each exit.
func (r *Runner) Run() {
	ensureDir(r.cfg.StdoutPath)
	for {
		select {
		case <-r.ctx.Done():
			r.killProcess()
			return
		default:
		}
		r.runOnce()
		select {
		case <-r.ctx.Done():
			r.killProcess()
			return
		case <-time.After(5 * time.Second):
			r.restarts.Add(1)
		}
	}
}

func (r *Runner) runOnce() {
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/c", r.command)
	} else {
		cmd = exec.Command("sh", "-c", r.command)
	}
	cmd.Env = os.Environ()

	stdoutFile, err := os.OpenFile(r.cfg.StdoutPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return
	}
	defer stdoutFile.Close()
	stderrFile, err := os.OpenFile(r.cfg.StderrPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return
	}
	defer stderrFile.Close()

	cmd.Stdout = stdoutFile
	cmd.Stderr = stderrFile
	cmd.Stdin = nil

	r.mu.Lock()
	r.cmd = cmd
	r.startedAt = time.Now()
	r.mu.Unlock()

	if err := cmd.Start(); err != nil {
		return
	}
	cmd.Wait()
}

func (r *Runner) killProcess() {
	r.mu.Lock()
	cmd := r.cmd
	r.mu.Unlock()
	if cmd != nil && cmd.Process != nil {
		cmd.Process.Kill()
	}
}

// Stop signals the runner to stop.
func (r *Runner) Stop() {
	r.cancel()
}

func ensureDir(filePath string) {
	dir := filepath.Dir(filePath)
	os.MkdirAll(dir, 0755)
}
