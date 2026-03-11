package main

import (
	_ "embed"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/getlantern/systray"
	"github.com/sqweek/dialog"
)

//go:generate go run ./scripts/genicon

//go:embed icon.ico
var iconData []byte

var (
	runner       *Runner
	uptimeItem   *systray.MenuItem
	restartsItem *systray.MenuItem
	stopStatus   = make(chan struct{})
)

func main() {
	cfg, err := LoadConfig()
	if err != nil {
		dialog.Message("Failed to load config: %v", err).Title("traycli Error").Error()
		os.Exit(1)
	}
	cc, err := ReadConfig(cfg)
	if err != nil {
		dialog.Message("Failed to read config: %v", err).Title("traycli Error").Error()
		os.Exit(1)
	}
	if cc == nil {
		dialog.Message("config.json not found. Creating template at %s", cfg.ConfigPath).Title("traycli Error").Error()
		if err := WriteEmptyConfig(cfg); err != nil {
			dialog.Message("Failed to create config: %v", err).Title("traycli Error").Error()
			os.Exit(1)
		}
		openFile(cfg.ConfigPath)
		os.Exit(1)
	}
	if len(cc.Cmd) == 0 {
		dialog.Message("config.json has empty cmd. Please edit %s and add a command.", cfg.ConfigPath).Title("traycli Error").Error()
		os.Exit(1)
	}

	runner = NewRunner(cfg, cc)
	systray.Run(onReady, onExit)
}

func openFile(path string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", path)
	case "windows":
		// Use file:// URL for reliable opening with default app
		abs, _ := filepath.Abs(path)
		fileURL := "file:///" + strings.ReplaceAll(filepath.ToSlash(abs), " ", "%20")
		cmd = exec.Command("rundll32.exe", "url.dll,FileProtocolHandler", fileURL)
	default:
		cmd = exec.Command("xdg-open", path)
	}
	cmd.Run()
}

func onReady() {
	systray.SetIcon(iconData)
	systray.SetTitle("CLI")
	systray.SetTooltip("CLI")

	uptimeItem = systray.AddMenuItem("", "")
	uptimeItem.Disable()
	restartsItem = systray.AddMenuItem("", "")
	restartsItem.Disable()
	systray.AddSeparator()
	mConfig := systray.AddMenuItem("config.json", "Edit config")
	mStdout := systray.AddMenuItem("stdout.txt", "Open stdout log")
	mStderr := systray.AddMenuItem("stderr.txt", "Open stderr log")
	systray.AddSeparator()
	mRestart := systray.AddMenuItem("Restart", "Restart the process")
	systray.AddSeparator()
	mQuit := systray.AddMenuItem("Quit", "Quit")

	go runner.Run()
	go func() {
		for range mConfig.ClickedCh {
			openFile(runner.cfg.ConfigPath)
		}
	}()
	go func() {
		for range mStdout.ClickedCh {
			openFile(runner.cfg.StdoutPath)
		}
	}()
	go func() {
		for range mStderr.ClickedCh {
			openFile(runner.cfg.StderrPath)
		}
	}()
	go updateStatus(stopStatus)

	go func() {
		for range mRestart.ClickedCh {
			runner.Restart()
		}
	}()
	go func() {
		<-mQuit.ClickedCh
		systray.Quit()
	}()
}

func onExit() {
	close(stopStatus)
	if runner != nil {
		runner.Stop()
		runner.Wait()
	}
}

func updateStatus(stop <-chan struct{}) {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-stop:
			return
		case <-ticker.C:
			uptime := runner.Uptime()
			restarts := runner.Restarts()
			uptimeItem.SetTitle(formatDuration(uptime))
			restartsItem.SetTitle("Restarts: " + strconv.FormatUint(restarts, 10))
		}
	}
}

func formatDuration(d time.Duration) string {
	if d < time.Minute {
		return "Uptime: " + d.Round(time.Second).String()
	}
	m := int(d.Minutes())
	s := int(d.Seconds()) % 60
	if m >= 60 {
		h := m / 60
		m = m % 60
		return "Uptime: " + strconv.FormatUint(uint64(h), 10) + "h " + strconv.FormatUint(uint64(m), 10) + "m " + strconv.FormatUint(uint64(s), 10) + "s"
	}
	return "Uptime: " + strconv.FormatUint(uint64(m), 10) + "m " + strconv.FormatUint(uint64(s), 10) + "s"
}
