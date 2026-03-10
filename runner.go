package main

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"
)

// Runner executes a command with stdout/stderr redirection and restart on exit.
type Runner struct {
	cfg       *Config
	cc        *CommandConfig
	ctx       context.Context
	cancel    context.CancelFunc
	restarts  atomic.Uint64
	startedAt time.Time
	mu        sync.Mutex
	cmd       *exec.Cmd
	wg        sync.WaitGroup
}

// NewRunner creates a runner for the given command config.
func NewRunner(cfg *Config, cc *CommandConfig) *Runner {
	ctx, cancel := context.WithCancel(context.Background())
	r := &Runner{
		cfg:    cfg,
		cc:     cc,
		ctx:    ctx,
		cancel: cancel,
	}
	r.wg.Add(1)
	return r
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
	defer r.wg.Done()
	ensureDir(r.cfg.StdoutPath)
	for {
		select {
		case <-r.ctx.Done():
			return
		default:
		}
		r.runOnce()
		select {
		case <-r.ctx.Done():
			return
		case <-time.After(5 * time.Second):
			r.restarts.Add(1)
		}
	}
}

func (r *Runner) runOnce() {
	cmd := exec.Command(r.cc.Cmd[0], r.cc.Cmd[1:]...)
	cmd.Env = buildEnv(os.Environ(), r.cc.Env)
	setNoWindow(cmd)

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

	if err := cmd.Start(); err != nil {
		return
	}

	r.mu.Lock()
	r.cmd = cmd
	r.startedAt = time.Now()
	r.mu.Unlock()

	done := make(chan struct{})
	go func() {
		cmd.Wait()
		close(done)
	}()

	select {
	case <-r.ctx.Done():
		gracefulStop(cmd, done)
	case <-done:
	}

	r.mu.Lock()
	r.cmd = nil
	r.startedAt = time.Time{}
	r.mu.Unlock()
}

// Stop signals the runner to stop.
func (r *Runner) Stop() {
	r.cancel()
}

// Wait blocks until Run has fully exited.
func (r *Runner) Wait() {
	r.wg.Wait()
}

func ensureDir(filePath string) {
	dir := filepath.Dir(filePath)
	os.MkdirAll(dir, 0755)
}

func buildEnv(base []string, overrides map[string]string) []string {
	if len(overrides) == 0 {
		return base
	}
	overrideKeys := make(map[string]bool)
	for k := range overrides {
		overrideKeys[k] = true
	}
	var result []string
	for _, s := range base {
		key := envKey(s)
		if key != "" && overrideKeys[key] {
			continue
		}
		result = append(result, s)
	}
	for k, v := range overrides {
		result = append(result, k+"="+v)
	}
	return result
}

func envKey(env string) string {
	for i := 0; i < len(env); i++ {
		if env[i] == '=' {
			return env[:i]
		}
	}
	return ""
}
